package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"laci-backend/models"
	"laci-backend/repositories"
	"laci-backend/utils"
)

var (
	ErrExKetumRestricted     = errors.New("Divisi ex-ketum hanya dapat diperoleh melalui proses resmi serah terima jabatan.")
	ErrUnauthorizedKetumRole = errors.New("Hanya Ketua Umum yang berwenang mengangkat Ketua Umum baru melalui serah terima jabatan.")
	ErrCannotDeleteSelf      = errors.New("Tidak dapat menghapus akun Anda sendiri")
	ErrCannotDeleteKetum     = errors.New("Akun Ketua Umum tidak dapat dihapus")
	ErrInvalidTargetUser     = errors.New("Akun tujuan tidak valid atau belum aktif")
	ErrTargetAlreadyKetum    = errors.New("Target user sudah menjabat sebagai Ketua Umum")
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func SanitizeUsername(name string, nim string) string {
	trimmedNIM := strings.TrimSpace(nim)
	if trimmedNIM != "" {
		return trimmedNIM
	}
	cleaned := strings.ToLower(strings.Join(strings.Fields(name), "."))
	if cleaned == "" {
		return fmt.Sprintf("user%d", uuid.New().ID()%10000)
	}
	return cleaned
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAll()
}

func (s *UserService) CreateUser(callerDivisi, nim, nama, email, username, divisi string) (*models.User, error) {
	divisi = strings.ToLower(strings.TrimSpace(divisi))
	if divisi == "" {
		divisi = "humas"
	}

	if divisi == "ex-ketum" {
		return nil, ErrExKetumRestricted
	}
	if divisi == "ketum" && callerDivisi != "ketum" {
		return nil, ErrUnauthorizedKetumRole
	}

	nim = strings.TrimSpace(nim)
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return nil, errors.New("Nama wajib diisi")
	}

	username = strings.TrimSpace(username)
	if username == "" {
		username = SanitizeUsername(nama, nim)
	}

	if s.userRepo.ExistsByUsernameOrNIM(username, nim) {
		return nil, fmt.Errorf("Username '%s' atau NIM '%s' sudah terdaftar", username, nim)
	}

	countNama, err := s.userRepo.CountByNamaExcludingID(nama, uuid.Nil)
	if err == nil && countNama > 0 {
		return nil, fmt.Errorf("Nama pengurus sudah terdaftar: %s", nama)
	}

	lockedPassword, err := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("Gagal menyiapkan akun")
	}

	activationToken := utils.GenerateToken(32)
	tokenExpiry := time.Now().Add(24 * time.Hour)

	newUser := models.User{
		ID:                       uuid.New(),
		NIM:                      nim,
		Nama:                     nama,
		Email:                    strings.TrimSpace(email),
		Username:                 username,
		Password:                 string(lockedPassword),
		Divisi:                   divisi,
		IsActive:                 false,
		ActivationToken:          &activationToken,
		ActivationTokenExpiresAt: &tokenExpiry,
	}

	if err := s.userRepo.Create(&newUser); err != nil {
		return nil, err
	}

	if newUser.Email != "" {
		utils.SendActivationEmail(newUser.Email, newUser.Nama, newUser.Username, activationToken)
	}

	return &newUser, nil
}

func (s *UserService) UpdateUser(userUUID uuid.UUID, callerDivisi, nim, nama, email, username, divisi, password string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userUUID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	updates := map[string]interface{}{}

	nim = strings.TrimSpace(nim)
	if nim != "" {
		count, err := s.userRepo.CountByNIMExcludingID(nim, userUUID)
		if err != nil || count > 0 {
			return nil, errors.New("NIM sudah terdaftar pada akun lain")
		}
		updates["nim"] = nim
	}

	nama = strings.TrimSpace(nama)
	if nama != "" {
		count, err := s.userRepo.CountByNamaExcludingID(nama, userUUID)
		if err != nil || count > 0 {
			return nil, errors.New("Nama sudah terdaftar pada akun lain")
		}
		updates["nama"] = nama
	}

	if email != "" {
		updates["email"] = strings.TrimSpace(email)
	}

	username = strings.TrimSpace(username)
	if username != "" {
		count, err := s.userRepo.CountByUsernameExcludingID(username, userUUID)
		if err != nil || count > 0 {
			return nil, errors.New("Username sudah digunakan oleh akun lain")
		}
		updates["username"] = username
	}

	divisi = strings.ToLower(strings.TrimSpace(divisi))
	if divisi != "" {
		if divisi == "ex-ketum" && user.Divisi != "ex-ketum" {
			return nil, ErrExKetumRestricted
		}
		if divisi == "ketum" && callerDivisi != "ketum" {
			return nil, ErrUnauthorizedKetumRole
		}
		updates["divisi"] = divisi
	}

	if password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err == nil {
			updates["password"] = string(hashed)
		}
	}

	if err := s.userRepo.Updates(user, updates); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(callerID, targetID uuid.UUID) error {
	if callerID == targetID {
		return ErrCannotDeleteSelf
	}

	target, err := s.userRepo.FindByID(targetID)
	if err != nil {
		return ErrUserNotFound
	}

	if strings.ToLower(target.Divisi) == "ketum" {
		return ErrCannotDeleteKetum
	}

	return s.userRepo.Delete(targetID)
}

