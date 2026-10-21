package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"laci-backend/config"
	"laci-backend/models"
)

type CreateProkerInput struct {
	NamaProker string `json:"nama_proker" binding:"required"`
	Deskripsi  string `json:"deskripsi"`
}

func GetProkers(c *gin.Context) {
	var listProker []models.Proker
	if err := config.DB.Order("nama_proker asc").Find(&listProker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listProker)
}

func CreateProker(c *gin.Context) {
	var input CreateProkerInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker is required"})
		return
	}

	trimmed := strings.TrimSpace(input.NamaProker)
	if trimmed == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker cannot be empty"})
		return
	}

	proker := models.Proker{
		ID:         uuid.New(),
		NamaProker: trimmed,
		Deskripsi:  strings.TrimSpace(input.Deskripsi),
	}

	if err := config.DB.Create(&proker).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Proker already exists or failed to save"})
		return
	}

	c.JSON(http.StatusCreated, proker)
}

func UpdateProker(c *gin.Context) {
	idParam := c.Param("id")
	prokerUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var proker models.Proker
	if err := config.DB.First(&proker, "id = ?", prokerUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proker not found"})
		return
	}

	var input CreateProkerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker is required"})
		return
	}

	trimmed := strings.TrimSpace(input.NamaProker)
	if trimmed == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker cannot be empty"})
		return
	}

	var count int64
	config.DB.Model(&models.Proker{}).Where("nama_proker = ? AND id != ?", trimmed, prokerUUID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama proker sudah digunakan"})
		return
	}

	proker.NamaProker = trimmed
	proker.Deskripsi = strings.TrimSpace(input.Deskripsi)

	if err := config.DB.Save(&proker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proker)
}

func DeleteProker(c *gin.Context) {
	idParam := c.Param("id")
	prokerUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var proker models.Proker
	if err := config.DB.First(&proker, "id = ?", prokerUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Proker not found"})
		return
	}

	// Detach relasi proker_id dari surat jika ada
	config.DB.Model(&models.Surat{}).Where("proker_id = ?", prokerUUID).Update("proker_id", nil)

	if err := config.DB.Delete(&proker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Program kerja berhasil dihapus"})
}
