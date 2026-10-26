package routes

import (
	"github.com/gin-gonic/gin"
	"laci-backend/controllers"
	"laci-backend/middlewares"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(middlewares.CORSMiddleware())
	r.Static("/uploads", "./uploads")
	r.Static("/api/uploads", "./uploads")

	// Public routes
	r.POST("/login", controllers.Login)
	r.GET("/programs", controllers.GetProkers)
	r.GET("/auth/verify-token", controllers.VerifyToken)
	r.POST("/auth/activate", controllers.ActivateAccount)
	r.POST("/auth/forgot-password", controllers.ForgotPassword)
	r.POST("/auth/reset-password", controllers.ResetPassword)

	// Protected routes (Require JWT)
	authGroup := r.Group("/")
	authGroup.Use(middlewares.AuthMiddleware())
	{
		// Profile & password routes
		authGroup.GET("/profile", controllers.GetProfile)
		authGroup.PUT("/profile/change-password", controllers.ChangePassword)
		authGroup.PUT("/profile/change-username", controllers.ChangeUsername)

		// User management routes (Sekre & Ketum)
		authGroup.GET("/users", controllers.GetUsers)
		authGroup.POST("/users", controllers.CreateUser)
		authGroup.PUT("/users/:id", controllers.UpdateUser)
		authGroup.DELETE("/users/:id", controllers.DeleteUser)
		authGroup.DELETE("/users/batch", controllers.BatchDeleteUsers)
		authGroup.POST("/users/import-csv", controllers.ImportUsersCSV)
		authGroup.POST("/users/import-spreadsheet", controllers.ImportUsersSpreadsheet)
		authGroup.POST("/users/transfer-ketum", controllers.TransferKetum)

		// Letters & Programs
		authGroup.POST("/programs", controllers.CreateProker)
		authGroup.PUT("/programs/:id", controllers.UpdateProker)
		authGroup.DELETE("/programs/:id", controllers.DeleteProker)
		authGroup.GET("/letters", controllers.GetSurat)
		authGroup.POST("/letters", controllers.CreateSurat)
		authGroup.PUT("/letters/:id", controllers.UpdateSurat)
		authGroup.POST("/letters/:id/upload", controllers.UploadSurat)
		authGroup.DELETE("/letters/:id", controllers.DeleteSurat)

		// Letter Generator & Templates (Google Drive / Apps Script)
		authGroup.GET("/letters/templates", controllers.GetTemplates)
		authGroup.POST("/letters/templates/sync", controllers.SyncTemplates)
		authGroup.POST("/letters/generate", controllers.GenerateSurat)
		authGroup.GET("/settings", controllers.GetAppSettings)
		authGroup.PUT("/settings", controllers.UpdateAppSettings)
	}
}
