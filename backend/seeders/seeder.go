package seeders

import (
	"log"

	"laci-backend/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User = models.User
type Proker = models.Proker

// SeedCore menjalankan seeder kebutuhan dasar (Users & Proker) untuk production & inisialisasi sistem
func SeedCore(db *gorm.DB) {
	SeedUsers(db)
	SeedProker(db)
}

func RunAllSeeders(db *gorm.DB) {
	SeedCore(db)
}

func SeedUsers(db *gorm.DB) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash default password: %v", err)
	}

	defaultUsers := []User{
		{
			ID:                    uuid.New(),
			NIM:                   "A11.2024.00001",
			Nama:                  "Sekretaris DOSCOM",
			Email:                 "sekre@doscom.org",
			Username:              "sekre",
			Password:              string(hashedPassword),
			DisplayPassword:       "123",
			Divisi:                "sekre",
			IsPasswordDefault:     true,
			HasSeenPasswordPrompt: true,
		},
		{
			ID:                    uuid.New(),
			NIM:                   "A11.2024.00002",
			Nama:                  "Humas DOSCOM",
			Email:                 "humas@doscom.org",
			Username:              "humas",
			Password:              string(hashedPassword),
			DisplayPassword:       "123",
			Divisi:                "humas",
			IsPasswordDefault:     true,
			HasSeenPasswordPrompt: false,
		},
	}

	for _, u := range defaultUsers {
		var existing User
		if err := db.Where("username = ?", u.Username).First(&existing).Error; err != nil {
			if err := db.Create(&u).Error; err != nil {
				log.Printf("Warning: failed to seed '%s': %v", u.Username, err)
			}
		} else {
			db.Model(&existing).Updates(map[string]interface{}{
				"password":         u.Password,
				"display_password": u.DisplayPassword,
			})
		}
	}
}

func SeedProker(db *gorm.DB) {
	defaultProkers := []Proker{
		{NamaProker: "MAKRAB", Deskripsi: "Malam Keakraban"},
		{NamaProker: "OPREC", Deskripsi: "Open Recruitment"},
		{NamaProker: "OOTS", Deskripsi: "Opensource On The School"},
		{NamaProker: "DST", Deskripsi: "Doscom Sharing Time"},
		{NamaProker: "DU", Deskripsi: "Doscom University"},
		{NamaProker: "RP", Deskripsi: "Release Party"},
	}

	for _, p := range defaultProkers {
		var proker Proker
		p.ID = uuid.New()
		if err := db.Where(Proker{NamaProker: p.NamaProker}).FirstOrCreate(&proker, p).Error; err != nil {
			log.Printf("Warning: failed to seed proker %s: %v", p.NamaProker, err)
		}
	}
}
