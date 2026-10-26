package controllers

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"laci-backend/config"
	"laci-backend/models"
	"laci-backend/utils"
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

func sanitizeUsername(name string, nim string) string {
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

// GetUsers menampilkan daftar seluruh user/pengurus
func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Order("nama ASC, username ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// CreateUser menambahkan user baru dan mengirim email aktivasi untuk pembuatan password mandiri
func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerDivisiVal, _ := c.Get("divisi")
	callerDivisi := strings.ToLower(fmt.Sprintf("%v", callerDivisiVal))

	nim := strings.TrimSpace(input.NIM)
	nama := strings.TrimSpace(input.Nama)
	divisi := strings.ToLower(strings.TrimSpace(input.Divisi))
	if divisi == "" {
		divisi = "humas"
	}

	// Guard: Divisi ex-ketum murni berasal dari serah terima jabatan
	if divisi == "ex-ketum" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Divisi ex-ketum hanya dapat diperoleh melalui proses resmi serah terima jabatan.",
		})
		return
	}

	// Guard: Hanya Ketua Umum yang boleh mengangkat Ketua Umum baru
	if divisi == "ketum" && callerDivisi != "ketum" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Hanya Ketua Umum yang berwenang mengangkat Ketua Umum baru melalui serah terima jabatan.",
		})
		return
	}

	username := strings.TrimSpace(input.Username)
	if username == "" {
		username = sanitizeUsername(nama, nim)
	}

	// Tolak jika ada data duplikat (NIM, Nama, atau Username)
	var existing models.User
	dupQuery := config.DB.Where("LOWER(username) = LOWER(?)", username)
	if nim != "" {
		dupQuery = dupQuery.Or("LOWER(nim) = LOWER(?)", nim)
	}
	if nama != "" {
		dupQuery = dupQuery.Or("LOWER(nama) = LOWER(?)", nama)
	}
	if err := dupQuery.First(&existing).Error; err == nil {
		if nim != "" && strings.EqualFold(existing.NIM, nim) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "NIM sudah terdaftar: " + nim})
			return
		}
		if nama != "" && strings.EqualFold(existing.Nama, nama) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nama pengurus sudah terdaftar: " + nama})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username sudah digunakan: " + username})
		return
	}

	// Buat password acak terkunci yang tidak dapat digunakan sebelum aktivasi
	lockedPasswordHash, err := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan akun"})
		return
	}

	activationToken := utils.GenerateToken(32)
	tokenExpiry := time.Now().Add(24 * time.Hour)

	newUser := models.User{
		ID:                       uuid.New(),
		NIM:                      nim,
		Nama:                     nama,
		Email:                    strings.TrimSpace(input.Email),
		Username:                 username,
		Password:                 string(lockedPasswordHash),
		Divisi:                   divisi,
		IsActive:                 false,
		ActivationToken:          &activationToken,
		ActivationTokenExpiresAt: &tokenExpiry,
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kirim email aktivasi secara asynchronous
	if newUser.Email != "" {
		utils.SendActivationEmail(newUser.Email, newUser.Nama, newUser.Username, activationToken)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User berhasil didaftarkan. Email aktivasi telah dikirimkan ke " + newUser.Email,
		"user":    newUser,
	})
}

