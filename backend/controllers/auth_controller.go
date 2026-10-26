package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"laci-backend/config"
	"laci-backend/middlewares"
	"laci-backend/models"
	"laci-backend/utils"
)

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangeUsernameInput struct {
	Username string `json:"username" binding:"required,min=3"`
}

type TokenActionInput struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" binding:"required,email"`
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username dan password wajib diisi"})
		return
	}

	identifier := strings.TrimSpace(input.Username)
	var user models.User
	if err := config.DB.Where("LOWER(username) = LOWER(?) OR LOWER(nim) = LOWER(?) OR LOWER(email) = LOWER(?)", identifier, identifier, identifier).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username atau password salah"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Akun belum aktif. Silakan buka tautan aktivasi yang telah dikirimkan ke email Anda untuk membuat kata sandi.",
		})
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
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(middlewares.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat sesi login"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":    tokenString,
		"divisi":   user.Divisi,
		"username": user.Username,
		"nama":     user.Nama,
		"nim":      user.NIM,
		"email":    user.Email,
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
		"id":        user.ID,
		"nim":       user.NIM,
		"nama":      user.Nama,
		"email":     user.Email,
		"username":  user.Username,
		"divisi":    user.Divisi,
		"is_active": user.IsActive,
	})
}

// VerifyToken memvalidasi apakah token aktivasi atau reset password valid & belum expired
func VerifyToken(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	actionType := strings.ToLower(strings.TrimSpace(c.Query("type"))) // "activation" atau "reset"

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token wajib disertakan"})
		return
	}

	var user models.User
	now := time.Now()

	if actionType == "activation" {
		if err := config.DB.Where("activation_token = ? AND activation_token_expires_at > ?", token, now).First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid": false,
				"error": "Tautan aktivasi tidak valid atau telah kedaluwarsa. Hubungi Sekretaris atau Ketua Umum.",
			})
			return
		}
	} else {
		// Default to reset
		if err := config.DB.Where("reset_password_token = ? AND reset_password_expires_at > ?", token, now).First(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid": false,
				"error": "Tautan reset password tidak valid atau telah kedaluwarsa. Silakan ajukan reset kembali.",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"email":    user.Email,
		"nama":     user.Nama,
		"username": user.Username,
	})
}

// ActivateAccount memproses aktivasi akun baru dan penetapan password mandiri
func ActivateAccount(c *gin.Context) {
	var input TokenActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru minimal 6 karakter"})
		return
	}

	token := strings.TrimSpace(input.Token)
	var user models.User
	now := time.Now()

	if err := config.DB.Where("activation_token = ? AND activation_token_expires_at > ?", token, now).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token aktivasi tidak valid atau telah kedaluwarsa"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	updates := map[string]interface{}{
		"password":                     string(hashedPassword),
		"is_active":                    true,
		"activation_token":             nil,
		"activation_token_expires_at": nil,
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengaktifkan akun"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Akun berhasil diaktifkan! Silakan masuk dengan kata sandi baru Anda.",
	})
}

// ForgotPassword memicu pengiriman email tautan pemulihan kata sandi (Anti-Enumeration Guard)
func ForgotPassword(c *gin.Context) {
	var input ForgotPasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format email tidak valid"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	var user models.User

	// Jika ditemukan dan aktif, generate token dan kirim email
	if err := config.DB.Where("LOWER(email) = ? AND is_active = true", email).First(&user).Error; err == nil {
		resetToken := utils.GenerateToken(32)
		expiresAt := time.Now().Add(1 * time.Hour)

		config.DB.Model(&user).Updates(map[string]interface{}{
			"reset_password_token":        resetToken,
			"reset_password_expires_at": expiresAt,
		})

		utils.SendPasswordResetEmail(user.Email, user.Nama, user.Username, resetToken)
	}

	// Selalu kembalikan respons 200 OK generik demi keamanan anti-enumeration
	c.JSON(http.StatusOK, gin.H{
		"message": "Jika email Anda terdaftar di sistem, tautan pemulihan kata sandi telah dikirim.",
	})
}

// ResetPassword memproses penetapan kata sandi baru via tautan pemulihan
func ResetPassword(c *gin.Context) {
	var input TokenActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password baru minimal 6 karakter"})
		return
	}

	token := strings.TrimSpace(input.Token)
	var user models.User
	now := time.Now()

	if err := config.DB.Where("reset_password_token = ? AND reset_password_expires_at > ?", token, now).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tautan reset kata sandi tidak valid atau telah kedaluwarsa"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	updates := map[string]interface{}{
		"password":                   string(hashedPassword),
		"reset_password_token":       nil,
		"reset_password_expires_at": nil,
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui kata sandi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kata sandi berhasil diperbarui! Silakan masuk dengan kata sandi baru.",
	})
}

func ChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama dan password baru minimal 6 karakter wajib diisi"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userIDStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama tidak cocok"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password baru"})
		return
	}

	if err := config.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

func ChangeUsername(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userIDStr, _ := userIDVal.(string)

	var input ChangeUsernameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username wajib diisi minimal 3 karakter"})
		return
	}

	newUsername := strings.TrimSpace(input.Username)
	if strings.ContainsAny(newUsername, " \t\n\r") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username tidak boleh mengandung spasi"})
		return
	}

	var count int64
	config.DB.Model(&models.User{}).Where("LOWER(username) = LOWER(?) AND id != ?", newUsername, userIDStr).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username sudah digunakan oleh akun lain"})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", userIDStr).Update("username", newUsername).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengubah username"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Username berhasil diperbarui!",
		"username": newUsername,
	})
}
