package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"laci-backend/config"
	"laci-backend/models"
)

var driveIDRegex = regexp.MustCompile(`[-\w]{25,}`)

// extractDriveID mengambil ID unik Google Drive dari URL atau string mentah
func extractDriveID(raw string) string {
	raw = strings.TrimSpace(raw)
	match := driveIDRegex.FindString(raw)
	if match != "" {
		return match
	}
	return raw
}

func getAppSetting(key, fallback string) string {
	var setting models.AppSetting
	if err := config.DB.Where("key = ?", key).First(&setting).Error; err == nil {
		if strings.TrimSpace(setting.Value) != "" {
			return setting.Value
		}
	}
	return config.GetEnv(key, fallback)
}

func setAppSetting(key, value string) error {
	setting := models.AppSetting{Key: key, Value: value}
	return config.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value"}),
	}).Create(&setting).Error
}

// GetTemplates menampilkan daftar seluruh template surat aktif
func GetTemplates(c *gin.Context) {
	var templates []models.SuratTemplate
	if err := config.DB.Where("is_active = ?", true).Order("nama_template ASC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar template: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

type SyncTemplatesInput struct {
	FolderURL string `json:"folder_url" binding:"required"`
	GASURL    string `json:"gas_url"`
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

// SyncTemplates memindai folder Google Drive via Google Apps Script (GAS) dan meng-update template di database
func SyncTemplates(c *gin.Context) {
	var input SyncTemplatesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link folder Google Drive wajib diisi"})
		return
	}

	folderID := extractDriveID(input.FolderURL)
	if folderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format link folder Google Drive tidak valid"})
		return
	}

	gasURL := strings.TrimSpace(input.GASURL)
	if gasURL != "" {
		_ = setAppSetting("gas_webhook_url", gasURL)
	} else {
		gasURL = getAppSetting("gas_webhook_url", "")
	}

	if gasURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "URL Webhook Google Apps Script belum dikonfigurasi. Silakan isi URL deployment Web App di form setup.",
		})
		return
	}

	_ = setAppSetting("master_template_folder_url", input.FolderURL)

	// Kirim HTTP POST ke Google Apps Script
	gasReqBody, _ := json.Marshal(map[string]interface{}{
		"action":    "sync_templates",
		"folder_id": folderID,
	})

	httpClient := &http.Client{Timeout: 45 * time.Second}
	resp, err := httpClient.Post(gasURL, "application/json", bytes.NewBuffer(gasReqBody))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Gagal menghubungi Google Apps Script: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	var gasResp GASSyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&gasResp); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": "Respon tidak valid dari Google Apps Script: " + err.Error(),
		})
		return
	}

	if gasResp.Status != "success" {
		msg := gasResp.Message
		if msg == "" {
			msg = "Google Apps Script mengembalikan error"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Simpan atau update ke tabel surat_templates
	var savedTemplates []models.SuratTemplate
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range gasResp.Templates {
			placeholdersJSON, _ := json.Marshal(item.Placeholders)
			tpl := models.SuratTemplate{
				NamaTemplate: item.Nama,
				GDocID:       item.GDocID,
				Placeholders: string(placeholdersJSON),
				IsActive:     true,
			}

			var existing models.SuratTemplate
			tx.Where("g_doc_id = ?", item.GDocID).Limit(1).Find(&existing)
			if existing.ID != uuid.Nil {
				existing.NamaTemplate = item.Nama
				existing.Placeholders = string(placeholdersJSON)
				existing.IsActive = true
				if err := tx.Save(&existing).Error; err != nil {
					return err
				}
				savedTemplates = append(savedTemplates, existing)
			} else {
				if err := tx.Create(&tpl).Error; err != nil {
					return err
				}
				savedTemplates = append(savedTemplates, tpl)
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan template ke database: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Berhasil menyinkronkan %d template surat", len(savedTemplates)),
		"templates": savedTemplates,
		"count":     len(savedTemplates),
	})
}

type GenerateSuratInput struct {
	TemplateID   string            `json:"template_id" binding:"required"`
	FolderURL    string            `json:"folder_url"`
	FileName     string            `json:"file_name"`
	NomorSurat   string            `json:"nomor_surat" binding:"required"`
	Perihal      string            `json:"perihal" binding:"required"`
	Tujuan       string            `json:"tujuan"`
	PICNama      string            `json:"pic_nama"`
	ProkerID     *string           `json:"proker_id"`
	Replacements map[string]string `json:"replacements"`
}

type GASGenerateResponse struct {
	Status   string `json:"status"`
	DocID    string `json:"doc_id"`
	GDocsURL string `json:"gdocs_url"`
	PDFURL   string `json:"pdf_url"`
	Message  string `json:"message"`
}

// GenerateSurat memanggil GAS untuk menyalin template, mereplace data, dan menyimpan record ke DB Laci
func GenerateSurat(c *gin.Context) {
	var input GenerateSuratInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data input tidak lengkap: " + err.Error()})
		return
	}

	tplUUID, err := uuid.Parse(input.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Template ID tidak valid"})
		return
	}

	var template models.SuratTemplate
	if err := config.DB.First(&template, "id = ?", tplUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template tidak ditemukan di sistem"})
		return
	}

	gasURL := getAppSetting("gas_webhook_url", "")
	if gasURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "URL Webhook Google Apps Script belum diset"})
		return
	}

	folderURL := strings.TrimSpace(input.FolderURL)
	if folderURL == "" {
		folderURL = getAppSetting("default_output_folder_url", "")
	}
	folderID := extractDriveID(folderURL)

	fileName := strings.TrimSpace(input.FileName)
	if fileName == "" {
		dateStr := time.Now().Format("02Jan2006")
		cleanTemplate := strings.ReplaceAll(template.NamaTemplate, "/", "_")

		if input.ProkerID != nil && strings.TrimSpace(*input.ProkerID) != "" {
			var proker models.Proker
			if err := config.DB.First(&proker, "id = ?", input.ProkerID).Error; err == nil {
				cleanProker := strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(proker.NamaProker)), " ", "-")
				fileName = fmt.Sprintf("%s_%s_%s", cleanTemplate, cleanProker, dateStr)
			}
		}
		if fileName == "" {
			fileName = fmt.Sprintf("%s_%s", cleanTemplate, dateStr)
		}
	}
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")

	// Otomatis masukkan nomor_surat, perihal, dan tujuan ke replacements jika belum ada
	if input.Replacements == nil {
		input.Replacements = make(map[string]string)
	}
	if _, ok := input.Replacements["nomor_surat"]; !ok {
		input.Replacements["nomor_surat"] = input.NomorSurat
	}
	if _, ok := input.Replacements["perihal"]; !ok {
		input.Replacements["perihal"] = input.Perihal
	}
	if _, ok := input.Replacements["tujuan"]; !ok {
		input.Replacements["tujuan"] = input.Tujuan
	}

	gasPayload, _ := json.Marshal(map[string]interface{}{
		"action":       "generate_surat",
		"template_id":  template.GDocID,
		"folder_id":    folderID,
		"file_name":    fileName,
		"replacements": input.Replacements,
	})

	httpClient := &http.Client{Timeout: 45 * time.Second}
	resp, err := httpClient.Post(gasURL, "application/json", bytes.NewBuffer(gasPayload))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Gagal menghubungi Google Apps Script: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	var gasResp GASGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&gasResp); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Respon tidak valid dari Google Apps Script: " + err.Error()})
		return
	}

	if gasResp.Status != "success" {
		msg := gasResp.Message
		if msg == "" {
			msg = "Gagal memproses dokumen di Google Drive"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Tentukan PICNama (default '-' jika belum ada pengurus yang bertugas membawa surat)
	picNama := strings.TrimSpace(input.PICNama)
	if picNama == "" {
		picNama = "-"
	}

	tujuanSurat := strings.TrimSpace(input.Tujuan)
	if tujuanSurat == "" {
		tujuanSurat = template.NamaTemplate
	}

	var prokerUUID *uuid.UUID
	if input.ProkerID != nil && strings.TrimSpace(*input.ProkerID) != "" {
		if parsed, err := uuid.Parse(*input.ProkerID); err == nil {
			prokerUUID = &parsed
		}
	}

	suratID := uuid.New()
	surat := models.Surat{
		ID:            suratID,
		NomorSurat:    input.NomorSurat,
		Perihal:       input.Perihal,
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

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&surat).Error; err != nil {
			return err
		}
		if err := tx.Create(&logTracking).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan surat ke database: " + err.Error()})
		return
	}

	surat.Logs = []models.LogTracking{logTracking}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Surat berhasil digenerate dan disimpan ke sistem",
		"surat":     surat,
		"gdocs_url": gasResp.GDocsURL,
		"pdf_url":   gasResp.PDFURL,
		"doc_id":    gasResp.DocID,
	})
}

// GetAppSettings mengambil konfigurasi sistem terkait Google Drive / GAS
func GetAppSettings(c *gin.Context) {
	var settings []models.AppSetting
	config.DB.Find(&settings)

	res := make(map[string]string)
	for _, s := range settings {
		res[s.Key] = s.Value
	}
	c.JSON(http.StatusOK, res)
}

// UpdateAppSettings memperbarui konfigurasi sistem
func UpdateAppSettings(c *gin.Context) {
	var body map[string]string
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload tidak valid"})
		return
	}

	for k, v := range body {
		_ = setAppSetting(k, strings.TrimSpace(v))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengaturan berhasil diperbarui"})
}
