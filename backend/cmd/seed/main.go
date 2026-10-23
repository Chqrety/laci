package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
	"laci-backend/config"
	"laci-backend/seeders"
)

func main() {
	includeDummy := flag.Bool("dummy", false, "Sertakan data dummy surat & tracking untuk keperluan dev")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		if err := godotenv.Load("../../.env"); err != nil {
			log.Println("Catatan: File .env tidak ditemukan, menggunakan nilai environment OS bawaan.")
		}
	}

	config.InitDB()

	log.Println("Menjalankan seeder inti (Users & Programs/Proker)...")
	seeders.SeedCore(config.DB)

	if *includeDummy {
		log.Println("Opsi -dummy aktif: Menjalankan seeder dummy surat & riwayat tracking...")
		seeders.SeedDummySurat(config.DB)
	}

	log.Println("Database seeder selesai dijalankan dengan sukses.")
}
