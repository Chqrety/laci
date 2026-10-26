package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"laci-backend/services"
)

type CreateProkerInput struct {
	NamaProker string `json:"nama_proker" binding:"required"`
	Deskripsi  string `json:"deskripsi"`
}

type ProkerHandler struct {
	prokerService *services.ProkerService
}

func NewProkerHandler(prokerService *services.ProkerService) *ProkerHandler {
	return &ProkerHandler{prokerService: prokerService}
}

func (h *ProkerHandler) GetProkers(c *gin.Context) {
	list, err := h.prokerService.GetAllProkers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *ProkerHandler) CreateProker(c *gin.Context) {
	var input CreateProkerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker is required"})
		return
	}

	proker, err := h.prokerService.CreateProker(input.NamaProker, input.Deskripsi)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, proker)
}

func (h *ProkerHandler) UpdateProker(c *gin.Context) {
	idParam := c.Param("id")
	prokerUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	var input CreateProkerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker is required"})
		return
	}

	proker, err := h.prokerService.UpdateProker(prokerUUID, input.NamaProker, input.Deskripsi)
	if err != nil {
		if errors.Is(err, services.ErrProkerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proker not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proker)
}

func (h *ProkerHandler) DeleteProker(c *gin.Context) {
	idParam := c.Param("id")
	prokerUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := h.prokerService.DeleteProker(prokerUUID); err != nil {
		if errors.Is(err, services.ErrProkerNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Proker not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Program kerja berhasil dihapus"})
}
