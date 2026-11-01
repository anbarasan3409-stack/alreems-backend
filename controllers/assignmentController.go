package controllers

import (
	"backend/database"
	"backend/models"
	"backend/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateAssignment - Create a new assignment
func CreateAssignment(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input struct {
			Title       string    `json:"title" binding:"required"`
			Description string    `json:"description"`
			DueDate     time.Time `json:"due_date" binding:"required"`
			CourseID    uint      `json:"course_id" binding:"required"`
			TotalMarks  int       `json:"total_marks"`
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

		assignment := models.Assignment{
			Title:       input.Title,
			Description: input.Description,
			DueDate:     input.DueDate,
			CourseID:    input.CourseID,
			TotalMarks:  input.TotalMarks,
			TeacherID:   *teacherID,
		}

		if err := handler.DB.Create(&assignment).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
			return
		}

		// Notify students in the course asynchronously
		go func(courseID uint, title string) {
			var course models.Course
			if err := handler.DB.Preload("Enrollments.Student").First(&course, courseID).Error; err == nil {
				for _, enrollment := range course.Enrollments {
					if enrollment.Student.Email != "" {
						subject := fmt.Sprintf("New Assignment: %s", title)
						body := fmt.Sprintf("<h3>Hello %s,</h3><p>A new assignment <b>%s</b> has been posted in your course.</p><p>Please log in to your portal to view it.</p>", enrollment.Student.Name, title)
						utils.SendEmail([]string{enrollment.Student.Email}, subject, body)
					}
					if enrollment.Student.PhoneNumber != "" {
						waMsg := fmt.Sprintf("Hi %s, a new assignment '%s' has been posted. Check your portal for details.", enrollment.Student.Name, title)
						utils.SendWhatsAppMessage(enrollment.Student.PhoneNumber, waMsg)
					}
				}
			}
		}(input.CourseID, input.Title)

		ctx.JSON(http.StatusCreated, gin.H{"message": "Assignment created successfully", "assignment": assignment})
	}
}

// GetAssignmentsByCourse - Fetch assignments for a course
func GetAssignmentsByCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		courseID := ctx.Param("courseId")
		var assignments []models.Assignment

		if err := handler.DB.Where("course_id = ?", courseID).Find(&assignments).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch assignments"})
			return
		}

		ctx.JSON(http.StatusOK, assignments)
	}
}

// DeleteAssignment - Delete an assignment
func DeleteAssignment(handler database.Handler) gin.HandlerFunc {
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

		var assignment models.Assignment
		if err := handler.DB.First(&assignment, id).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Assignment not found"})
			return
		}

		if assignment.TeacherID != *teacherID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this assignment"})
			return
		}

		if err := handler.DB.Delete(&assignment).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete assignment"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Assignment deleted successfully"})
	}
}

// SubmitAssignment - Student submits an assignment
func SubmitAssignment(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input struct {
			AssignmentID uint   `json:"assignment_id" binding:"required"`
			FileURL      string `json:"file_url" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var studentID *uint
		if user, exists := ctx.Get("user"); exists {
			if uMap, ok := user.(map[string]interface{}); ok {
				if idFloat, ok := uMap["id"].(float64); ok {
					id := uint(idFloat)
					studentID = &id
				}
			}
		}

		if studentID == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		submission := models.AssignmentSubmission{
			AssignmentID: input.AssignmentID,
			StudentID:    *studentID,
			FileURL:      input.FileURL,
			Status:       "submitted",
		}

		if err := handler.DB.Create(&submission).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit assignment. You may have already submitted it."})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{"message": "Assignment submitted successfully", "submission": submission})
	}
}

// GetAssignmentSubmissions - Teacher gets all submissions for an assignment
func GetAssignmentSubmissions(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		assignmentID := ctx.Param("assignmentId")
		var submissions []models.AssignmentSubmission

		if err := handler.DB.Preload("Student").Where("assignment_id = ?", assignmentID).Find(&submissions).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
			return
		}

		ctx.JSON(http.StatusOK, submissions)
	}
}

// GradeAssignment - Teacher grades a submission
func GradeAssignment(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		submissionID := ctx.Param("submissionId")
		var input struct {
			MarksObtained int    `json:"marks_obtained" binding:"required"`
			Feedback      string `json:"feedback"`
		}

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var submission models.AssignmentSubmission
		if err := handler.DB.First(&submission, submissionID).Error; err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
			return
		}

		submission.MarksObtained = input.MarksObtained
		submission.Feedback = input.Feedback
		submission.Status = "graded"

		if err := handler.DB.Save(&submission).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grade assignment"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Assignment graded successfully", "submission": submission})
	}
}
