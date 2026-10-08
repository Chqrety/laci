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
