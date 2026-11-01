package controllers

import (
	"backend/database"
	"backend/models"
	"backend/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SubmitEnquiry - Public endpoint to submit a new enquiry
func SubmitEnquiry(c *gin.Context) {
	var input struct {
		Name             string `json:"name" binding:"required"`
		Email            string `json:"email" binding:"required,email"`
		Phone            string `json:"phone"`
		Subject          string `json:"subject"`
		Message          string `json:"message" binding:"required"`
		CourseOfInterest string `json:"course_of_interest"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enquiry := models.Enquiry{
		Name:             input.Name,
		Email:            input.Email,
		Phone:            input.Phone,
		Subject:          input.Subject,
		Message:          input.Message,
		CourseOfInterest: input.CourseOfInterest,
		Status:           "new",
	}

	db := database.Init()
	if err := db.Create(&enquiry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit enquiry"})
		return
	}

	// Send an automated acknowledgment email asynchronously
	go func() {
		subject := "Thank you for your enquiry - Al Reems Academy"
		body := fmt.Sprintf(`
			<h3>Hello %s,</h3>
			<p>Thank you for reaching out to Al Reems Academy.</p>
			<p>We have received your enquiry regarding "%s" and our team will get back to you shortly.</p>
			<p>Best regards,<br>Al Reems Academy Team</p>
		`, enquiry.Name, enquiry.Subject)
		utils.SendEmail([]string{enquiry.Email}, subject, body)
		
		// Send a WhatsApp message if phone is provided
		if enquiry.Phone != "" {
			waMessage := fmt.Sprintf("Hi %s, we've received your enquiry at Al Reems Academy. Our team will contact you shortly.", enquiry.Name)
			utils.SendWhatsAppMessage(enquiry.Phone, waMessage)
		}
	}()

	c.JSON(http.StatusCreated, gin.H{"message": "Enquiry submitted successfully", "enquiry": enquiry})
}

// GetAllEnquiries - Admin endpoint to view all enquiries
func GetAllEnquiries(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var enquiries []models.Enquiry

		// Ensure the user is an admin
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if err := handler.DB.Order("created_at desc").Find(&enquiries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch enquiries"})
			return
		}

		c.JSON(http.StatusOK, enquiries)
	}
}

// UpdateEnquiryStatus - Admin endpoint to update the status of an enquiry
func UpdateEnquiryStatus(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var input struct {
			Status string `json:"status" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Ensure the user is an admin
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var enquiry models.Enquiry
		if err := handler.DB.First(&enquiry, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Enquiry not found"})
			return
		}

		enquiry.Status = input.Status
		if err := handler.DB.Save(&enquiry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update enquiry status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Enquiry status updated successfully", "enquiry": enquiry})
	}
}

// DeleteEnquiry - Admin endpoint to delete an enquiry
func DeleteEnquiry(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Ensure the user is an admin
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var enquiry models.Enquiry
		if err := handler.DB.First(&enquiry, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Enquiry not found"})
			return
		}

		if err := handler.DB.Delete(&enquiry).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete enquiry"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Enquiry deleted successfully"})
	}
}
