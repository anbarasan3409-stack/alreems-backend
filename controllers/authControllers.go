package controllers

import (
	"context"

	"backend/database"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

func GoogleCallbackController(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" {
			c.JSON(400, gin.H{"error": "Authorization code not provided"})
			return
		}

		if state != "signup" && state != "login" && state != "admin_login" && state != "teacher_login" && state != "teacher_signup" {
			c.JSON(400, gin.H{"error": "Invalid state parameter"})
			return
		}

		token, err := utils.GoogleOauthConfig.Exchange(context.Background(), code)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to exchange token", "details": err.Error()})
			return
		}

		userInfo, err := utils.GetGoogleUserInfo(token.AccessToken)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get user info", "details": err.Error()})
			return
		}

		email, ok := userInfo["email"].(string)
		if !ok {
			c.JSON(500, gin.H{"error": "Email not found in user info"})
			return
		}

		if state == "admin_login" {
			HandleAdminLogin(c, handler, email, token)
			return
		}

		if state == "teacher_login" {
			HandleTeacherLogin(c, handler, email, token)
			return
		}

		if state == "teacher_signup" {
			name, _ := userInfo["name"].(string)
			HandleTeacherSignup(c, handler, email, name, token)
			return
		}

		if state == "signup" {
			name, _ := userInfo["name"].(string)
			HandleStudentSignup(c, handler, email, name, token)
			return
		} else if state == "login" {
			HandleStudentLogin(c, handler, email, token)
		}
	}
}