// UpdateUser mengubah data pengurus atau divisi
func UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerDivisiVal, _ := c.Get("divisi")
	callerDivisi := strings.ToLower(fmt.Sprintf("%v", callerDivisiVal))

	updates := map[string]interface{}{}
	if input.NIM != "" {
		newNIM := strings.TrimSpace(input.NIM)
		var count int64
		config.DB.Model(&models.User{}).Where("LOWER(nim) = LOWER(?) AND id != ?", newNIM, userUUID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "NIM sudah terdaftar pada akun lain"})
			return
		}
		updates["nim"] = newNIM
	}
	if input.Nama != "" {
		newNama := strings.TrimSpace(input.Nama)
		var count int64
		config.DB.Model(&models.User{}).Where("LOWER(nama) = LOWER(?) AND id != ?", newNama, userUUID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Nama sudah terdaftar pada akun lain"})
			return
		}
		updates["nama"] = newNama
	}
	if input.Email != "" {
		updates["email"] = strings.TrimSpace(input.Email)
	}
	if input.Username != "" {
		newUsername := strings.TrimSpace(input.Username)
		var count int64
		config.DB.Model(&models.User{}).Where("LOWER(username) = LOWER(?) AND id != ?", newUsername, userUUID).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username sudah digunakan oleh akun lain"})
			return
		}
		updates["username"] = newUsername
	}
	if input.Divisi != "" {
		newDiv := strings.ToLower(strings.TrimSpace(input.Divisi))
		if newDiv == "ex-ketum" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Divisi ex-ketum hanya dapat diperoleh melalui proses resmi serah terima jabatan."})
			return
		}
		if newDiv == "ketum" && callerDivisi != "ketum" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Hanya Ketua Umum yang dapat mengangkat Ketua Umum baru."})
			return
		}
		updates["divisi"] = newDiv
	}
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err == nil {
			updates["password"] = string(hashed)
		}
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config.DB.First(&user, "id = ?", userUUID)
	c.JSON(http.StatusOK, user)
}

// TransferKetum melakukan suksesi kepemimpinan satu arah (Ketum aktif -> Ex-Ketum, Target -> Ketum)
func TransferKetum(c *gin.Context) {
	callerIDVal, _ := c.Get("userID")
	callerIDStr, _ := callerIDVal.(string)
	callerUUID, err := uuid.Parse(callerIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Sesi tidak valid"})
		return
	}

	var input TransferKetumInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user dan konfirmasi password wajib diisi"})
		return
	}

	var callerUser models.User
	if err := config.DB.First(&callerUser, "id = ?", callerUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Akun Ketua Umum tidak ditemukan"})
		return
	}

	if strings.ToLower(callerUser.Divisi) != "ketum" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya Ketua Umum aktif yang dapat melakukan serah terima jabatan"})
		return
	}

	// Verifikasi keamanan password Ketua Umum saat ini
	if err := bcrypt.CompareHashAndPassword([]byte(callerUser.Password), []byte(input.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Konfirmasi password salah"})
		return
	}

	targetUUID, err := uuid.Parse(input.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID target user tidak valid"})
		return
	}

	if targetUUID == callerUUID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Anda sudah menjabat sebagai Ketua Umum"})
		return
	}

	var targetUser models.User
	if err := config.DB.First(&targetUser, "id = ?", targetUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Anggota target suksesi tidak ditemukan"})
		return
	}

	// Eksekusi suksesi atomik
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&callerUser).Update("divisi", "ex-ketum").Error; err != nil {
			return err
		}
		if err := tx.Model(&targetUser).Update("divisi", "ketum").Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses serah terima jabatan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    fmt.Sprintf("Suksesi berhasil! Jabatan Ketua Umum telah diserahkan kepada %s.", targetUser.Nama),
		"ketum_baru": targetUser.Nama,
	})
}

// DeleteUser menghapus user
func DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	currentUserID, _ := c.Get("userID")
	if currentUserIDStr, ok := currentUserID.(string); ok && userUUID.String() == currentUserIDStr {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak dapat menghapus akun Anda sendiri"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User tidak ditemukan"})
		return
	}

	// Mencegah penghapusan Ketua Umum aktif
	if strings.ToLower(user.Divisi) == "ketum" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Ketua Umum aktif tidak dapat dihapus. Silakan transfer jabatan terlebih dahulu."})
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User berhasil dihapus"})
}

type BatchDeleteUsersInput struct {
	IDs []string `json:"ids" binding:"required"`
}

