package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"laci-backend/services"
)

type CreateInstansiInput struct {
	NamaInstansi string `json:"nama_instansi" binding:"required"`
	Kategori     string `json:"kategori"`
	Keterangan   string `json:"keterangan"`
}

type InstansiHandler struct {
	instansiService *services.InstansiService
}

func NewInstansiHandler(instansiService *services.InstansiService) *InstansiHandler {
	return &InstansiHandler{instansiService: instansiService}
}

func (h *InstansiHandler) GetInstansi(c *gin.Context) {
	list, err := h.instansiService.GetAllInstansi()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *InstansiHandler) CreateInstansi(c *gin.Context) {
	var input CreateInstansiInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama instansi wajib diisi"})
		return
	}

	ins, err := h.instansiService.CreateInstansi(input.NamaInstansi, input.Kategori, input.Keterangan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ins)
}

func (h *InstansiHandler) UpdateInstansi(c *gin.Context) {
	idParam := c.Param("id")
	insUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	var input CreateInstansiInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama instansi wajib diisi"})
		return
	}

	ins, err := h.instansiService.UpdateInstansi(insUUID, input.NamaInstansi, input.Kategori, input.Keterangan)
	if err != nil {
		if errors.Is(err, services.ErrInstansiNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instansi tidak ditemukan"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ins)
}

func (h *InstansiHandler) DeleteInstansi(c *gin.Context) {
	idParam := c.Param("id")
	insUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format ID tidak valid"})
		return
	}

	if err := h.instansiService.DeleteInstansi(insUUID); err != nil {
		if errors.Is(err, services.ErrInstansiNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Instansi tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Instansi berhasil dihapus"})
}
