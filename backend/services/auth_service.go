package services

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"laci-backend/middlewares"
	"laci-backend/models"
	"laci-backend/repositories"
	"laci-backend/utils"
)

var (
	ErrInvalidCredentials = errors.New("Username atau password salah")
	ErrAccountInactive    = errors.New("Akun belum aktif. Silakan buka tautan aktivasi yang telah dikirimkan ke email Anda untuk membuat kata sandi.")
	ErrUserNotFound       = errors.New("User tidak ditemukan")
	ErrInvalidOldPassword = errors.New("Password lama tidak cocok")
	ErrUsernameTaken      = errors.New("Username sudah digunakan oleh akun lain")
	ErrInvalidToken       = errors.New("Tautan tidak valid atau telah kedaluwarsa")
)

type LoginResult struct {
	Token    string `json:"token"`
	Divisi   string `json:"divisi"`
	Username string `json:"username"`
	Nama     string `json:"nama"`
	NIM      string `json:"nim"`
	Email    string `json:"email"`
}

type TokenVerifyResult struct {
	Valid    bool   `json:"valid"`
	Email    string `json:"email"`
	Nama     string `json:"nama"`
	Username string `json:"username"`
}

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Login(identifier, password string) (*LoginResult, error) {
	user, err := s.userRepo.FindByUsernameOrEmailOrNIM(identifier)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
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
		return nil, err
	}

	return &LoginResult{
		Token:    tokenString,
		Divisi:   user.Divisi,
		Username: user.Username,
		Nama:     user.Nama,
		NIM:      user.NIM,
		Email:    user.Email,
	}, nil
}

func (s *AuthService) VerifyToken(token, actionType string) (*TokenVerifyResult, error) {
	now := time.Now()
	var user *models.User
	var err error

	if strings.ToLower(actionType) == "activation" {
		user, err = s.userRepo.FindByActivationToken(token, now)
	} else {
		user, err = s.userRepo.FindByResetToken(token, now)
	}

	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	return &TokenVerifyResult{
		Valid:    true,
		Email:    user.Email,
		Nama:     user.Nama,
		Username: user.Username,
	}, nil
}

func (s *AuthService) ActivateAccount(token, password string) error {
	now := time.Now()
	user, err := s.userRepo.FindByActivationToken(token, now)
	if err != nil || user == nil {
		return ErrInvalidToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password":                     string(hashed),
		"is_active":                    true,
		"activation_token":             nil,
		"activation_token_expires_at": nil,
	}

	return s.userRepo.Updates(user, updates)
}

func (s *AuthService) ForgotPassword(email string) {
	clean := strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.FindByEmail(clean)
	if err == nil && user != nil {
		resetToken := utils.GenerateToken(32)
		expiresAt := time.Now().Add(1 * time.Hour)

		_ = s.userRepo.Updates(user, map[string]interface{}{
			"reset_password_token":        resetToken,
			"reset_password_expires_at": expiresAt,
		})

		utils.SendPasswordResetEmail(user.Email, user.Nama, user.Username, resetToken)
	}
}

func (s *AuthService) ResetPassword(token, password string) error {
	now := time.Now()
	user, err := s.userRepo.FindByResetToken(token, now)
	if err != nil || user == nil {
		return ErrInvalidToken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"password":                   string(hashed),
		"reset_password_token":       nil,
		"reset_password_expires_at": nil,
	}

	return s.userRepo.Updates(user, updates)
}

func (s *AuthService) GetProfile(userIDStr string) (*models.User, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.userRepo.FindByID(uid)
}

func (s *AuthService) ChangePassword(userIDStr, oldPassword, newPassword string) error {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return ErrInvalidOldPassword
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.userRepo.UpdateColumn(user, "password", string(hashed))
}

func (s *AuthService) ChangeUsername(userIDStr, newUsername string) (string, error) {
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", ErrUserNotFound
	}

	cleanUsername := strings.TrimSpace(newUsername)
	count, err := s.userRepo.CountByUsernameExcludingID(cleanUsername, uid)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "", ErrUsernameTaken
	}

	user, err := s.userRepo.FindByID(uid)
	if err != nil {
		return "", ErrUserNotFound
	}

	if err := s.userRepo.UpdateColumn(user, "username", cleanUsername); err != nil {
		return "", err
	}

	return cleanUsername, nil
}
