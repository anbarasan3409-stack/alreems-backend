package controllers

import (
	"fmt"
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

// Helper function to get SameSite mode from environment
func getCookieSameSite(sameSite string) http.SameSite {
	switch strings.ToLower(sameSite) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode // Default to Lax
	}
}

func HandleAdminLogin(c *gin.Context, handler database.Handler, email string, token *oauth2.Token) {
	var admin models.Admin
	if err := handler.DB.Where("email = ?", email).First(&admin).Error; err != nil {
		c.JSON(403, gin.H{"error": "Access denied. This email is not authorized for admin access."})
		return
	}
	admin.GoogleAccessToken = token.AccessToken

	now := time.Now()
	admin.LastLogin = &now

	refreshToken, err := utils.GenerateAdminRefreshToken(admin.ID, admin.Email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to generate refresh token", "details": err.Error()})
		return
	}
	admin.RefreshToken = refreshToken
	if err := handler.DB.Save(&admin).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to save admin data", "details": err.Error()})
		return
	}
	adminURL := os.Getenv("ADMIN_URL")

	// Get cookie settings from environment
	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE") == "true"
	cookieSameSite := getCookieSameSite(os.Getenv("COOKIE_SAMESITE"))

	// Set cookie with environment-based configuration
	c.SetSameSite(cookieSameSite)
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*3600, // 7 days
		"/",
		cookieDomain,
		cookieSecure,
		true, // HttpOnly = true (more secure)
	)

	fmt.Printf("Setting cookie for admin login. Token: %s..., URL: %s, Domain: %s, Secure: %v, SameSite: %v\n",
		refreshToken[:20], adminURL, cookieDomain, cookieSecure, cookieSameSite)

	// Redirect to callback page instead of root
	c.Redirect(http.StatusFound, adminURL+"/auth/callback")
}

func GetAdminProfile(c *gin.Context, handler database.Handler) {
	adminID, exists := c.Get("adminID")
	if !exists {
		c.JSON(401, gin.H{"error": "Admin ID not found in context"})
		return
	}

	var admin models.Admin
	if err := handler.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
		c.JSON(404, gin.H{"error": "Admin not found"})
		return
	}

	// Return admin data without sensitive information
	c.JSON(200, gin.H{
		"id":    admin.ID,
		"name":  admin.Name,
		"email": admin.Email,
	})
}

func AdminRefreshTokenController(handler database.Handler) gin.HandlerFunc {
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
		adminID, err := utils.ValidateAdminRefreshToken(refreshToken)
		if err != nil {
			ctx.JSON(401, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Get admin from database
		var admin models.Admin
		if err := handler.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Admin not found"})
			return
		}

		// Generate new refresh token
		newRefreshToken, err := utils.GenerateAdminRefreshToken(admin.ID, admin.Email)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to generate new refresh token"})
			return
		}

		// Update admin's refresh token in database
		admin.RefreshToken = newRefreshToken
		if err := handler.DB.Save(&admin).Error; err != nil {
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
			"admin": gin.H{
				"id":    admin.ID,
				"name":  admin.Name,
				"email": admin.Email,
			},
		})
	}
}

func AdminLogoutController(handler database.Handler) gin.HandlerFunc {
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

		// Validate the refresh token to get admin ID
		adminID, err := utils.ValidateAdminRefreshToken(refreshToken)
		if err != nil {
			// Even if token is invalid, we still want to clear the cookie
			ctx.SetSameSite(cookieSameSite)
			ctx.SetCookie("refresh_token", "", -1, "/", cookieDomain, cookieSecure, true)
			ctx.JSON(200, gin.H{"message": "Logged out successfully"})
			return
		}

		// Clear the refresh token in database
		var admin models.Admin
		if err := handler.DB.Where("id = ?", adminID).First(&admin).Error; err == nil {
			admin.RefreshToken = ""
			handler.DB.Save(&admin)
		}

		// Clear the cookie
		ctx.SetSameSite(cookieSameSite)
		ctx.SetCookie("refresh_token", "", -1, "/", cookieDomain, cookieSecure, true)

		ctx.JSON(200, gin.H{"message": "Logged out successfully"})
	}
}
