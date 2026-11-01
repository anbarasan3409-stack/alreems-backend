package controllers

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateMaterial - Upload a new learning material
func CreateMaterial(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			FileType    string `json:"file_type"`
			FileURL     string `json:"file_url" binding:"required"`
			CourseID    uint   `json:"course_id" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var teacherID *uint
		if user, exists := ctx.Get("user"); exists {
			if uMap, ok := user.(map[string]interface{}); ok {
				if idFloat, ok := uMap["id"].(float64); ok {
					id := uint(idFloat)
					teacherID = &id
				}
			}
		}

		if teacherID == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		material := models.LearningMaterial{
			Title:       input.Title,
			Description: input.Description,
			FileType:    input.FileType,
			FileURL:     input.FileURL,
			CourseID:    input.CourseID,
			TeacherID:   *teacherID,
		}

		if err := handler.DB.Create(&material).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create material"})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": "Material created successfully", "material": material})
	}
}

// GetMaterialsByCourse - Fetch materials for a course
func GetMaterialsByCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		courseID := ctx.Param("courseId")
		var materials []models.LearningMaterial

		if err := handler.DB.Where("course_id = ?", courseID).Find(&materials).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch materials"})
			return
		}

		ctx.JSON(http.StatusOK, materials)
	}
}

// DeleteMaterial - Delete a learning material
func DeleteMaterial(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		var teacherID *uint
		if user, exists := ctx.Get("user"); exists {
			if uMap, ok := user.(map[string]interface{}); ok {
				if idFloat, ok := uMap["id"].(float64); ok {
					id := uint(idFloat)
					teacherID = &id
				}
			}
		}

		if teacherID == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var material models.LearningMaterial
		if err := handler.DB.First(&material, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Material not found"})
			return
		}

		if material.TeacherID != *teacherID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this material"})
			return
		}

		if err := handler.DB.Delete(&material).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete material"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Material deleted successfully"})
	}
}
