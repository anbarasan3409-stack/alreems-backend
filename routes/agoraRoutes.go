package routes

import (
	"backend/controllers"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

// SetupAgoraRoutes sets up Agora-related routes
func SetupAgoraRoutes(router *gin.Engine) {
	agora := router.Group("/api/agora")
	{
		// Public routes
		agora.GET("/config", controllers.GetAgoraConfig)

		// Protected routes for students
		studentAgora := agora.Group("/student")
		studentAgora.Use(middlewares.StudentAuthMiddleware())
		studentAgora.POST("/token", controllers.GenerateAgoraToken)

		// Protected routes for teachers
		teacherAgora := agora.Group("/teacher")
		teacherAgora.Use(middlewares.TeacherAuthMiddleware())
		teacherAgora.POST("/token", controllers.GenerateAgoraToken)
	}
}
