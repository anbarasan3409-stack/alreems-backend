package controllers

import (
	"backend/database"
	"backend/models"
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExportStudentsCSV - Admin endpoint to download students as CSV
func ExportStudentsCSV(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure the user is an admin
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var students []models.Student
		if err := handler.DB.Find(&students).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
			return
		}

		b := &bytes.Buffer{}
		writer := csv.NewWriter(b)

		// Header
		writer.Write([]string{"ID", "Name", "Email", "Phone", "Class", "Joined At"})

		// Data
		for _, s := range students {
			writer.Write([]string{
				fmt.Sprintf("%d", s.ID),
				s.Name,
				s.Email,
				s.PhoneNumber,
				s.Class,
				s.CreatedAt.Format("2006-01-02"),
			})
		}
		writer.Flush()

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", "attachment; filename=students.csv")
		c.Header("Content-Type", "text/csv")
		c.String(http.StatusOK, b.String())
	}
}

// ExportEnquiriesCSV - Admin endpoint to download enquiries as CSV
func ExportEnquiriesCSV(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure the user is an admin
		if _, exists := c.Get("adminID"); !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		var enquiries []models.Enquiry
		if err := handler.DB.Order("created_at desc").Find(&enquiries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch enquiries"})
			return
		}

		b := &bytes.Buffer{}
		writer := csv.NewWriter(b)

		// Header
		writer.Write([]string{"ID", "Name", "Email", "Phone", "Subject", "Status", "Course Of Interest", "Date"})

		// Data
		for _, e := range enquiries {
			writer.Write([]string{
				fmt.Sprintf("%d", e.ID),
				e.Name,
				e.Email,
				e.Phone,
				e.Subject,
				e.Status,
				e.CourseOfInterest,
				e.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		writer.Flush()

		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", "attachment; filename=enquiries.csv")
		c.Header("Content-Type", "text/csv")
		c.String(http.StatusOK, b.String())
	}
}
