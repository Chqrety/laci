package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"laci-backend/services"
)

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenActionInput struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangeUsernameInput struct {
	Username string `json:"username" binding:"required,min=3"`
}

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password wajib diisi"})
		return
	}

	result, err := h.authService.Login(input.Username, input.Password)
	if err != nil {
		if errors.Is(err, services.ErrAccountInactive) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    result.Token,
		"divisi":   result.Divisi,
		"username": result.Username,
		"nama":     result.Nama,
		"nim":      result.NIM,
		"email":    result.Email,
	})
}

func (h *AuthHandler) VerifyToken(c *gin.Context) {
	token := c.Query("token")
	actionType := c.Query("type")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token wajib disertakan"})
		return
	}

	result, err := h.authService.VerifyToken(token, actionType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid": false,
			"error": "Tautan tidak valid atau telah kedaluwarsa.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    result.Valid,
		"email":    result.Email,
		"nama":     result.Nama,
		"username": result.Username,
	})
}

func (h *AuthHandler) ActivateAccount(c *gin.Context) {
	var input TokenActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru minimal 6 karakter"})
		return
	}

	if err := h.authService.ActivateAccount(input.Token, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Akun berhasil diaktifkan! Silakan masuk dengan kata sandi baru Anda.",
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email valid wajib disertakan"})
		return
	}

	h.authService.ForgotPassword(input.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Jika email Anda terdaftar di sistem, tautan pemulihan kata sandi telah dikirim.",
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input TokenActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru minimal 6 karakter"})
		return
	}

	if err := h.authService.ResetPassword(input.Token, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tautan reset kata sandi tidak valid atau telah kedaluwarsa"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kata sandi berhasil diperbarui! Silakan masuk dengan kata sandi baru.",
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	user, err := h.authService.GetProfile(userIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        user.ID,
		"nim":       user.NIM,
		"nama":      user.Nama,
		"email":     user.Email,
		"username":  user.Username,
		"divisi":    user.Divisi,
		"is_active": user.IsActive,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama dan password baru minimal 6 karakter wajib diisi"})
		return
	}

	if err := h.authService.ChangePassword(userIDStr, input.OldPassword, input.NewPassword); err != nil {
		if errors.Is(err, services.ErrInvalidOldPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

func (h *AuthHandler) ChangeUsername(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	var input ChangeUsernameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username wajib diisi minimal 3 karakter"})
		return
	}

	newUsername, err := h.authService.ChangeUsername(userIDStr, input.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Username berhasil diperbarui!",
		"username": newUsername,
	})
}
