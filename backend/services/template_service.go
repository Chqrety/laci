package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"laci-backend/models"
	"laci-backend/repositories"
)

var driveIDRegex = regexp.MustCompile(`[-\w]{25,}`)

func ExtractDriveID(raw string) string {
	raw = strings.TrimSpace(raw)
	match := driveIDRegex.FindString(raw)
	if match != "" {
		return match
	}
	return raw
}

type GASTemplateItem struct {
	GDocID       string   `json:"gdoc_id"`
	Nama         string   `json:"nama"`
	Placeholders []string `json:"placeholders"`
}

type GASSyncResponse struct {
	Status    string            `json:"status"`
	Templates []GASTemplateItem `json:"templates"`
	Count     int               `json:"count"`
	Message   string            `json:"message"`
}

type GASGenerateResponse struct {
	Status   string `json:"status"`
	DocID    string `json:"doc_id"`
	GDocsURL string `json:"gdocs_url"`
	PDFURL   string `json:"pdf_url"`
	Message  string `json:"message"`
}

type GenerateSuratParams struct {
	TemplateID   string
	FolderURL    string
	FileName     string
	NomorSurat   string
	Perihal      string
	Tujuan       string
	PICNama      string
	ProkerID     *string
	Replacements map[string]string
}

type TemplateService struct {
	templateRepo *repositories.TemplateRepository
	suratRepo    *repositories.SuratRepository
	prokerRepo   *repositories.ProkerRepository
}

func NewTemplateService(
	templateRepo *repositories.TemplateRepository,
	suratRepo *repositories.SuratRepository,
	prokerRepo *repositories.ProkerRepository,
) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		suratRepo:    suratRepo,
		prokerRepo:   prokerRepo,
	}
}

func (s *TemplateService) GetActiveTemplates() ([]models.SuratTemplate, error) {
	return s.templateRepo.FindActiveTemplates()
}

func (s *TemplateService) SyncTemplates(folderURL, gasURL string) ([]models.SuratTemplate, error) {
	folderID := ExtractDriveID(folderURL)
	if folderID == "" {
		return nil, errors.New("Format link folder Google Drive tidak valid")
	}

	cleanGASURL := strings.TrimSpace(gasURL)
	if cleanGASURL != "" {
		_ = s.templateRepo.SetSetting("gas_webhook_url", cleanGASURL)
	} else {
		cleanGASURL = s.templateRepo.GetSetting("gas_webhook_url", "")
	}

	if cleanGASURL == "" {
		return nil, errors.New("URL Webhook Google Apps Script belum dikonfigurasi. Silakan isi URL deployment Web App di form setup.")
	}

	_ = s.templateRepo.SetSetting("master_template_folder_url", folderURL)

	gasReqBody, _ := json.Marshal(map[string]interface{}{
		"action":    "sync_templates",
		"folder_id": folderID,
	})

	httpClient := &http.Client{Timeout: 45 * time.Second}
	resp, err := httpClient.Post(cleanGASURL, "application/json", bytes.NewBuffer(gasReqBody))
	if err != nil {
		return nil, fmt.Errorf("Gagal menghubungi Google Apps Script: %w", err)
	}
	defer resp.Body.Close()

	var gasResp GASSyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&gasResp); err != nil {
		return nil, fmt.Errorf("Respon tidak valid dari Google Apps Script: %w", err)
	}

	if gasResp.Status != "success" {
		msg := gasResp.Message
		if msg == "" {
			msg = "Google Apps Script mengembalikan error"
		}
		return nil, errors.New(msg)
	}

	var toUpsert []models.SuratTemplate
	for _, item := range gasResp.Templates {
		placeholdersJSON, _ := json.Marshal(item.Placeholders)
		toUpsert = append(toUpsert, models.SuratTemplate{
			NamaTemplate: item.Nama,
			GDocID:       item.GDocID,
			Placeholders: string(placeholdersJSON),
			IsActive:     true,
		})
	}

	return s.templateRepo.UpsertTemplates(toUpsert)
}

