package controllers

import (
	"backend/database"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SubmitFeedback - Authenticated student submits feedback
func SubmitFeedback(c *gin.Context) {
	var input struct {
		StudentName string `json:"student_name" binding:"required"`
		Course      string `json:"course"`
		Rating      int    `json:"rating" binding:"required,min=1,max=5"`
		Message     string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Optionally extract student_id from context if authenticated
	studentID := uint(0)
	if id, exists := c.Get("studentID"); exists {
		if sid, ok := id.(uint); ok {
			studentID = sid
		}
	}

	feedback := models.Feedback{
		StudentID:   studentID,
		StudentName: input.StudentName,
		Course:      input.Course,
		Rating:      input.Rating,
		Message:     input.Message,
		Featured:    false,
	}

	db := database.Init()
	if err := db.Create(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Feedback submitted successfully", "feedback": feedback})
}

// GetAllFeedbacks - Admin endpoint to view all feedbacks
func GetAllFeedbacks(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var feedbacks []models.Feedback
		if err := handler.DB.Order("created_at desc").Find(&feedbacks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch feedbacks"})
			return
		}

		c.JSON(http.StatusOK, feedbacks)
	}
}

// ToggleFeaturedFeedback - Admin toggles a feedback as featured (testimonial)
func ToggleFeaturedFeedback(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		id := c.Param("id")
		var feedback models.Feedback
		if err := handler.DB.First(&feedback, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
			return
		}

		feedback.Featured = !feedback.Featured
		if err := handler.DB.Save(&feedback).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update feedback"})
			return
		}

		status := "unfeatured"
		if feedback.Featured {
			status = "featured"
		}
		c.JSON(http.StatusOK, gin.H{"message": "Feedback " + status + " successfully", "feedback": feedback})
	}
}

// DeleteFeedback - Admin deletes a feedback
func DeleteFeedback(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		id := c.Param("id")
		var feedback models.Feedback
		if err := handler.DB.First(&feedback, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Feedback not found"})
			return
		}

		if err := handler.DB.Delete(&feedback).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete feedback"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Feedback deleted successfully"})
	}
}

// GetFeaturedFeedbacks - Public endpoint to fetch featured testimonials
func GetFeaturedFeedbacks(c *gin.Context) {
	db := database.Init()
	var feedbacks []models.Feedback
	if err := db.Where("featured = ?", true).Order("created_at desc").Find(&feedbacks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch testimonials"})
		return
	}
	c.JSON(http.StatusOK, feedbacks)
}
