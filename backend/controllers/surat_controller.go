package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"laci-backend/config"
	"laci-backend/models"
)

type CreateSuratInput struct {
	NomorSurat    string `json:"nomor_surat" binding:"required"`
	Perihal       string `json:"perihal" binding:"required"`
	StatusSaatIni string `json:"status_saat_ini"`
	PICNama       string `json:"pic_nama" binding:"required"`
	Catatan       string `json:"catatan"`
}

type UpdateSuratInput struct {
	StatusSaatIni string  `json:"status_saat_ini"`
	StatusBaru    string  `json:"status_baru"`
	Status        string  `json:"status"`
	Catatan       *string `json:"catatan"`
	CatatanLog    *string `json:"catatan_log"`
}

func GetSurat(c *gin.Context) {
	var listSurat []models.Surat
	if err := config.DB.Preload("Logs").Find(&listSurat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listSurat)
}

func CreateSurat(c *gin.Context) {
	var input CreateSuratInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)

	status := input.StatusSaatIni
	if role == "humas" || strings.TrimSpace(status) == "" {
		status = "Standby"
	}

	suratID := uuid.New()
	surat := models.Surat{
		ID:            suratID,
		NomorSurat:    input.NomorSurat,
		Perihal:       input.Perihal,
		StatusSaatIni: status,
		PICNama:       input.PICNama,
		Catatan:       input.Catatan,
	}

	logTracking := models.LogTracking{
		ID:          uuid.New(),
		SuratID:     suratID,
		StatusBaru:  status,
		CatatanLog:  input.Catatan,
		WaktuUpdate: time.Now(),
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&surat).Error; err != nil {
			return err
		}
		if err := tx.Create(&logTracking).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	surat.Logs = []models.LogTracking{logTracking}
	c.JSON(http.StatusCreated, surat)
}

func UpdateSurat(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var surat models.Surat
	if err := config.DB.First(&surat, "id = ?", suratUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
		return
	}

	var input UpdateSuratInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newStatus := input.StatusSaatIni
	if newStatus == "" {
		newStatus = input.StatusBaru
	}
	if newStatus == "" {
		newStatus = input.Status
	}

	// If no status is supplied, retain the current status
	if newStatus == "" {
		newStatus = surat.StatusSaatIni
	}

	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)

	// Humas is permitted to edit notes, while status transitions are preserved
	if role == "humas" && newStatus != surat.StatusSaatIni {
		newStatus = surat.StatusSaatIni
	}

	// Determine log note and whether to update surat note
	note := surat.Catatan
	if input.Catatan != nil {
		note = *input.Catatan
	}
	catatanLog := note
	if input.CatatanLog != nil {
		catatanLog = *input.CatatanLog
	}

	logTracking := models.LogTracking{
		ID:          uuid.New(),
		SuratID:     surat.ID,
		StatusBaru:  newStatus,
		CatatanLog:  catatanLog,
		WaktuUpdate: time.Now(),
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"status_saat_ini": newStatus,
		}
		if input.Catatan != nil {
			updates["catatan"] = *input.Catatan
		}
		if err := tx.Model(&surat).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Create(&logTracking).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	surat.StatusSaatIni = newStatus
	if input.Catatan != nil {
		surat.Catatan = *input.Catatan
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Surat updated successfully",
		"surat":   surat,
		"log":     logTracking,
	})
}

func UploadSurat(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var surat models.Surat
	if err := config.DB.First(&surat, "id = ?", suratUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Key 'file' is required."})
		return
	}

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s_%s%s", suratUUID.String(), uuid.New().String()[:8], ext)
	dst := filepath.Join("uploads", newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	fileArsipURL := "/uploads/" + newFileName
	updates := map[string]interface{}{
		"file_arsip": fileArsipURL,
		"arsip_url":  fileArsipURL,
	}
	if err := config.DB.Model(&surat).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	surat.FileArsip = fileArsipURL
	surat.ArsipURL = &fileArsipURL
	c.JSON(http.StatusOK, gin.H{
		"message":    "File arsip berhasil diunggah",
		"file_arsip": fileArsipURL,
		"arsip_url":  fileArsipURL,
		"surat":      surat,
	})
}

func UploadArsip(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var surat models.Surat
	if err := config.DB.First(&surat, "id = ?", suratUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("arsip")
	}
	if err != nil {
		file, err = c.FormFile("image")
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Key 'file', 'arsip', or 'image' is required."})
		return
	}

	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s_%s%s", suratUUID.String(), uuid.New().String()[:8], ext)
	dst := filepath.Join("uploads", newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	arsipURL := "/uploads/" + newFileName
	updates := map[string]interface{}{
		"file_arsip": arsipURL,
		"arsip_url":  arsipURL,
	}
	if err := config.DB.Model(&surat).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	surat.FileArsip = arsipURL
	surat.ArsipURL = &arsipURL
	c.JSON(http.StatusOK, gin.H{
		"message":    "Arsip uploaded successfully",
		"file_arsip": arsipURL,
		"arsip_url":  arsipURL,
		"surat":      surat,
	})
}

func DeleteSurat(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		var surat models.Surat
		if err := tx.First(&surat, "id = ?", suratUUID).Error; err != nil {
			return err
		}

		if err := tx.Where("surat_id = ?", suratUUID).Delete(&models.LogTracking{}).Error; err != nil {
			return err
		}

		if err := tx.Delete(&surat).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Surat dan log terkait berhasil dihapus"})
}