func (s *UserService) DeleteBatch(callerID uuid.UUID, ids []uuid.UUID) (int64, error) {
	var validIDs []uuid.UUID
	for _, id := range ids {
		if id == callerID {
			continue
		}
		user, err := s.userRepo.FindByID(id)
		if err == nil && strings.ToLower(user.Divisi) != "ketum" {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return 0, errors.New("Tidak ada akun yang valid untuk dihapus")
	}

	return s.userRepo.DeleteBatch(validIDs)
}

func (s *UserService) TransferKetum(currentKetumID uuid.UUID, targetUserIDStr, currentPassword string) (*models.User, error) {
	currentKetum, err := s.userRepo.FindByID(currentKetumID)
	if err != nil {
		return nil, errors.New("Akun Ketua Umum tidak ditemukan")
	}

	if strings.ToLower(currentKetum.Divisi) != "ketum" {
		return nil, errors.New("Hanya Ketua Umum aktif yang dapat melakukan serah terima jabatan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentKetum.Password), []byte(currentPassword)); err != nil {
		return nil, errors.New("Kata sandi Ketua Umum saat ini tidak sesuai")
	}

	targetUUID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		return nil, errors.New("Target user ID tidak valid")
	}

	if targetUUID == currentKetumID {
		return nil, errors.New("Anda sudah menjabat sebagai Ketua Umum")
	}

	targetUser, err := s.userRepo.FindByID(targetUUID)
	if err != nil || !targetUser.IsActive {
		return nil, ErrInvalidTargetUser
	}

	if strings.ToLower(targetUser.Divisi) == "ketum" {
		return nil, ErrTargetAlreadyKetum
	}

	if err := s.userRepo.TransferKetum(currentKetumID, targetUUID); err != nil {
		return nil, fmt.Errorf("Gagal melakukan serah terima jabatan: %w", err)
	}

	targetUser.Divisi = "ketum"
	return targetUser, nil
}

func (s *UserService) ParseAndSaveUsersFromCSV(r io.Reader) ([]models.User, []string, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true
	reader.LazyQuotes = true

	var imported []models.User
	var skipped []string
	isFirstRow := true

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		if len(record) == 0 {
			continue
		}

		if isFirstRow {
			isFirstRow = false
			firstCell := strings.ToLower(strings.TrimSpace(record[0]))
			if firstCell == "nim" || firstCell == "nama" || firstCell == "name" {
				continue
			}
		}

		nim := ""
		nama := ""
		email := ""
		divisi := "humas"

		if len(record) >= 1 {
			nim = strings.TrimSpace(record[0])
		}
		if len(record) >= 2 {
			nama = strings.TrimSpace(record[1])
		}
		if len(record) >= 3 {
			email = strings.TrimSpace(record[2])
		}
		if len(record) >= 4 && strings.TrimSpace(record[3]) != "" {
			divisi = strings.ToLower(strings.TrimSpace(record[3]))
		}

		if divisi == "ketum" {
			divisi = "pengurus"
		}

		if nama == "" && nim != "" {
			nama = nim
		}
		if nama == "" {
			continue
		}

		username := SanitizeUsername(nama, nim)

		if s.userRepo.ExistsByUsernameOrNIM(username, nim) {
			skipped = append(skipped, fmt.Sprintf("%s (%s) - username/nim sudah ada", nama, username))
			continue
		}

		activationToken := utils.GenerateToken(32)
		tokenExpiry := time.Now().Add(24 * time.Hour)
		dummyPassword, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)

		user := models.User{
			ID:                       uuid.New(),
			NIM:                      nim,
			Nama:                     nama,
			Email:                    email,
			Username:                 username,
			Password:                 string(dummyPassword),
			Divisi:                   divisi,
			IsActive:                 false,
			ActivationToken:          &activationToken,
			ActivationTokenExpiresAt: &tokenExpiry,
		}

		if err := s.userRepo.Create(&user); err != nil {
			skipped = append(skipped, fmt.Sprintf("%s - db error: %v", nama, err))
			continue
		}

		if user.Email != "" {
			utils.SendActivationEmail(user.Email, user.Nama, user.Username, activationToken)
		}

		imported = append(imported, user)
	}

	return imported, skipped, nil
}

func ExtractSpreadsheetID(link string) string {
	link = strings.TrimSpace(link)
	if strings.Contains(link, "/d/") {
		parts := strings.Split(link, "/d/")
		if len(parts) > 1 {
			subParts := strings.Split(parts[1], "/")
			return subParts[0]
		}
	}
	return link
}

func (s *UserService) ImportSpreadsheet(sheetURL string) ([]models.User, []string, error) {
	sheetID := ExtractSpreadsheetID(sheetURL)
	if sheetID == "" {
		return nil, nil, errors.New("Format link Google Sheets tidak valid")
	}

	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if strings.Contains(req.URL.Host, "accounts.google.com") {
				return errors.New("spreadsheet bersifat privat. Pastikan akses diset ke 'Siapa saja yang memiliki link'")
			}
			return nil
		},
	}

	resp, err := client.Get(exportURL)
	if err != nil {
		return nil, nil, fmt.Errorf("Gagal mengakses spreadsheet: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, errors.New("Gagal mengunduh spreadsheet. Pastikan izin sharing terbuka publik.")
	}

	return s.ParseAndSaveUsersFromCSV(resp.Body)
}
