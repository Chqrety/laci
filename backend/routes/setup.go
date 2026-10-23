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
	r.GET("/programs", controllers.GetProkers)

	// Protected routes (Require JWT)
	authGroup := r.Group("/")
	authGroup.Use(middlewares.AuthMiddleware())
	{
		// Profile & password routes (untuk semua user)
		authGroup.GET("/profile", controllers.GetProfile)
		authGroup.POST("/profile/dismiss-prompt", controllers.DismissPasswordPrompt)
		authGroup.PUT("/profile/change-password", controllers.ChangePassword)
		authGroup.PUT("/profile/change-username", controllers.ChangeUsername)

		// User management routes (khusus sekre)
		authGroup.GET("/users", controllers.GetUsers)
		authGroup.POST("/users", controllers.CreateUser)
		authGroup.PUT("/users/:id", controllers.UpdateUser)
		authGroup.DELETE("/users/:id", controllers.DeleteUser)
		authGroup.DELETE("/users/batch", controllers.BatchDeleteUsers)
		authGroup.PUT("/users/:id/reset-prompt", controllers.ResetPasswordPrompt)
		authGroup.POST("/users/import-csv", controllers.ImportUsersCSV)
		authGroup.POST("/users/import-spreadsheet", controllers.ImportUsersSpreadsheet)

		// Letters & Programs
		authGroup.POST("/programs", controllers.CreateProker)
		authGroup.PUT("/programs/:id", controllers.UpdateProker)
		authGroup.DELETE("/programs/:id", controllers.DeleteProker)
		authGroup.GET("/letters", controllers.GetSurat)
		authGroup.POST("/letters", controllers.CreateSurat)
		authGroup.PUT("/letters/:id", controllers.UpdateSurat)
		authGroup.POST("/letters/:id/upload", controllers.UploadSurat)
		authGroup.DELETE("/letters/:id", controllers.DeleteSurat)
	}
}
