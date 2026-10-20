package main

import (
	"log"

	"github.com/joho/godotenv"
	"laci-backend/config"
	"laci-backend/seeders"
)

func main() {
	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("Catatan: File .env tidak ditemukan, menggunakan nilai environment OS bawaan.")
		}
	}

	config.InitDB()

	log.Println("Menjalankan database seeders...")
	seeders.RunAllSeeders(config.DB)
	log.Println("Database seeders selesai dijalankan dengan sukses.")
}
