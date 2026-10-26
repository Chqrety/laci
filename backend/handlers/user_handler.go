package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"laci-backend/services"
)

type CreateUserInput struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama" binding:"required"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Divisi   string `json:"divisi"`
}

type UpdateUserInput struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Divisi   string `json:"divisi"`
	Password string `json:"password"`
}

type TransferKetumInput struct {
	TargetUserID    string `json:"target_user_id" binding:"required"`
	CurrentPassword string `json:"current_password" binding:"required"`
}

type BatchDeleteInput struct {
	IDs []string `json:"ids" binding:"required"`
}

type ImportSpreadsheetInput struct {
	SheetURL string `json:"sheet_url" binding:"required"`
}

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerDivisiVal, _ := c.Get("divisi")
	callerDivisi := strings.ToLower(fmt.Sprintf("%v", callerDivisiVal))

	newUser, err := h.userService.CreateUser(
		callerDivisi,
		input.NIM,
		input.Nama,
		input.Email,
		input.Username,
		input.Divisi,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User berhasil didaftarkan. Email aktivasi telah dikirimkan ke " + newUser.Email,
		"user":    newUser,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerDivisiVal, _ := c.Get("divisi")
	callerDivisi := strings.ToLower(fmt.Sprintf("%v", callerDivisiVal))

	updatedUser, err := h.userService.UpdateUser(
		userUUID,
		callerDivisi,
		input.NIM,
		input.Nama,
		input.Email,
		input.Username,
		input.Divisi,
		input.Password,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data pengurus berhasil diperbarui",
		"user":    updatedUser,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	targetUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	callerIDVal, _ := c.Get("userID")
	callerUUID, _ := uuid.Parse(fmt.Sprintf("%v", callerIDVal))

	if err := h.userService.DeleteUser(callerUUID, targetUUID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil dihapus"})
}

func (h *UserHandler) BatchDeleteUsers(c *gin.Context) {
	var input BatchDeleteInput
	if err := c.ShouldBindJSON(&input); err != nil || len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Daftar ID user wajib disertakan"})
		return
	}

	var parsedIDs []uuid.UUID
	for _, idStr := range input.IDs {
		if parsed, err := uuid.Parse(idStr); err == nil {
			parsedIDs = append(parsedIDs, parsed)
		}
	}

	callerIDVal, _ := c.Get("userID")
	callerUUID, _ := uuid.Parse(fmt.Sprintf("%v", callerIDVal))

	deletedCount, err := h.userService.DeleteBatch(callerUUID, parsedIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Berhasil menghapus %d akun", deletedCount),
		"deleted": deletedCount,
	})
}

func (h *UserHandler) TransferKetum(c *gin.Context) {
	callerIDVal, _ := c.Get("userID")
	callerUUID, _ := uuid.Parse(fmt.Sprintf("%v", callerIDVal))

	var input TransferKetumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user ID dan kata sandi Ketum saat ini wajib diisi"})
		return
	}

	newKetum, err := h.userService.TransferKetum(callerUUID, input.TargetUserID, input.CurrentPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Serah terima jabatan berhasil! '%s' sekarang menjabat sebagai Ketua Umum.", newKetum.Nama),
		"ketum_baru": newKetum.Nama,
	})
}

func (h *UserHandler) ImportUsersCSV(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File CSV wajib diunggah"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca file"})
		return
	}
	defer file.Close()

	imported, skipped, err := h.userService.ParseAndSaveUsersFromCSV(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format CSV tidak valid: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Berhasil mengimpor %d user. Tautan aktivasi telah dikirim via email.", len(imported)),
		"imported": imported,
		"skipped":  skipped,
	})
}

func (h *UserHandler) ImportUsersSpreadsheet(c *gin.Context) {
	var input ImportSpreadsheetInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.SheetURL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link Google Sheets wajib diisi"})
		return
	}

	imported, skipped, err := h.userService.ImportSpreadsheet(input.SheetURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Berhasil mengimpor %d user. Tautan aktivasi telah dikirim via email.", len(imported)),
		"imported": imported,
		"skipped":  skipped,
	})
}
