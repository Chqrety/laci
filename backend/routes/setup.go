package routes

import (
	"github.com/gin-gonic/gin"
	"laci-backend/controllers"
	"laci-backend/middlewares"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middlewares.CORSMiddleware())
	r.Static("/uploads", "./uploads")

	// Public routes
	r.POST("/login", controllers.Login)
	r.GET("/proker", controllers.GetProkers)

	// Protected routes (Require JWT)
	authGroup := r.Group("/")
	authGroup.Use(middlewares.AuthMiddleware())
	{
		// Profile & password routes (untuk semua user)
		authGroup.GET("/profile", controllers.GetProfile)
		authGroup.POST("/profile/dismiss-prompt", controllers.DismissPasswordPrompt)
		authGroup.PUT("/profile/change-password", controllers.ChangePassword)

		// User management routes (khusus sekre)
		authGroup.GET("/users", controllers.GetUsers)
		authGroup.POST("/users", controllers.CreateUser)
		authGroup.PUT("/users/:id", controllers.UpdateUser)
		authGroup.DELETE("/users/:id", controllers.DeleteUser)
		authGroup.DELETE("/users/batch", controllers.BatchDeleteUsers)
		authGroup.PUT("/users/:id/reset-prompt", controllers.ResetPasswordPrompt)
		authGroup.POST("/users/import-csv", controllers.ImportUsersCSV)
		authGroup.POST("/users/import-spreadsheet", controllers.ImportUsersSpreadsheet)

		// Surat & Proker
		authGroup.POST("/proker", controllers.CreateProker)
		authGroup.GET("/surat", controllers.GetSurat)
		authGroup.POST("/surat", controllers.CreateSurat)
		authGroup.PUT("/surat/:id/status", controllers.UpdateSurat)
		authGroup.PUT("/surat/:id", controllers.UpdateSurat)
		authGroup.POST("/surat/:id/upload", controllers.UploadSurat)
		authGroup.POST("/surat/:id/arsip", controllers.UploadArsip)
		authGroup.DELETE("/surat/:id", controllers.DeleteSurat)
	}
}
