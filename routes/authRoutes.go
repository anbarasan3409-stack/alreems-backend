package routes

import (
	"backend/controllers"
	"backend/database"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine, handler database.Handler) {
	googleAuthGroup := incomingRoutes.Group("/api/auth/google")

	googleAuthGroup.GET("/signup", func(ctx *gin.Context) {
		ctx.JSON(403, gin.H{"error": "Student signup is disabled. Please contact the administrator."})
	})

	googleAuthGroup.GET("/login", func(ctx *gin.Context) {
		url := utils.GoogleOauthConfig.AuthCodeURL("login")
		ctx.Redirect(302, url)
	})
	googleAuthGroup.GET("/callback", controllers.GoogleCallbackController(handler))

	// Teacher OAuth routes
	teacherAuthGroup := incomingRoutes.Group("/api/auth/teacher/google")
	teacherAuthGroup.GET("/signup", func(ctx *gin.Context) {
		url := utils.GoogleOauthConfig.AuthCodeURL("teacher_signup")
		ctx.Redirect(302, url)
	})
	teacherAuthGroup.GET("/login", func(ctx *gin.Context) {
		url := utils.GoogleOauthConfig.AuthCodeURL("teacher_login")
		ctx.Redirect(302, url)
	})

	adminAuthGroup := incomingRoutes.Group("/api/auth/admin/google")
	adminAuthGroup.GET("/login", func(ctx *gin.Context) {
		url := utils.GoogleOauthConfig.AuthCodeURL("admin_login")
		ctx.Redirect(302, url)
	})
}
