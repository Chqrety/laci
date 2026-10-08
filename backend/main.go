package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Password string    `gorm:"type:varchar(255);not null" json:"-"`
	Role     string    `gorm:"type:varchar(50);not null" json:"role"` // 'sekre' or 'humas'
}

func (User) TableName() string {
	return "users"
}

type Surat struct {
	ID            uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NomorSurat    string        `gorm:"type:varchar(255);not null" json:"nomor_surat"`
	Perihal       string        `gorm:"type:varchar(255);not null" json:"perihal"`
	StatusSaatIni string        `gorm:"type:varchar(100);not null" json:"status_saat_ini"`
	PICNama       string        `gorm:"type:varchar(100);not null" json:"pic_nama"`
	ArsipURL      *string       `gorm:"type:text" json:"arsip_url"`
	FileArsip     string        `gorm:"type:text" json:"file_arsip"`
	Catatan       string        `gorm:"type:text" json:"catatan"`
	Logs          []LogTracking `gorm:"foreignKey:SuratID" json:"logs,omitempty"`
}

func (Surat) TableName() string {
	return "surat"
}

type LogTracking struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	SuratID     uuid.UUID `gorm:"type:uuid;not null;index" json:"surat_id"`
	StatusBaru  string    `gorm:"type:varchar(100);not null" json:"status_baru"`
	CatatanLog  string    `gorm:"type:text" json:"catatan_log"`
	WaktuUpdate time.Time `gorm:"type:timestamp;not null" json:"waktu_update"`
}

func (LogTracking) TableName() string {
	return "log_tracking"
}

