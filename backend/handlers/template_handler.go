package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"laci-backend/services"
)

type SyncTemplatesInput struct {
	FolderURL string `json:"folder_url" binding:"required"`
	GASURL    string `json:"gas_url"`
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

type TemplateHandler struct {
	templateService *services.TemplateService
}

func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{templateService: templateService}
}

func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	templates, err := h.templateService.GetActiveTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil daftar template: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (h *TemplateHandler) SyncTemplates(c *gin.Context) {
	var input SyncTemplatesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link folder Google Drive wajib diisi"})
		return
	}

	saved, err := h.templateService.SyncTemplates(input.FolderURL, input.GASURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Berhasil menyinkronkan %d template surat", len(saved)),
		"templates": saved,
		"count":     len(saved),
	})
}

func (h *TemplateHandler) GenerateSurat(c *gin.Context) {
	var input GenerateSuratInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data input tidak lengkap: " + err.Error()})
		return
	}

	surat, gasResp, err := h.templateService.GenerateSurat(services.GenerateSuratParams{
		TemplateID:   input.TemplateID,
		FolderURL:    input.FolderURL,
		FileName:     input.FileName,
		NomorSurat:   input.NomorSurat,
		Perihal:      input.Perihal,
		Tujuan:       input.Tujuan,
		PICNama:      input.PICNama,
		ProkerID:     input.ProkerID,
		Replacements: input.Replacements,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":   "Surat berhasil dibuat dan disimpan di Google Drive",
		"surat":     surat,
		"gdocs_url": gasResp.GDocsURL,
		"pdf_url":   gasResp.PDFURL,
		"doc_id":    gasResp.DocID,
	})
}

func (h *TemplateHandler) GetAppSettings(c *gin.Context) {
	settings, err := h.templateService.GetAppSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *TemplateHandler) UpdateAppSettings(c *gin.Context) {
	var payload map[string]string
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format payload tidak valid"})
		return
	}

	if err := h.templateService.UpdateAppSettings(payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengaturan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengaturan berhasil diperbarui"})
}
