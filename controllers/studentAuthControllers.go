package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"backend/database"
	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func HandleStudentSignup(c *gin.Context, handler database.Handler, email string, name string, token *oauth2.Token) {
	var existingStudent models.Student
	if err := handler.DB.Where("email = ?", email).First(&existingStudent).Error; err == nil {
		c.JSON(409, gin.H{"error": "Student already exists, please login instead"})
		return
	}

	newStudent := models.Student{
		Email:             email,
		Name:              name,
		PhoneNumber:       "",
		Address:           "",
		ProfilePicture:    "",
		GoogleAccessToken: token.AccessToken,
		RefreshToken:      "",
	}

	if err := handler.DB.Create(&newStudent).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create student", "details": err.Error()})
		return
	}
	refreshToken, err := utils.GenerateRefreshToken(newStudent.ID, newStudent.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token", "details": err.Error()})
		return
	}
	newStudent.RefreshToken = refreshToken
	if err := handler.DB.Save(&newStudent).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save refresh token", "details": err.Error()})
		return
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Default for development
	}

	// Get cookie settings from environment
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

	c.SetSameSite(cookieSameSite)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", cookieDomain, cookieSecure, true) // 7 days, httpOnly
	c.Redirect(http.StatusFound, frontendURL+"/auth/callback?success=true&type=signup")
}
func HandleStudentLogin(c *gin.Context, handler database.Handler, email string, token *oauth2.Token) {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000" // Default for development
	}

	var student models.Student
	if err := handler.DB.Where("email = ?", email).First(&student).Error; err != nil {
		c.Redirect(http.StatusFound, frontendURL+"/auth/callback?error=Account+not+found.+Please+sign+up+first.")
		return
	}

	now := time.Now()
	student.LastLogin = &now
	refreshToken, err := utils.GenerateRefreshToken(student.ID, student.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token", "details": err.Error()})
		return
	}
	student.RefreshToken = refreshToken
	if err := handler.DB.Save(&student).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save refresh token", "details": err.Error()})
		return
	}



	// Get cookie settings from environment
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

	c.SetSameSite(cookieSameSite)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", cookieDomain, cookieSecure, true) // 7 days, httpOnly

	c.Redirect(http.StatusFound, frontendURL+"/auth/callback?success=true&type=login")
}

// Get student profile (authenticated endpoint)
func GetStudentProfile(c *gin.Context, handler database.Handler) {
	studentID, exists := c.Get("studentID")
	if !exists {
		c.JSON(401, gin.H{"error": "Student ID not found in context"})
		return
	}

	var student models.Student
	if err := handler.DB.Where("id = ?", studentID).First(&student).Error; err != nil {
		c.JSON(404, gin.H{"error": "Student not found"})
		return
	}

	// Return student data without sensitive information
	c.JSON(200, gin.H{
		"ID":             student.ID,
		"name":           student.Name,
		"email":          student.Email,
		"phoneNumber":    student.PhoneNumber,
		"address":        student.Address,
		"profilePicture": student.ProfilePicture,
		"lastLogin":      student.LastLogin,
	})
}

// Refresh token controller for students
func StudentRefreshTokenController(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var refreshToken string

		// First try to get from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// If no Authorization header, try to get from cookie
			cookie, err := ctx.Cookie("refresh_token")
			if err != nil {
				ctx.JSON(401, gin.H{"error": "No refresh token provided"})
				return
			}
			refreshToken = cookie
		}

		// Validate the refresh token
		claims, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Get student from database
		var student models.Student
		if err := handler.DB.Where("id = ?", claims.StudentID).First(&student).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Generate new refresh token
		newRefreshToken, err := utils.GenerateRefreshToken(student.ID, student.Email)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to generate new refresh token"})
			return
		}

		// Update student's refresh token in database
		student.RefreshToken = newRefreshToken
		if err := handler.DB.Save(&student).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update refresh token"})
			return
		}

		// Set new cookie with environment-based configuration
		cookieDomain := os.Getenv("COOKIE_DOMAIN")
		cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
		cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

		ctx.SetSameSite(cookieSameSite)
		ctx.SetCookie("refresh_token", newRefreshToken, 7*24*3600, "/", cookieDomain, cookieSecure, true)

		ctx.JSON(200, gin.H{
			"message": "Token refreshed successfully",
			"student": gin.H{
				"ID":             student.ID,
				"name":           student.Name,
				"email":          student.Email,
				"phoneNumber":    student.PhoneNumber,
				"address":        student.Address,
				"profilePicture": student.ProfilePicture,
				"lastLogin":      student.LastLogin,
			},
		})
	}
}

// Logout controller for students
func StudentLogoutController(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var refreshToken string

		// First try to get from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// If no Authorization header, try to get from cookie
			cookie, err := ctx.Cookie("refresh_token")
			if err != nil {
				ctx.JSON(401, gin.H{"error": "No refresh token provided"})
				return
			}
			refreshToken = cookie
		}

		// Get cookie settings from environment
		cookieDomain := os.Getenv("COOKIE_DOMAIN")
		cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
		cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

		// Validate the refresh token to get student ID
		claims, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil {
			// Even if token is invalid, we still want to clear the cookie
			ctx.SetSameSite(cookieSameSite)
			ctx.SetCookie("refresh_token", "", -1, "/", cookieDomain, cookieSecure, true)
			ctx.JSON(200, gin.H{"message": "Logged out successfully"})
			return
		}

		// Clear the refresh token in database
		var student models.Student
		if err := handler.DB.Where("id = ?", claims.StudentID).First(&student).Error; err == nil {
			student.RefreshToken = ""
			handler.DB.Save(&student)
		}

		// Clear the cookie
		ctx.SetSameSite(cookieSameSite)
		ctx.SetCookie("refresh_token", "", -1, "/", cookieDomain, cookieSecure, true)

		ctx.JSON(200, gin.H{"message": "Logged out successfully"})
	}
}