func (s *TemplateService) GenerateSurat(p GenerateSuratParams) (*models.Surat, *GASGenerateResponse, error) {
	tplUUID, err := uuid.Parse(p.TemplateID)
	if err != nil {
		return nil, nil, errors.New("Template ID tidak valid")
	}

	template, err := s.templateRepo.FindTemplateByID(tplUUID)
	if err != nil {
		return nil, nil, errors.New("Template tidak ditemukan di sistem")
	}

	gasURL := s.templateRepo.GetSetting("gas_webhook_url", "")
	if gasURL == "" {
		return nil, nil, errors.New("URL Webhook Google Apps Script belum diset")
	}

	folderURL := strings.TrimSpace(p.FolderURL)
	if folderURL == "" {
		folderURL = s.templateRepo.GetSetting("default_output_folder_url", "")
	}
	folderID := ExtractDriveID(folderURL)

	fileName := strings.TrimSpace(p.FileName)
	if fileName == "" {
		dateStr := time.Now().Format("02Jan2006")
		cleanTemplate := strings.ReplaceAll(template.NamaTemplate, "/", "_")

		if p.ProkerID != nil && strings.TrimSpace(*p.ProkerID) != "" {
			if prokerUUID, err := uuid.Parse(*p.ProkerID); err == nil {
				if proker, err := s.prokerRepo.FindByID(prokerUUID); err == nil {
					cleanProker := strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(proker.NamaProker)), " ", "-")
					fileName = fmt.Sprintf("%s_%s_%s", cleanTemplate, cleanProker, dateStr)
				}
			}
		}
		if fileName == "" {
			fileName = fmt.Sprintf("%s_%s", cleanTemplate, dateStr)
		}
	}
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")

	if p.Replacements == nil {
		p.Replacements = make(map[string]string)
	}
	if _, ok := p.Replacements["nomor_surat"]; !ok {
		p.Replacements["nomor_surat"] = p.NomorSurat
	}
	if _, ok := p.Replacements["perihal"]; !ok {
		p.Replacements["perihal"] = p.Perihal
	}
	if _, ok := p.Replacements["tujuan"]; !ok {
		p.Replacements["tujuan"] = p.Tujuan
	}

	gasPayload, _ := json.Marshal(map[string]interface{}{
		"action":       "generate_surat",
		"template_id":  template.GDocID,
		"folder_id":    folderID,
		"file_name":    fileName,
		"replacements": p.Replacements,
	})

	httpClient := &http.Client{Timeout: 45 * time.Second}
	resp, err := httpClient.Post(gasURL, "application/json", bytes.NewBuffer(gasPayload))
	if err != nil {
		return nil, nil, fmt.Errorf("Gagal menghubungi Google Apps Script: %w", err)
	}
	defer resp.Body.Close()

	var gasResp GASGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gasResp); err != nil {
		return nil, nil, fmt.Errorf("Respon tidak valid dari Google Apps Script: %w", err)
	}

	if gasResp.Status != "success" {
		msg := gasResp.Message
		if msg == "" {
			msg = "Gagal memproses dokumen di Google Drive"
		}
		return nil, nil, errors.New(msg)
	}

	picNama := strings.TrimSpace(p.PICNama)
	if picNama == "" {
		picNama = "-"
	}

	tujuanSurat := strings.TrimSpace(p.Tujuan)
	if tujuanSurat == "" {
		tujuanSurat = template.NamaTemplate
	}

	var prokerUUID *uuid.UUID
	if p.ProkerID != nil && strings.TrimSpace(*p.ProkerID) != "" {
		if parsed, err := uuid.Parse(*p.ProkerID); err == nil {
			prokerUUID = &parsed
		}
	}

	suratID := uuid.New()
	surat := models.Surat{
		ID:            suratID,
		NomorSurat:    p.NomorSurat,
		Perihal:       p.Perihal,
		Tujuan:        tujuanSurat,
		StatusSaatIni: "Standby",
		PICNama:       picNama,
		ArsipURL:      &gasResp.PDFURL,
		GDocsURL:      &gasResp.GDocsURL,
		Catatan:       fmt.Sprintf("Dibuat otomatis dari template: %s", template.NamaTemplate),
		ProkerID:      prokerUUID,
	}

	logTracking := models.LogTracking{
		ID:          uuid.New(),
		SuratID:     suratID,
		StatusBaru:  "Standby",
		CatatanLog:  fmt.Sprintf("Surat dibuat otomatis dari template '%s'", template.NamaTemplate),
		WaktuUpdate: time.Now(),
	}

	if err := s.suratRepo.CreateWithLog(&surat, &logTracking); err != nil {
		return nil, nil, err
	}

	surat.Logs = []models.LogTracking{logTracking}
	return &surat, &gasResp, nil
}

func (s *TemplateService) GetAppSettings() (map[string]string, error) {
	return s.templateRepo.GetAllSettings()
}

func (s *TemplateService) UpdateAppSettings(settings map[string]string) error {
	for k, v := range settings {
		if err := s.templateRepo.SetSetting(k, v); err != nil {
			return err
		}
	}
	return nil
}
