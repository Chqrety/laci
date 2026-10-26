package routes

import (
	"github.com/gin-gonic/gin"
	"laci-backend/config"
	"laci-backend/handlers"
	"laci-backend/middlewares"
	"laci-backend/repositories"
	"laci-backend/services"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middlewares.CORSMiddleware())
	r.Static("/uploads", "./uploads")
	r.Static("/api/uploads", "./uploads")

	// Repositories
	userRepo := repositories.NewUserRepository(config.DB)
	suratRepo := repositories.NewSuratRepository(config.DB)
	prokerRepo := repositories.NewProkerRepository(config.DB)
	templateRepo := repositories.NewTemplateRepository(config.DB)

	// Services
	authService := services.NewAuthService(userRepo)
	userService := services.NewUserService(userRepo)
	suratService := services.NewSuratService(suratRepo)
	prokerService := services.NewProkerService(prokerRepo)
	templateService := services.NewTemplateService(templateRepo, suratRepo, prokerRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	suratHandler := handlers.NewSuratHandler(suratService)
	prokerHandler := handlers.NewProkerHandler(prokerService)
	templateHandler := handlers.NewTemplateHandler(templateService)

	// Public routes
	r.POST("/login", authHandler.Login)
	r.GET("/programs", prokerHandler.GetProkers)
	r.GET("/auth/verify-token", authHandler.VerifyToken)
	r.POST("/auth/activate", authHandler.ActivateAccount)
	r.POST("/auth/forgot-password", authHandler.ForgotPassword)
	r.POST("/auth/reset-password", authHandler.ResetPassword)

	// Protected routes (Require JWT)
	authGroup := r.Group("/")
	authGroup.Use(middlewares.AuthMiddleware())
	{
		// Profile & password routes
		authGroup.GET("/profile", authHandler.GetProfile)
		authGroup.PUT("/profile/change-password", authHandler.ChangePassword)
		authGroup.PUT("/profile/change-username", authHandler.ChangeUsername)

		// User management routes (Sekre & Ketum)
		authGroup.GET("/users", userHandler.GetUsers)
		authGroup.POST("/users", userHandler.CreateUser)
		authGroup.PUT("/users/:id", userHandler.UpdateUser)
		authGroup.DELETE("/users/:id", userHandler.DeleteUser)
		authGroup.DELETE("/users/batch", userHandler.BatchDeleteUsers)
		authGroup.POST("/users/import-csv", userHandler.ImportUsersCSV)
		authGroup.POST("/users/import-spreadsheet", userHandler.ImportUsersSpreadsheet)
		authGroup.POST("/users/transfer-ketum", userHandler.TransferKetum)

		// Letters & Programs
		authGroup.POST("/programs", prokerHandler.CreateProker)
		authGroup.PUT("/programs/:id", prokerHandler.UpdateProker)
		authGroup.DELETE("/programs/:id", prokerHandler.DeleteProker)
		authGroup.GET("/letters", suratHandler.GetSurat)
		authGroup.POST("/letters", suratHandler.CreateSurat)
		authGroup.PUT("/letters/:id", suratHandler.UpdateSurat)
		authGroup.POST("/letters/:id/upload", suratHandler.UploadSurat)
		authGroup.DELETE("/letters/:id", suratHandler.DeleteSurat)

		// Letter Generator & Templates (Google Drive / Apps Script)
		authGroup.GET("/letters/templates", templateHandler.GetTemplates)
		authGroup.POST("/letters/templates/sync", templateHandler.SyncTemplates)
		authGroup.POST("/letters/generate", templateHandler.GenerateSurat)
		authGroup.GET("/settings", templateHandler.GetAppSettings)
		authGroup.PUT("/settings", templateHandler.UpdateAppSettings)
	}
}
