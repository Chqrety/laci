package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"laci-backend/config"
	"laci-backend/routes"
	"laci-backend/seeders"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Catatan: File .env tidak ditemukan, menggunakan nilai environment OS bawaan.")
	}

	config.InitDB()
	seeders.RunAllSeeders(config.DB)

	r := gin.Default()
	routes.SetupRoutes(r)

	port := config.GetEnv("PORT", "8080")
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to run: %v", err)
	}
}