// BatchDeleteUsers menghapus banyak user sekaligus berdasarkan daftar ID
func BatchDeleteUsers(c *gin.Context) {
	var input BatchDeleteUsersInput
	if err := c.ShouldBindJSON(&input); err != nil || len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Daftar ID user wajib diisi"})
		return
	}

	currentUserID, _ := c.Get("userID")
	currentUserIDStr, _ := currentUserID.(string)

	var parsedUUIDs []uuid.UUID
	for _, idStr := range input.IDs {
		if idStr == currentUserIDStr {
			continue
		}
		if u, err := uuid.Parse(idStr); err == nil {
			parsedUUIDs = append(parsedUUIDs, u)
		}
	}

	if len(parsedUUIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada ID valid yang diberikan"})
		return
	}

	// Lindungi akun ketum dari batch delete
	if err := config.DB.Where("id IN ? AND LOWER(divisi) != 'ketum'", parsedUUIDs).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data pengurus: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Berhasil menghapus pengurus yang dipilih"),
		"deleted": len(parsedUUIDs),
	})
}

type ImportSpreadsheetInput struct {
	SheetURL string `json:"sheet_url" binding:"required"`
}

func parseAndSaveUsersFromCSV(r io.Reader) ([]models.User, []string, error) {
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

		// Keamanan: CSV dilarang mengimpor akun sebagai ketum
		if divisi == "ketum" {
			divisi = "pengurus"
		}

		if nama == "" && nim != "" {
			nama = nim
		}
		if nama == "" {
			continue
		}

		username := sanitizeUsername(nama, nim)

		var exists int64
		config.DB.Model(&models.User{}).Where("username = ? OR (nim != '' AND nim = ?)", username, nim).Count(&exists)
		if exists > 0 {
			skipped = append(skipped, fmt.Sprintf("%s (%s) - username/nim sudah ada", nama, username))
			continue
		}

		activationToken := utils.GenerateToken(32)
		tokenExpiry := time.Now().Add(24 * time.Hour)
		dummyLockedPassword, _ := bcrypt.GenerateFromPassword([]byte(uuid.New().String()), bcrypt.DefaultCost)

		user := models.User{
			ID:                       uuid.New(),
			NIM:                      nim,
			Nama:                     nama,
			Email:                    email,
			Username:                 username,
			Password:                 string(dummyLockedPassword),
			Divisi:                   divisi,
			IsActive:                 false,
			ActivationToken:          &activationToken,
			ActivationTokenExpiresAt: &tokenExpiry,
		}

		if err := config.DB.Create(&user).Error; err != nil {
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

// ImportUsersCSV memproses file CSV (kolom: nim, nama, email, divisi)
func ImportUsersCSV(c *gin.Context) {
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

	imported, skipped, err := parseAndSaveUsersFromCSV(file)
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

// extractSpreadsheetID mengekstrak Google Spreadsheet ID dari link
func extractSpreadsheetID(link string) string {
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

// ImportUsersSpreadsheet menarik data langsung dari link publik Google Sheets
func ImportUsersSpreadsheet(c *gin.Context) {
	var input ImportSpreadsheetInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.SheetURL) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Link Google Sheets wajib diisi"})
		return
	}

	sheetID := extractSpreadsheetID(input.SheetURL)
	if sheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format link Google Sheets tidak valid"})
		return
	}

	exportURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if strings.Contains(req.URL.Host, "accounts.google.com") {
				return fmt.Errorf("spreadsheet bersifat privat. Pastikan akses diset ke 'Siapa saja yang memiliki link'")
			}
			return nil
		},
	}

	resp, err := client.Get(exportURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal mengakses spreadsheet: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal mengunduh spreadsheet. Pastikan izin sharing terbuka publik."})
		return
	}

	imported, skipped, err := parseAndSaveUsersFromCSV(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca konten spreadsheet: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Berhasil mengimpor %d user. Tautan aktivasi telah dikirim via email.", len(imported)),
		"imported": imported,
		"skipped":  skipped,
	})
}