type Proker struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	NamaProker string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"nama_proker"`
	Deskripsi  string    `gorm:"type:text" json:"deskripsi"`
}

func (Proker) TableName() string {
	return "proker"
}

var (
	db        *gorm.DB
	jwtSecret = []byte(getEnv("JWT_SECRET", "supersecretkey123"))
)

func initDB() {
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "5433")
	dbUser := getEnv("DB_USER", "admin")
	dbPassword := getEnv("DB_PASSWORD", "password123")
	dbName := getEnv("DB_NAME", "laci_doscom")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode,
	)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&User{}, &Surat{}, &LogTracking{}, &Proker{}); err != nil {
		log.Fatalf("Failed to auto migrate database schemas: %v", err)
	}

	// Seed default admin user ('sekre') if not exists
	var sekreCount int64
	db.Model(&User{}).Where("username = ?", "sekre").Count(&sekreCount)
	if sekreCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default password: %v", err)
		}
		defaultUser := User{
			ID:       uuid.New(),
			Username: "sekre",
			Password: string(hashedPassword),
			Role:     "sekre",
		}
		if err := db.Create(&defaultUser).Error; err != nil {
			log.Printf("Warning: failed to seed default admin user: %v", err)
		} else {
			log.Println("Default user 'sekre' created successfully.")
		}
	}

	// Seed default user ('humas') if not exists
	var humasCount int64
	db.Model(&User{}).Where("username = ?", "humas").Count(&humasCount)
	if humasCount == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash default password: %v", err)
		}
		humasUser := User{
			ID:       uuid.New(),
			Username: "humas",
			Password: string(hashedPassword),
			Role:     "humas",
		}
		if err := db.Create(&humasUser).Error; err != nil {
			log.Printf("Warning: failed to seed default user 'humas': %v", err)
		} else {
			log.Println("Default user 'humas' created successfully.")
		}
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer <token>"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Set("userID", claims["sub"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

func main() {
	initDB()

	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	r.Static("/uploads", "./uploads")

	// 1. POST /login (Accepts username & password, returns JWT token and role)
	r.POST("/login", func(c *gin.Context) {
		var input struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
			return
		}

		var user User
		if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		claims := jwt.MapClaims{
			"sub":      user.ID.String(),
			"username": user.Username,
			"role":     user.Role,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"role":  user.Role,
		})
	})

	// GET /proker (Mengembalikan list semua proker)
	r.GET("/proker", func(c *gin.Context) {
		var listProker []Proker
		if err := db.Order("nama_proker asc").Find(&listProker).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, listProker)
	})

	// Protected Routes (Require JWT)
	authGroup := r.Group("/")
	authGroup.Use(authMiddleware())

	// POST /proker (Requires JWT. Menerima JSON {nama_proker, deskripsi}, simpan ke DB)
	authGroup.POST("/proker", func(c *gin.Context) {
		var input struct {
			NamaProker string `json:"nama_proker" binding:"required"`
			Deskripsi  string `json:"deskripsi"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker is required"})
			return
		}

		trimmed := strings.TrimSpace(input.NamaProker)
		if trimmed == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nama_proker cannot be empty"})
			return
		}

		proker := Proker{
			ID:         uuid.New(),
			NamaProker: trimmed,
			Deskripsi:  strings.TrimSpace(input.Deskripsi),
		}

		if err := db.Create(&proker).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Proker already exists or failed to save"})
			return
		}

		c.JSON(http.StatusCreated, proker)
	})

	// 2. GET /surat (Requires JWT. Returns list of surat)
	authGroup.GET("/surat", func(c *gin.Context) {
		var listSurat []Surat
		if err := db.Preload("Logs").Find(&listSurat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, listSurat)
	})

	// 3. POST /surat (Requires JWT)
	authGroup.POST("/surat", func(c *gin.Context) {
		var input struct {
			NomorSurat    string `json:"nomor_surat" binding:"required"`
			Perihal       string `json:"perihal" binding:"required"`
			StatusSaatIni string `json:"status_saat_ini"`
			PICNama       string `json:"pic_nama" binding:"required"`
			Catatan       string `json:"catatan"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		roleVal, _ := c.Get("role")
		role, _ := roleVal.(string)

		status := input.StatusSaatIni
		if role == "humas" || strings.TrimSpace(status) == "" {
			status = "Standby"
		}

		suratID := uuid.New()
		surat := Surat{
			ID:            suratID,
			NomorSurat:    input.NomorSurat,
			Perihal:       input.Perihal,
			StatusSaatIni: status,
			PICNama:       input.PICNama,
			Catatan:       input.Catatan,
		}

		logTracking := LogTracking{
			ID:          uuid.New(),
			SuratID:     suratID,
			StatusBaru:  status,
			CatatanLog:  input.Catatan,
			WaktuUpdate: time.Now(),
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&surat).Error; err != nil {
				return err
			}
			if err := tx.Create(&logTracking).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		surat.Logs = []LogTracking{logTracking}
		c.JSON(http.StatusCreated, surat)
	})

	// 4. PUT /surat/:id/status & PUT /surat/:id (Requires JWT. Allows Sekre & Humas to update catatan, status preserved or updated)
	updateSuratHandler := func(c *gin.Context) {
		idParam := c.Param("id")
		suratUUID, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
			return
		}

		var surat Surat
		if err := db.First(&surat, "id = ?", suratUUID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
			return
		}

		var input struct {
			StatusSaatIni string  `json:"status_saat_ini"`
			StatusBaru    string  `json:"status_baru"`
			Status        string  `json:"status"`
			Catatan       *string `json:"catatan"`
			CatatanLog    *string `json:"catatan_log"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newStatus := input.StatusSaatIni
		if newStatus == "" {
			newStatus = input.StatusBaru
		}
		if newStatus == "" {
			newStatus = input.Status
		}

		// If no status is supplied, retain the current status
		if newStatus == "" {
			newStatus = surat.StatusSaatIni
		}

		roleVal, _ := c.Get("role")
		role, _ := roleVal.(string)

		// Humas is permitted to edit notes, while status transitions are preserved
		if role == "humas" && newStatus != surat.StatusSaatIni {
			newStatus = surat.StatusSaatIni
		}

		// Determine log note and whether to update surat note
		note := surat.Catatan
		if input.Catatan != nil {
			note = *input.Catatan
		}
		catatanLog := note
		if input.CatatanLog != nil {
			catatanLog = *input.CatatanLog
		}

		logTracking := LogTracking{
			ID:          uuid.New(),
			SuratID:     surat.ID,
			StatusBaru:  newStatus,
			CatatanLog:  catatanLog,
			WaktuUpdate: time.Now(),
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			updates := map[string]interface{}{
				"status_saat_ini": newStatus,
			}
			if input.Catatan != nil {
				updates["catatan"] = *input.Catatan
			}
			if err := tx.Model(&surat).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Create(&logTracking).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		surat.StatusSaatIni = newStatus
		if input.Catatan != nil {
			surat.Catatan = *input.Catatan
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Surat updated successfully",
			"surat":   surat,
			"log":     logTracking,
		})
	}

	authGroup.PUT("/surat/:id/status", updateSuratHandler)
	authGroup.PUT("/surat/:id", updateSuratHandler)

	// 5. POST /surat/:id/upload (Requires JWT. Upload file via multipart/form-data key 'file', save to ./uploads)
	authGroup.POST("/surat/:id/upload", func(c *gin.Context) {
		idParam := c.Param("id")
		suratUUID, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
			return
		}

		var surat Surat
		if err := db.First(&surat, "id = ?", suratUUID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Key 'file' is required."})
			return
		}

		if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
			return
		}

		ext := filepath.Ext(file.Filename)
		newFileName := fmt.Sprintf("%s_%s%s", suratUUID.String(), uuid.New().String()[:8], ext)
		dst := filepath.Join("uploads", newFileName)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		fileArsipURL := "/uploads/" + newFileName
		updates := map[string]interface{}{
			"file_arsip": fileArsipURL,
			"arsip_url":  fileArsipURL,
		}
		if err := db.Model(&surat).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		surat.FileArsip = fileArsipURL
		surat.ArsipURL = &fileArsipURL
		c.JSON(http.StatusOK, gin.H{
			"message":    "File arsip berhasil diunggah",
			"file_arsip": fileArsipURL,
			"arsip_url":  fileArsipURL,
			"surat":      surat,
		})
	})

	// 6. POST /surat/:id/arsip (Requires JWT. Backward compatible upload image/blob)
	authGroup.POST("/surat/:id/arsip", func(c *gin.Context) {
		idParam := c.Param("id")
		suratUUID, err := uuid.Parse(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
			return
		}

		var surat Surat
		if err := db.First(&surat, "id = ?", suratUUID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Surat not found"})
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			file, err = c.FormFile("arsip")
		}
		if err != nil {
			file, err = c.FormFile("image")
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Key 'file', 'arsip', or 'image' is required."})
			return
		}

		ext := filepath.Ext(file.Filename)
		newFileName := fmt.Sprintf("%s_%s%s", suratUUID.String(), uuid.New().String()[:8], ext)
		dst := filepath.Join("uploads", newFileName)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		arsipURL := "/uploads/" + newFileName
		updates := map[string]interface{}{
			"file_arsip": arsipURL,
			"arsip_url":  arsipURL,
		}
		if err := db.Model(&surat).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		surat.FileArsip = arsipURL
		surat.ArsipURL = &arsipURL
		c.JSON(http.StatusOK, gin.H{
			"message":    "Arsip uploaded successfully",
			"file_arsip": arsipURL,
			"arsip_url":  arsipURL,
			"surat":      surat,
		})
	})

	port := getEnv("PORT", "8080")
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
