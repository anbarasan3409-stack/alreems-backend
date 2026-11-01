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

func HandleTeacherSignup(c *gin.Context, handler database.Handler, email string, name string, token *oauth2.Token) {
	var existingTeacher models.Teacher
	if err := handler.DB.Where("email = ?", email).First(&existingTeacher).Error; err == nil {
		c.JSON(409, gin.H{"error": "Teacher already exists, please login instead"})
		return
	}

	newTeacher := models.Teacher{
		Email:             email,
		Name:              name,
		PhoneNumber:       "",
		Address:           "",
		ProfilePicture:    "",
		Department:        "",
		Designation:       "",
		GoogleAccessToken: token.AccessToken,
		RefreshToken:      "",
	}

	if err := handler.DB.Create(&newTeacher).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to create teacher", "details": err.Error()})
		return
	}

	refreshToken, err := utils.GenerateTeacherRefreshToken(newTeacher.ID, newTeacher.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token", "details": err.Error()})
		return
	}

	newTeacher.RefreshToken = refreshToken
	if err := handler.DB.Save(&newTeacher).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save refresh token", "details": err.Error()})
		return
	}

	teachersURL := os.Getenv("TEACHERS_URL")

	// Get cookie settings from environment
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

	c.SetSameSite(cookieSameSite)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", cookieDomain, cookieSecure, true) // 7 days, httpOnly
	c.Redirect(http.StatusFound, teachersURL+"/dashboard")
}

func HandleTeacherLogin(c *gin.Context, handler database.Handler, email string, token *oauth2.Token) {
	var teacher models.Teacher
	if err := handler.DB.Where("email = ?", email).First(&teacher).Error; err != nil {
		c.JSON(404, gin.H{"error": "Teacher not found, please signup first"})
		return
	}

	now := time.Now()
	teacher.LastLogin = &now
	refreshToken, err := utils.GenerateTeacherRefreshToken(teacher.ID, teacher.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token", "details": err.Error()})
		return
	}
	teacher.RefreshToken = refreshToken
	if err := handler.DB.Save(&teacher).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save refresh token", "details": err.Error()})
		return
	}
	teachersURL := os.Getenv("TEACHERS_URL")

	// Get cookie settings from environment
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

	c.SetSameSite(cookieSameSite)
	c.SetCookie("refresh_token", refreshToken, 7*24*3600, "/", cookieDomain, cookieSecure, true) // 7 days, httpOnly
	c.Redirect(http.StatusFound, teachersURL+"/auth/callback")
}

// Get teacher profile (authenticated endpoint)
func GetTeacherProfile(c *gin.Context, handler database.Handler) {
	teacherID, exists := c.Get("teacherID")
	if !exists {
		c.JSON(401, gin.H{"error": "Teacher ID not found in context"})
		return
	}

	var teacher models.Teacher
	if err := handler.DB.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
		c.JSON(404, gin.H{"error": "Teacher not found"})
		return
	}

	// Return teacher data without sensitive information
	c.JSON(200, gin.H{
		"ID":             teacher.ID,
		"name":           teacher.Name,
		"email":          teacher.Email,
		"phoneNumber":    teacher.PhoneNumber,
		"address":        teacher.Address,
		"profilePicture": teacher.ProfilePicture,
		"department":     teacher.Department,
		"designation":    teacher.Designation,
		"lastLogin":      teacher.LastLogin,
	})
}

// Refresh token controller for teachers
func TeacherRefreshTokenController(handler database.Handler) gin.HandlerFunc {
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
		teacherID, err := utils.ValidateTeacherRefreshToken(refreshToken)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Get teacher from database
		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Teacher not found"})
			return
		}

		// Generate new refresh token
		newRefreshToken, err := utils.GenerateTeacherRefreshToken(teacher.ID, teacher.Email)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to generate new refresh token"})
			return
		}

		// Update teacher's refresh token in database
		teacher.RefreshToken = newRefreshToken
		if err := handler.DB.Save(&teacher).Error; err != nil {
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
			"teacher": gin.H{
				"ID":             teacher.ID,
				"name":           teacher.Name,
				"email":          teacher.Email,
				"phoneNumber":    teacher.PhoneNumber,
				"address":        teacher.Address,
				"profilePicture": teacher.ProfilePicture,
				"department":     teacher.Department,
				"designation":    teacher.Designation,
				"lastLogin":      teacher.LastLogin,
			},
		})
	}
}

// Logout controller for teachers
func TeacherLogoutController(handler database.Handler) gin.HandlerFunc {
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

		// Validate the refresh token to get teacher ID
		teacherID, err := utils.ValidateTeacherRefreshToken(refreshToken)
		if err != nil {
			// Even if token is invalid, we still want to clear the cookie
			ctx.SetSameSite(cookieSameSite)
			ctx.SetCookie("refresh_token", "", -1, "/", cookieDomain, cookieSecure, true)
			ctx.JSON(200, gin.H{"message": "Logged out successfully"})
			return
		}

		// Clear the refresh token in database
		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", teacherID).First(&teacher).Error; err == nil {
			teacher.RefreshToken = ""
			handler.DB.Save(&teacher)
		}

		// Clear the cookie
		ctx.SetSameSite(cookieSameSite)
		ctx.SetCookie("refresh_token", "", -1, "/", cookieDomain, cookieSecure, true)

		ctx.JSON(200, gin.H{"message": "Logged out successfully"})
	}
}
