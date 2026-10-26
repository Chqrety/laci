package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"laci-backend/models"
)

var DB *gorm.DB

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func InitDB() {
	dbHost := GetEnv("DB_HOST", "127.0.0.1")
	dbPort := GetEnv("DB_PORT", "5432")
	dbUser := GetEnv("DB_USER", "root")
	dbPassword := GetEnv("DB_PASSWORD", "password")
	dbName := GetEnv("DB_NAME", "laci_doscom")
	dbSSLMode := GetEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	RunMigrations(DB)

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Surat{},
		&models.LogTracking{},
		&models.Proker{},
		&models.SuratTemplate{},
		&models.AppSetting{},
	); err != nil {
		log.Fatalf("Failed to auto migrate database schemas: %v", err)
	}
}

// RunMigrations menjalankan migrasi database idempoten v1 ke v2
func RunMigrations(db *gorm.DB) {
	// 1. Tambahkan kolom is_active jika belum ada dan set true untuk seluruh user lama
	_ = db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'is_active') THEN 
				ALTER TABLE users ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
			END IF;
		END $$;
	`).Error

	// 2. Tambahkan kolom token aktivasi dan reset password
	_ = db.Exec(`
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'activation_token') THEN 
				ALTER TABLE users ADD COLUMN activation_token VARCHAR(255);
				CREATE INDEX IF NOT EXISTS idx_users_activation_token ON users (activation_token);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'activation_token_expires_at') THEN 
				ALTER TABLE users ADD COLUMN activation_token_expires_at TIMESTAMP;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'reset_password_token') THEN 
				ALTER TABLE users ADD COLUMN reset_password_token VARCHAR(255);
				CREATE INDEX IF NOT EXISTS idx_users_reset_password_token ON users (reset_password_token);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'reset_password_expires_at') THEN 
				ALTER TABLE users ADD COLUMN reset_password_expires_at TIMESTAMP;
			END IF;
		END $$;
	`).Error

	// 3. Drop kolom bawaan lama yang rentan (display_password, is_password_default, has_seen_password_prompt)
	_ = db.Exec(`
		ALTER TABLE users 
		DROP COLUMN IF EXISTS display_password,
		DROP COLUMN IF EXISTS is_password_default,
		DROP COLUMN IF EXISTS has_seen_password_prompt;
	`).Error
}
