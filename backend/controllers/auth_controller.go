package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"laci-backend/config"
	"laci-backend/middlewares"
	"laci-backend/models"
)

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ? OR nim = ? OR email = ?", input.Username, input.Username, input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	claims := jwt.MapClaims{
		"sub":      user.ID.String(),
		"username": user.Username,
		"divisi":   user.Divisi,
		"role":     user.Divisi,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middlewares.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":                     tokenString,
		"divisi":                    user.Divisi,
		"role":                      user.Divisi,
		"username":                  user.Username,
		"nama":                      user.Nama,
		"nim":                       user.NIM,
		"email":                     user.Email,
		"is_password_default":       user.IsPasswordDefault,
		"has_seen_password_prompt":  user.HasSeenPasswordPrompt,
	})
}

func GetProfile(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	var user models.User
	if err := config.DB.First(&user, "id = ?", userIDStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":                        user.ID,
		"nim":                       user.NIM,
		"nama":                      user.Nama,
		"email":                     user.Email,
		"username":                  user.Username,
		"divisi":                    user.Divisi,
		"role":                      user.Divisi,
		"is_password_default":       user.IsPasswordDefault,
		"has_seen_password_prompt":  user.HasSeenPasswordPrompt,
	})
}

func DismissPasswordPrompt(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	if err := config.DB.Model(&models.User{}).Where("id = ?", userIDStr).Update("has_seen_password_prompt", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan preferensi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Prompt telah dimatikan"})
}

func ChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru minimal 6 karakter"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userIDStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// Jika bukan default dan old password diberikan, verifikasi dulu
	if input.OldPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama tidak cocok"})
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password baru"})
		return
	}

	updates := map[string]interface{}{
		"password":                 string(hashedPassword),
		"display_password":         input.NewPassword,
		"is_password_default":      false,
		"has_seen_password_prompt": true,
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}
