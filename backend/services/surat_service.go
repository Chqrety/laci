package services

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"laci-backend/models"
	"laci-backend/repositories"
)

var (
	ErrSuratNotFound   = errors.New("Surat not found")
	ErrArsipRestricted = errors.New("Unggah bukti arsip fisik hanya dapat dilakukan saat surat telah selesai ditandatangani dan berstatus Selesai.")
)

type SuratService struct {
	suratRepo *repositories.SuratRepository
}

func NewSuratService(suratRepo *repositories.SuratRepository) *SuratService {
	return &SuratService{suratRepo: suratRepo}
}

func (s *SuratService) GetAllSurat() ([]models.Surat, error) {
	return s.suratRepo.FindAllWithLogs()
}

func (s *SuratService) CreateSurat(callerDivisi, nomorSurat, perihal, tujuan, status, picNama, catatan string) (*models.Surat, error) {
	divisi := strings.ToLower(strings.TrimSpace(callerDivisi))
	if (divisi != "sekre" && divisi != "ketum") || strings.TrimSpace(status) == "" {
		status = "Standby"
	}

	suratID := uuid.New()
	surat := models.Surat{
		ID:            suratID,
		NomorSurat:    nomorSurat,
		Perihal:       perihal,
		Tujuan:        tujuan,
		StatusSaatIni: status,
		PICNama:       picNama,
		Catatan:       catatan,
	}

	logTracking := models.LogTracking{
		ID:          uuid.New(),
		SuratID:     suratID,
		StatusBaru:  status,
		CatatanLog:  catatan,
		WaktuUpdate: time.Now(),
	}

	if err := s.suratRepo.CreateWithLog(&surat, &logTracking); err != nil {
		return nil, err
	}

	surat.Logs = []models.LogTracking{logTracking}
	return &surat, nil
}

func (s *SuratService) UpdateSurat(suratUUID uuid.UUID, statusSaatIni string, tujuan, catatan, catatanLog *string) (*models.Surat, *models.LogTracking, error) {
	surat, err := s.suratRepo.FindByID(suratUUID)
	if err != nil {
		return nil, nil, ErrSuratNotFound
	}

	newStatus := strings.TrimSpace(statusSaatIni)
	if newStatus == "" {
		newStatus = surat.StatusSaatIni
	}

	note := surat.Catatan
	if catatan != nil {
		note = *catatan
	}
	logNote := note
	if catatanLog != nil {
		logNote = *catatanLog
	}

	logTracking := models.LogTracking{
		ID:          uuid.New(),
		SuratID:     surat.ID,
		StatusBaru:  newStatus,
		CatatanLog:  logNote,
		WaktuUpdate: time.Now(),
	}

	updates := map[string]interface{}{
		"status_saat_ini": newStatus,
	}
	if tujuan != nil {
		updates["tujuan"] = *tujuan
	}
	if catatan != nil {
		updates["catatan"] = *catatan
	}

	if err := s.suratRepo.UpdateWithLog(surat, updates, &logTracking); err != nil {
		return nil, nil, err
	}

	surat.StatusSaatIni = newStatus
	if catatan != nil {
		surat.Catatan = *catatan
	}

	return surat, &logTracking, nil
}

func (s *SuratService) SaveUploadedArsip(suratUUID uuid.UUID, callerDivisi, originalFilename string, fileReader io.Reader) (string, *models.Surat, error) {
	surat, err := s.suratRepo.FindByID(suratUUID)
	if err != nil {
		return "", nil, ErrSuratNotFound
	}

	divisi := strings.ToLower(callerDivisi)
	if divisi == "humas" || divisi == "pengurus" || divisi == "ex-ketum" {
		statusLower := strings.ToLower(surat.StatusSaatIni)
		if !strings.Contains(statusLower, "selesai") && !strings.Contains(statusLower, "arsip") {
			return "", nil, ErrArsipRestricted
		}
	}

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		return "", nil, errors.New("Gagal membuat direktori uploads")
	}

	ext := filepath.Ext(originalFilename)
	newFileName := fmt.Sprintf("%s_%s%s", suratUUID.String(), uuid.New().String()[:8], ext)
	dst := filepath.Join("uploads", newFileName)

	out, err := os.Create(dst)
	if err != nil {
		return "", nil, errors.New("Gagal menyimpan berkas arsip")
	}
	defer out.Close()

	if _, err := io.Copy(out, fileReader); err != nil {
		return "", nil, errors.New("Gagal menyalin isi berkas arsip")
	}

	fileArsipURL := "/uploads/" + newFileName
	if err := s.suratRepo.UpdateArsipURL(suratUUID, fileArsipURL); err != nil {
		return "", nil, err
	}

	surat.ArsipURL = &fileArsipURL
	return fileArsipURL, surat, nil
}

func (s *SuratService) DeleteSurat(suratUUID uuid.UUID) error {
	return s.suratRepo.DeleteWithLogs(suratUUID)
}
