package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"laci-backend/services"
)

type CreateSuratInput struct {
	NomorSurat    string `json:"nomor_surat" binding:"required"`
	Perihal       string `json:"perihal" binding:"required"`
	Tujuan        string `json:"tujuan"`
	StatusSaatIni string `json:"status_saat_ini"`
	PICNama       string `json:"pic_nama" binding:"required"`
	Catatan       string `json:"catatan"`
}

type UpdateSuratInput struct {
	StatusSaatIni string  `json:"status_saat_ini"`
	Tujuan        *string `json:"tujuan"`
	Catatan       *string `json:"catatan"`
	CatatanLog    *string `json:"catatan_log"`
}

type SuratHandler struct {
	suratService *services.SuratService
}

func NewSuratHandler(suratService *services.SuratService) *SuratHandler {
	return &SuratHandler{suratService: suratService}
}

func (h *SuratHandler) GetSurat(c *gin.Context) {
	listSurat, err := h.suratService.GetAllSurat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, listSurat)
}

func (h *SuratHandler) CreateSurat(c *gin.Context) {
	var input CreateSuratInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	divisiVal, _ := c.Get("divisi")
	divisi, _ := divisiVal.(string)

	surat, err := h.suratService.CreateSurat(
		divisi,
		input.NomorSurat,
		input.Perihal,
		input.Tujuan,
		input.StatusSaatIni,
		input.PICNama,
		input.Catatan,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, surat)
}

func (h *SuratHandler) UpdateSurat(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var input UpdateSuratInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	surat, log, err := h.suratService.UpdateSurat(
		suratUUID,
		input.StatusSaatIni,
		input.Tujuan,
		input.Catatan,
		input.CatatanLog,
	)
	if err != nil {
		if errors.Is(err, services.ErrSuratNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Surat updated successfully",
		"surat":   surat,
		"log":     log,
	})
}

func (h *SuratHandler) UploadSurat(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	divisiVal, _ := c.Get("divisi")
	divisi := fmt.Sprintf("%v", divisiVal)

	file, err := c.FormFile("file")
	if err != nil {
		file, err = c.FormFile("arsip")
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Key 'file' is required."})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}
	defer src.Close()

	fileArsipURL, updatedSurat, err := h.suratService.SaveUploadedArsip(suratUUID, divisi, file.Filename, src)
	if err != nil {
		if errors.Is(err, services.ErrArsipRestricted) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrSuratNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "File arsip berhasil diunggah",
		"arsip_url": fileArsipURL,
		"surat":     updatedSurat,
	})
}

func (h *SuratHandler) DeleteSurat(c *gin.Context) {
	idParam := c.Param("id")
	suratUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := h.suratService.DeleteSurat(suratUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Surat dan log terkait berhasil dihapus"})
}
