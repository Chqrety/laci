package controllers

import (
	"crypto/rand"
	"encoding/csv"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"laci-backend/config"
	"laci-backend/models"
	"laci-backend/utils"
)

type CreateUserInput struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama" binding:"required"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
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

// generateRandomPassword menghasilkan password acak sederhana 8 karakter
func generateRandomPassword(length int) string {
	const charset = "abcdefghjkmnpqrstuvwxyz23456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "doscom123"
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
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

// GetUsers menampilkan daftar seluruh user/pengurus.
// Khusus Sekre, display_password dikembalikan agar bisa di-reveal di UI.
func GetUsers(c *gin.Context) {
	divisiVal, _ := c.Get("divisi")
	divisi, _ := divisiVal.(string)

	var users []models.User
	if err := config.DB.Order("nama ASC, username ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Jika bukan sekre, jangan kirim display_password
	if strings.ToLower(divisi) != "sekre" {
		for i := range users {
			users[i].DisplayPassword = ""
		}
	}

	c.JSON(http.StatusOK, users)
}

// CreateUser menambahkan user manual oleh Sekre
func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nim := strings.TrimSpace(input.NIM)
	nama := strings.TrimSpace(input.Nama)
	divisi := strings.ToLower(strings.TrimSpace(input.Divisi))
	if divisi != "sekre" && divisi != "humas" {
		divisi = "humas"
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

	rawPassword := strings.TrimSpace(input.Password)
	if rawPassword == "" {
		rawPassword = generateRandomPassword(8)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	newUser := models.User{
		ID:                    uuid.New(),
		NIM:                   strings.TrimSpace(input.NIM),
		Nama:                  strings.TrimSpace(input.Nama),
		Email:                 strings.TrimSpace(input.Email),
		Username:              username,
		Password:              string(hashedPassword),
		DisplayPassword:       rawPassword,
		Divisi:                divisi,
		IsPasswordDefault:     true,
		HasSeenPasswordPrompt: false,
	}

	if err := config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kirim email kredensial secara background jika email diisi
	if newUser.Email != "" {
		utils.SendAccountCredentialsEmail(utils.MailAccountPayload{
			ToEmail:  newUser.Email,
			Nama:     newUser.Nama,
			NIM:      newUser.NIM,
			Username: newUser.Username,
			Password: rawPassword,
			Divisi:   newUser.Divisi,
		})
	}

	c.JSON(http.StatusCreated, newUser)
}

// UpdateUser mengubah data pengurus atau mereset passwordnya
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

	currentUserID, _ := c.Get("userID")
	currentUserIDStr, _ := currentUserID.(string)

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
		if newDiv != "sekre" && newDiv != "humas" {
			newDiv = "humas"
		}
		if userUUID.String() == currentUserIDStr && newDiv != "sekre" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak dapat mengubah divisi akun sendiri menjadi non-sekre"})
			return
		}
		updates["divisi"] = newDiv
	}
	if input.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err == nil {
			updates["password"] = string(hashed)
			updates["display_password"] = input.Password
			updates["is_password_default"] = true
			updates["has_seen_password_prompt"] = false
		}
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config.DB.First(&user, "id = ?", userUUID)
	c.JSON(http.StatusOK, user)
}

// ResetPasswordPrompt memicu ulang modal pengingat ganti password untuk user tertentu
func ResetPasswordPrompt(c *gin.Context) {
	idParam := c.Param("id")
	userUUID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", userUUID).Update("has_seen_password_prompt", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mereset reminder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder ganti password berhasil diaktifkan kembali untuk user ini"})
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

	if err := config.DB.Delete(&models.User{}, "id = ?", userUUID).Error; err != nil {
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
			continue // Lindungi akun sendiri agar tidak terhapus saat batch delete
		}
		if u, err := uuid.Parse(idStr); err == nil {
			parsedUUIDs = append(parsedUUIDs, u)
		}
	}

	if len(parsedUUIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada ID valid yang diberikan"})
		return
	}

	if err := config.DB.Where("id IN ?", parsedUUIDs).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus data pengurus: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Berhasil menghapus %d pengurus", len(parsedUUIDs)),
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

		// Lewati header jika baris pertama berisi kata kunci
		if isFirstRow {
			isFirstRow = false
			firstCell := strings.ToLower(strings.TrimSpace(record[0]))
			if firstCell == "nim" || firstCell == "nama" || firstCell == "name" {
				continue
			}
		}

		// Kolom yang diharapkan: nim, nama, email, divisi
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
			divisi = strings.TrimSpace(record[3])
		}

		// Jika nama kosong tapi ada nim, gunakan nim
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

		rawPassword := generateRandomPassword(8)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
		if err != nil {
			skipped = append(skipped, fmt.Sprintf("%s - gagal hash password", nama))
			continue
		}

		user := models.User{
			ID:                    uuid.New(),
			NIM:                   nim,
			Nama:                  nama,
			Email:                 email,
			Username:              username,
			Password:              string(hashedPassword),
			DisplayPassword:       rawPassword,
			Divisi:                divisi,
			IsPasswordDefault:     true,
			HasSeenPasswordPrompt: false,
		}

		if err := config.DB.Create(&user).Error; err != nil {
			skipped = append(skipped, fmt.Sprintf("%s - db error: %v", nama, err))
			continue
		}

		if user.Email != "" {
			utils.SendAccountCredentialsEmail(utils.MailAccountPayload{
				ToEmail:  user.Email,
				Nama:     user.Nama,
				NIM:      user.NIM,
				Username: user.Username,
				Password: rawPassword,
				Divisi:   user.Divisi,
			})
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
		"message":  fmt.Sprintf("Berhasil mengimpor %d user", len(imported)),
		"imported": imported,
		"skipped":  skipped,
	})
}

// extractSpreadsheetID mengekstrak Google Spreadsheet ID dari berbagai format link
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
			// Cegah redirect ke halaman login google sign-in jika spreadsheet private
			if strings.Contains(req.URL.Host, "accounts.google.com") {
				return fmt.Errorf("spreadsheet bersifat privat. Pastikan akses diset ke 'Siapa saja yang memiliki link'")
			}
			return nil
		},
	}

	resp, err := client.Get(exportURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal mengakses spreadsheet. Pastikan izin sharing 'Siapa saja yang memiliki link dapat melihat': " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal mengunduh spreadsheet dari Google. Pastikan link dapat diakses publik/siapa saja."})
		return
	}

	imported, skipped, err := parseAndSaveUsersFromCSV(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gagal membaca konten spreadsheet: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Berhasil mengimpor %d user dari Google Sheets", len(imported)),
		"imported": imported,
		"skipped":  skipped,
	})
}
