package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"backend/utils"

	"github.com/gin-gonic/gin"
)

func StudentAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var refreshToken string

		// First try to get from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
			fmt.Printf("Found student token in Authorization header\n")
		} else {
			// If no Authorization header, try to get from cookie
			cookie, err := ctx.Cookie("refresh_token")
			if err != nil {
				fmt.Printf("No refresh token found in cookie or header. Cookie error: %v\n", err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No authentication token provided"})
				return
			}
			refreshToken = cookie
			tokenPreview := refreshToken
			if len(refreshToken) > 20 {
				tokenPreview = refreshToken[:20] + "..."
			}
			fmt.Printf("Found student token in cookie: %s\n", tokenPreview)
		}

		// Validate refresh token
		claims, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil {
			fmt.Printf("Student token validation failed: %v\n", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		fmt.Printf("Student token validated successfully for student ID: %d\n", claims.StudentID)
		ctx.Set("studentID", claims.StudentID)
		ctx.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var refreshToken string

		// First try to get from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
			fmt.Printf("Found token in Authorization header\n")
		} else {
			// If no Authorization header, try to get from cookie
			cookie, err := ctx.Cookie("refresh_token")
			if err != nil {
				fmt.Printf("No refresh token found in cookie or header. Cookie error: %v\n", err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No authentication token provided"})
				return
			}
			refreshToken = cookie
			tokenPreview := refreshToken
			if len(refreshToken) > 20 {
				tokenPreview = refreshToken[:20] + "..."
			}
			fmt.Printf("Found token in cookie: %s\n", tokenPreview)
		}

		adminID, err := utils.ValidateAdminRefreshToken(refreshToken)
		if err != nil {
			fmt.Printf("Token validation failed: %v\n", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		fmt.Printf("Token validated successfully for admin ID: %d\n", adminID)
		ctx.Set("adminID", adminID)
		ctx.Next()
	}
}

func TeacherAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var refreshToken string

		// First try to get from Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			refreshToken = strings.TrimPrefix(authHeader, "Bearer ")
			fmt.Printf("Found teacher token in Authorization header\n")
		} else {
			// If no Authorization header, try to get from cookie
			cookie, err := ctx.Cookie("refresh_token")
			if err != nil {
				fmt.Printf("No refresh token found in cookie or header. Cookie error: %v\n", err)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No authentication token provided"})
				return
			}
			refreshToken = cookie
			tokenPreview := refreshToken
			if len(refreshToken) > 20 {
				tokenPreview = refreshToken[:20] + "..."
			}
			fmt.Printf("Found teacher token in cookie: %s\n", tokenPreview)
		}

		// Validate teacher token
		teacherID, err := utils.ValidateTeacherRefreshToken(refreshToken)
		if err != nil {
			fmt.Printf("Teacher token validation failed: %v\n", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		fmt.Printf("Teacher token validated successfully for teacher ID: %d\n", teacherID)
		ctx.Set("teacherID", teacherID)
		ctx.Next()
	}
}
