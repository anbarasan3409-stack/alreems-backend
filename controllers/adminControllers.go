package controllers

import (
	"backend/database"
	"backend/models"
	"backend/utils"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Utility function to generate unique channel name
func generateChannelName() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("session_%s_%d", hex.EncodeToString(bytes), time.Now().Unix())
}

// Helper function to create session response
func createSessionResponse(session models.Session) models.SessionResponse {
	teacherName := ""
	if session.Teacher.ID != 0 {
		teacherName = session.Teacher.Name
	}

	courseName := ""
	var courseIDValue *uint
	if session.CourseID != nil && session.Course.ID != 0 {
		courseName = session.Course.Name
		courseIDValue = session.CourseID
	}

	status := session.Status
	if !session.EndTime.IsZero() && time.Now().After(session.EndTime) && status != "ongoing" && status != "live" {
		status = "completed"
	}

	return models.SessionResponse{
		ID:             session.ID,
		Title:          session.Title,
		Description:    session.Description,
		TeacherID:      session.TeacherID,
		TeacherName:    teacherName,
		CourseID:       courseIDValue,
		CourseName:     courseName,
		StartTime:      session.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		EndTime:        session.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		Duration:       session.Duration,
		Location:       session.Location,
		ChannelName:    session.ChannelName,
		WhiteboardUUID: session.WhiteboardUUID,
		MaxStudents:    session.MaxStudents,
		Status:         status,
		CreatedAt:      session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func GetAllStudents(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var students []models.Student
		if err := handler.DB.Find(&students).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch students"})
			return
		}

		// Return students without sensitive information
		var response []gin.H
		for _, student := range students {
			response = append(response, gin.H{
				"id":             student.ID,
				"name":           student.Name,
				"email":          student.Email,
				"phoneNumber":    student.PhoneNumber,
				"address":        student.Address,
				"profilePicture": student.ProfilePicture,
				"lastLogin":      student.LastLogin,
				"createdAt":      student.CreatedAt,
				"updatedAt":      student.UpdatedAt,
			})
		}

		ctx.JSON(200, response)
	}
}

func GetAdminStudentByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		studentID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid student ID"})
			return
		}

		var student models.Student
		if err := handler.DB.Where("id = ?", studentID).First(&student).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Return student without sensitive information
		response := gin.H{
			"id":             student.ID,
			"name":           student.Name,
			"email":          student.Email,
			"phoneNumber":    student.PhoneNumber,
			"address":        student.Address,
			"profilePicture": student.ProfilePicture,
			"lastLogin":      student.LastLogin,
			"createdAt":      student.CreatedAt,
			"updatedAt":      student.UpdatedAt,
		}

		ctx.JSON(200, response)
	}
}

func CreateStudentByAdmin(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var studentData struct {
			Name        string `json:"name" binding:"required"`
			Email       string `json:"email" binding:"required,email"`
			PhoneNumber string `json:"phoneNumber"`
			Address     string `json:"address"`
			Class       string `json:"class"`
		}

		if err := ctx.ShouldBindJSON(&studentData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		// Check if student with this email already exists
		var existingStudent models.Student
		if err := handler.DB.Where("email = ?", studentData.Email).First(&existingStudent).Error; err == nil {
			ctx.JSON(409, gin.H{"error": "Student with this email already exists"})
			return
		}

		newStudent := models.Student{
			Name:              studentData.Name,
			Email:             studentData.Email,
			PhoneNumber:       studentData.PhoneNumber,
			Address:           studentData.Address,
			Class:             studentData.Class,
			ProfilePicture:    "",
			GoogleAccessToken: "",
			RefreshToken:      "",
		}

		if err := handler.DB.Create(&newStudent).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to create student", "details": err.Error()})
			return
		}

		ctx.JSON(201, gin.H{
			"id":             newStudent.ID,
			"name":           newStudent.Name,
			"email":          newStudent.Email,
			"phoneNumber":    newStudent.PhoneNumber,
			"address":        newStudent.Address,
			"class":          newStudent.Class,
			"profilePicture": newStudent.ProfilePicture,
			"createdAt":      newStudent.CreatedAt,
			"updatedAt":      newStudent.UpdatedAt,
		})
	}
}

func DeleteStudent(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		studentID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid student ID"})
			return
		}

		var student models.Student
		if err := handler.DB.Where("id = ?", studentID).First(&student).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Check if student has any associated enrollments (including soft-deleted ones)
		var enrollmentCount int64
		if err := handler.DB.Unscoped().Model(&models.Enrollment{}).Where("student_id = ?", studentID).Count(&enrollmentCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to check student enrollments", "details": err.Error()})
			return
		}

		log.Printf("Found %d enrollments for student ID %d", enrollmentCount, studentID)

		// If student has enrollments, delete them first (cascade delete) - use Unscoped for hard delete
		if enrollmentCount > 0 {
			log.Printf("Deleting %d enrollments for student ID %d", enrollmentCount, studentID)
			if err := handler.DB.Unscoped().Where("student_id = ?", studentID).Delete(&models.Enrollment{}).Error; err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to delete student enrollments", "details": err.Error()})
				return
			}
			log.Printf("Successfully deleted enrollments for student ID %d", studentID)
		}

		// Now delete the student
		if err := handler.DB.Delete(&student).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete student", "details": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Student deleted successfully"})
	}
}

func GetAllTeachers(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var teachers []models.Teacher
		if err := handler.DB.Find(&teachers).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch teachers"})
			return
		}

		// Return teachers without sensitive information
		var response []gin.H
		for _, teacher := range teachers {
			response = append(response, gin.H{
				"id":             teacher.ID,
				"name":           teacher.Name,
				"email":          teacher.Email,
				"phoneNumber":    teacher.PhoneNumber,
				"address":        teacher.Address,
				"profilePicture": teacher.ProfilePicture,
				"department":     teacher.Department,
				"designation":    teacher.Designation,
				"subject":        teacher.Subject,
				"qualifications": teacher.Qualifications,
				"experience":     teacher.Experience,
				"joiningDate":    teacher.JoiningDate,
				"status":         teacher.Status,
				"lastLogin":      teacher.LastLogin,
				"createdAt":      teacher.CreatedAt,
				"updatedAt":      teacher.UpdatedAt,
			})
		}

		ctx.JSON(200, response)
	}
}

func GetAnalytics(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var courseCount, studentCount, teacherCount int64
		var paymentCount int64
		if err := handler.DB.Model(&models.Course{}).Count(&courseCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count courses"})
			return
		}
		if err := handler.DB.Model(&models.Student{}).Count(&studentCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count students"})
			return
		}
		if err := handler.DB.Model(&models.Teacher{}).Count(&teacherCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count teachers"})
			return
		}
		if err := handler.DB.Model(&models.Payment{}).Count(&paymentCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count payments"})
			return
		}
		ctx.JSON(200, gin.H{
			"courses":  courseCount,
			"students": studentCount,
			"teachers": teacherCount,
			"payments": paymentCount,
		})
	}
}

func CreateCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var course models.Course
		if err := ctx.ShouldBindJSON(&course); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}
		if err := handler.DB.Create(&course).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to create course"})
			return
		}
		ctx.JSON(201, course)
	}
}

func EditCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var course models.Course
		if err := handler.DB.First(&course, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found"})
			return
		}
		if err := ctx.ShouldBindJSON(&course); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}
		if err := handler.DB.Save(&course).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update course"})
			return
		}
		ctx.JSON(200, course)
	}
}

func DeleteCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		courseID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid course ID"})
			return
		}

		var course models.Course
		if err := handler.DB.Where("id = ?", courseID).First(&course).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found"})
			return
		}

		// Use GORM Transaction to safely delete in cascade
		err = handler.DB.Transaction(func(tx *gorm.DB) error {
			// 1. Delete all enrollments referencing this course
			if err := tx.Unscoped().Where("course_id = ?", courseID).Delete(&models.Enrollment{}).Error; err != nil {
				return fmt.Errorf("failed to delete course enrollments: %w", err)
			}

			// 2. Delete sessions and their attendances
			var sessionIDs []uint
			if err := tx.Unscoped().Model(&models.Session{}).Where("course_id = ?", courseID).Pluck("id", &sessionIDs).Error; err != nil {
				return fmt.Errorf("failed to fetch course sessions: %w", err)
			}
			if len(sessionIDs) > 0 {
				if err := tx.Unscoped().Where("session_id IN ?", sessionIDs).Delete(&models.Attendance{}).Error; err != nil {
					return fmt.Errorf("failed to delete session attendances: %w", err)
				}
				if err := tx.Unscoped().Where("id IN ?", sessionIDs).Delete(&models.Session{}).Error; err != nil {
					return fmt.Errorf("failed to delete sessions: %w", err)
				}
			}

			// 3. Delete assignments and their submissions
			var assignmentIDs []uint
			if err := tx.Unscoped().Model(&models.Assignment{}).Where("course_id = ?", courseID).Pluck("id", &assignmentIDs).Error; err != nil {
				return fmt.Errorf("failed to fetch course assignments: %w", err)
			}
			if len(assignmentIDs) > 0 {
				if err := tx.Unscoped().Where("assignment_id IN ?", assignmentIDs).Delete(&models.AssignmentSubmission{}).Error; err != nil {
					return fmt.Errorf("failed to delete assignment submissions: %w", err)
				}
				if err := tx.Unscoped().Where("id IN ?", assignmentIDs).Delete(&models.Assignment{}).Error; err != nil {
					return fmt.Errorf("failed to delete assignments: %w", err)
				}
			}

			// 4. Delete learning materials
			if err := tx.Unscoped().Where("course_id = ?", courseID).Delete(&models.LearningMaterial{}).Error; err != nil {
				return fmt.Errorf("failed to delete learning materials: %w", err)
			}

			// 5. Delete payments associated with this course
			if err := tx.Unscoped().Where("course_id = ?", courseID).Delete(&models.Payment{}).Error; err != nil {
				return fmt.Errorf("failed to delete course payments: %w", err)
			}

			// 6. Delete subscriptions associated with this course
			if err := tx.Unscoped().Where("course_id = ?", courseID).Delete(&models.Subscription{}).Error; err != nil {
				return fmt.Errorf("failed to delete course subscriptions: %w", err)
			}

			// 7. Delete the course itself
			if err := tx.Unscoped().Delete(&course).Error; err != nil {
				return fmt.Errorf("failed to delete course: %w", err)
			}

			return nil
		})

		if err != nil {
			log.Printf("Error deleting course %d: %v", courseID, err)
			ctx.JSON(500, gin.H{"error": "Failed to delete course", "details": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Course and associated records deleted successfully"})
	}
}

func GetAllPayments(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var payments []models.Payment
		if err := handler.DB.Preload("Student").Preload("Course").Order("id desc").Find(&payments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch payments"})
			return
		}
		ctx.JSON(200, payments)
	}
}

func CreatePayment(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input struct {
			StudentID        uint     `json:"studentId"`
			StudentIDAlt     uint     `json:"studentID"`
			CourseID         uint     `json:"courseId"`
			CourseIDAlt      uint     `json:"courseID"`
			Amount           *float64 `json:"amount"`
			Status           string   `json:"status"`
			Method           string   `json:"method"`
			PaymentDate      string   `json:"paymentDate"`
			TransactionID    string   `json:"transactionId"`
			TransactionIDAlt string   `json:"transactionID"`
		}

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request payload", "details": err.Error()})
			return
		}

		studentID := input.StudentID
		if studentID == 0 {
			studentID = input.StudentIDAlt
		}
		courseID := input.CourseID
		if courseID == 0 {
			courseID = input.CourseIDAlt
		}

		if studentID == 0 {
			ctx.JSON(400, gin.H{"error": "Student selection is required"})
			return
		}
		if courseID == 0 {
			ctx.JSON(400, gin.H{"error": "Course selection is required"})
			return
		}
		if input.Amount == nil || *input.Amount < 0 {
			ctx.JSON(400, gin.H{"error": "A valid payment amount is required"})
			return
		}

		// Verify student exists
		var student models.Student
		if err := handler.DB.First(&student, studentID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Verify course exists
		var course models.Course
		if err := handler.DB.First(&course, courseID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found"})
			return
		}

		status := strings.ToLower(strings.TrimSpace(input.Status))
		if status == "" {
			status = "completed"
		}

		method := strings.TrimSpace(input.Method)
		if method == "" {
			method = "Cash"
		}

		var paymentDate time.Time
		dateStr := strings.TrimSpace(input.PaymentDate)
		if dateStr != "" {
			parsed, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				parsedDate, errDate := time.Parse("2006-01-02", dateStr)
				if errDate == nil {
					now := time.Now()
					if parsedDate.Format("2006-01-02") == now.Format("2006-01-02") {
						parsed = now
					} else {
						parsed = time.Date(parsedDate.Year(), parsedDate.Month(), parsedDate.Day(), 12, 0, 0, 0, time.UTC)
					}
					err = nil
				}
			}
			if err == nil {
				paymentDate = parsed
			} else {
				paymentDate = time.Now()
			}
		} else {
			paymentDate = time.Now()
		}

		txnID := strings.TrimSpace(input.TransactionID)
		if txnID == "" {
			txnID = strings.TrimSpace(input.TransactionIDAlt)
		}
		if txnID == "" {
			txnID = fmt.Sprintf("MANUAL-%d-%d", studentID, time.Now().Unix())
		}

		payment := models.Payment{
			StudentID:     studentID,
			CourseID:      courseID,
			Amount:        *input.Amount,
			Status:        status,
			PaymentDate:   &paymentDate,
			Method:        method,
			TransactionID: txnID,
		}

		if err := handler.DB.Create(&payment).Error; err != nil {
			log.Printf("Error creating payment record: %v", err)
			ctx.JSON(500, gin.H{"error": "Failed to create payment record", "details": err.Error()})
			return
		}

		// If payment is completed, ensure student enrollment exists
		if status == "completed" {
			var count int64
			handler.DB.Model(&models.Enrollment{}).Where("student_id = ? AND course_id = ?", studentID, courseID).Count(&count)
			if count == 0 {
				enrollment := models.Enrollment{
					StudentID: studentID,
					CourseID:  courseID,
					Progress:  0,
				}
				handler.DB.Create(&enrollment)
			}
		}

		// Reload with preloads
		handler.DB.Preload("Student").Preload("Course").First(&payment, payment.ID)

		ctx.JSON(201, payment)
	}
}

func UpdatePaymentStatus(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid payment ID"})
			return
		}

		var input struct {
			Status string `json:"status" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(400, gin.H{"error": "Status is required"})
			return
		}

		var payment models.Payment
		if err := handler.DB.First(&payment, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Payment not found"})
			return
		}

		payment.Status = strings.ToLower(strings.TrimSpace(input.Status))
		if err := handler.DB.Save(&payment).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update payment status"})
			return
		}

		handler.DB.Preload("Student").Preload("Course").First(&payment, payment.ID)
		ctx.JSON(200, payment)
	}
}

func DeletePayment(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid payment ID"})
			return
		}

		var payment models.Payment
		if err := handler.DB.First(&payment, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Payment not found"})
			return
		}

		if err := handler.DB.Delete(&payment).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete payment"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Payment deleted successfully"})
	}
}

func GetSystemSettings(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var settings models.PaymentReminderSettings
		if err := handler.DB.First(&settings).Error; err != nil {
			settings = models.PaymentReminderSettings{
				Enabled:                    false,
				ScheduleType:               "monthly",
				DayOfMonth:                 1,
				SendHour:                   9,
				SendMinute:                 0,
				EmailSubject:               "Payment Reminder: Upcoming Course Tuition Fee - Al Reems Academy",
				CustomMessage:              "This is a friendly reminder regarding the upcoming tuition fee payment for the course(s) you are currently enrolled in at Al Reems Academy.",
				StudentRegistrationEnabled: true,
				TeacherRegistrationEnabled: true,
			}
			if createErr := handler.DB.Create(&settings).Error; createErr != nil {
				log.Println("Error creating default settings:", createErr)
			}
		}

		var activeEnrollmentsCount int64
		handler.DB.Model(&models.Enrollment{}).Where("status = ?", "active").Count(&activeEnrollmentsCount)

		var activeStudentsCount int64
		handler.DB.Model(&models.Enrollment{}).Where("status = ?", "active").Distinct("student_id").Count(&activeStudentsCount)

		ctx.JSON(200, gin.H{
			"id":                         settings.ID,
			"studentRegistrationEnabled": settings.StudentRegistrationEnabled,
			"teacherRegistrationEnabled": settings.TeacherRegistrationEnabled,
			"enabled":                    settings.Enabled,
			"scheduleType":               settings.ScheduleType,
			"dayOfMonth":                 settings.DayOfMonth,
			"specificDate":               settings.SpecificDate,
			"sendHour":                   settings.SendHour,
			"sendMinute":                 settings.SendMinute,
			"emailSubject":               settings.EmailSubject,
			"customMessage":              settings.CustomMessage,
			"lastRunAt":                  settings.LastRunAt,
			"lastRunStatus":              settings.LastRunStatus,
			"activeEnrollmentsCount":     activeEnrollmentsCount,
			"activeStudentsCount":        activeStudentsCount,
		})
	}
}

func UpdateSystemSettings(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var input struct {
			StudentRegistrationEnabled *bool   `json:"studentRegistrationEnabled"`
			TeacherRegistrationEnabled *bool   `json:"teacherRegistrationEnabled"`
			Enabled                    *bool   `json:"enabled"`
			ScheduleType               *string `json:"scheduleType"`
			DayOfMonth                 *int    `json:"dayOfMonth"`
			SpecificDate               *string `json:"specificDate"`
			SendHour                   *int    `json:"sendHour"`
			SendMinute                 *int    `json:"sendMinute"`
			EmailSubject               *string `json:"emailSubject"`
			CustomMessage              *string `json:"customMessage"`
		}

		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
			return
		}

		var settings models.PaymentReminderSettings
		if err := handler.DB.First(&settings).Error; err != nil {
			settings = models.PaymentReminderSettings{
				ScheduleType: "monthly",
				DayOfMonth:   1,
				SendHour:     9,
				SendMinute:   0,
				EmailSubject: "Payment Reminder: Upcoming Course Tuition Fee - Al Reems Academy",
			}
			handler.DB.Create(&settings)
		}

		if input.StudentRegistrationEnabled != nil {
			settings.StudentRegistrationEnabled = *input.StudentRegistrationEnabled
		}
		if input.TeacherRegistrationEnabled != nil {
			settings.TeacherRegistrationEnabled = *input.TeacherRegistrationEnabled
		}
		if input.Enabled != nil {
			settings.Enabled = *input.Enabled
		}
		if input.ScheduleType != nil && *input.ScheduleType != "" {
			settings.ScheduleType = *input.ScheduleType
		}
		if input.DayOfMonth != nil {
			settings.DayOfMonth = *input.DayOfMonth
		}
		if input.SpecificDate != nil {
			settings.SpecificDate = *input.SpecificDate
		}
		if input.SendHour != nil {
			settings.SendHour = *input.SendHour
		}
		if input.SendMinute != nil {
			settings.SendMinute = *input.SendMinute
		}
		if input.EmailSubject != nil && *input.EmailSubject != "" {
			settings.EmailSubject = *input.EmailSubject
		}
		if input.CustomMessage != nil {
			settings.CustomMessage = *input.CustomMessage
		}

		if err := handler.DB.Save(&settings).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to save settings"})
			return
		}

		var activeEnrollmentsCount int64
		handler.DB.Model(&models.Enrollment{}).Where("status = ?", "active").Count(&activeEnrollmentsCount)

		var activeStudentsCount int64
		handler.DB.Model(&models.Enrollment{}).Where("status = ?", "active").Distinct("student_id").Count(&activeStudentsCount)

		ctx.JSON(200, gin.H{
			"id":                         settings.ID,
			"studentRegistrationEnabled": settings.StudentRegistrationEnabled,
			"teacherRegistrationEnabled": settings.TeacherRegistrationEnabled,
			"enabled":                    settings.Enabled,
			"scheduleType":               settings.ScheduleType,
			"dayOfMonth":                 settings.DayOfMonth,
			"specificDate":               settings.SpecificDate,
			"sendHour":                   settings.SendHour,
			"sendMinute":                 settings.SendMinute,
			"emailSubject":               settings.EmailSubject,
			"customMessage":              settings.CustomMessage,
			"lastRunAt":                  settings.LastRunAt,
			"lastRunStatus":              settings.LastRunStatus,
			"activeEnrollmentsCount":     activeEnrollmentsCount,
			"activeStudentsCount":        activeStudentsCount,
		})
	}
}

// SendPaymentRemindersNow manually triggers the payment reminder broadcast to all active enrolled students
func SendPaymentRemindersNow(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sentCount, failCount, err := utils.ExecutePaymentReminderJob(handler.DB, "")
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to process payment reminders: " + err.Error()})
			return
		}

		ctx.JSON(200, gin.H{
			"message":   fmt.Sprintf("Payment reminders successfully processed. Sent: %d, Failed: %d", sentCount, failCount),
			"sentCount": sentCount,
			"failCount": failCount,
		})
	}
}

// SendSessionReminderNow manually triggers the class reminder emails for a specific session to both teacher and enrolled students
func SendSessionReminderNow(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		sessionID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid session ID"})
			return
		}

		var session models.Session
		if err := handler.DB.Preload("Teacher").Preload("Course").Where("id = ?", sessionID).First(&session).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		go utils.SendClassReminders(handler.DB, session, true)

		ctx.JSON(200, gin.H{
			"message": fmt.Sprintf("Class reminder emails dispatched for session '%s'", session.Title),
		})
	}
}

// SendTestPaymentReminder sends a preview payment reminder email to the provided email address
func SendTestPaymentReminder(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Email) == "" {
			ctx.JSON(400, gin.H{"error": "A valid destination email address is required"})
			return
		}

		_, _, err := utils.ExecutePaymentReminderJob(handler.DB, req.Email)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to send test email: " + err.Error()})
			return
		}

		ctx.JSON(200, gin.H{
			"message": fmt.Sprintf("Test payment reminder email sent to %s", req.Email),
		})
	}
}

func GetLoginActivities(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		page := 1
		limit := 20

		if pageStr := ctx.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if limitStr := ctx.Query("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
				limit = l
			}
		}

		// Mock login activities data
		// In a real implementation, this would fetch from database
		activities := []gin.H{
			{
				"user":      "Admin User",
				"action":    "login",
				"timestamp": "2025-09-27T10:30:00Z",
				"ipAddress": "192.168.1.100",
				"device":    "Chrome on Windows",
			},
			{
				"user":      "Admin User",
				"action":    "logout",
				"timestamp": "2025-09-27T09:45:00Z",
				"ipAddress": "192.168.1.100",
				"device":    "Chrome on Windows",
			},
			{
				"user":      "Admin User",
				"action":    "login",
				"timestamp": "2025-09-26T14:20:00Z",
				"ipAddress": "192.168.1.100",
				"device":    "Safari on macOS",
			},
		}

		ctx.JSON(200, gin.H{
			"activities":      activities,
			"totalPages":      1,
			"currentPage":     page,
			"totalActivities": len(activities),
			"limit":           limit,
		})
	}
}

func UpdateAdminProfile(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		adminID, exists := ctx.Get("adminID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Admin ID not found in context"})
			return
		}

		var profileData map[string]interface{}
		if err := ctx.ShouldBindJSON(&profileData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		var admin models.Admin
		if err := handler.DB.Where("id = ?", adminID).First(&admin).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Admin not found"})
			return
		}

		// Update allowed fields
		if name, ok := profileData["name"].(string); ok {
			admin.Name = name
		}
		if phone, ok := profileData["phoneNumber"].(string); ok {
			admin.PhoneNumber = phone
		}
		if address, ok := profileData["address"].(string); ok {
			admin.Address = address
		}

		if err := handler.DB.Save(&admin).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update admin profile"})
			return
		}

		ctx.JSON(200, gin.H{
			"id":          admin.ID,
			"name":        admin.Name,
			"email":       admin.Email,
			"phoneNumber": admin.PhoneNumber,
			"address":     admin.Address,
		})
	}
}

// Teacher CRUD Operations

func GetTeacherByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		teacherID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Teacher not found"})
			return
		}

		// Return teacher without sensitive information
		ctx.JSON(200, gin.H{
			"id":             teacher.ID,
			"name":           teacher.Name,
			"email":          teacher.Email,
			"phoneNumber":    teacher.PhoneNumber,
			"address":        teacher.Address,
			"profilePicture": teacher.ProfilePicture,
			"department":     teacher.Department,
			"designation":    teacher.Designation,
			"subject":        teacher.Subject,
			"qualifications": teacher.Qualifications,
			"experience":     teacher.Experience,
			"joiningDate":    teacher.JoiningDate,
			"status":         teacher.Status,
			"lastLogin":      teacher.LastLogin,
			"createdAt":      teacher.CreatedAt,
			"updatedAt":      teacher.UpdatedAt,
		})
	}
}

func CreateTeacher(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var teacherData struct {
			Name           string `json:"name" binding:"required"`
			Email          string `json:"email" binding:"required,email"`
			PhoneNumber    string `json:"phoneNumber" binding:"required"`
			Address        string `json:"address"`
			Department     string `json:"department" binding:"required"`
			Designation    string `json:"designation" binding:"required"`
			Subject        string `json:"subject" binding:"required"`
			Qualifications string `json:"qualifications" binding:"required"`
			Experience     string `json:"experience" binding:"required"`
			JoiningDate    string `json:"joiningDate" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&teacherData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		// Check if teacher with this email already exists
		var existingTeacher models.Teacher
		if err := handler.DB.Where("email = ?", teacherData.Email).First(&existingTeacher).Error; err == nil {
			ctx.JSON(409, gin.H{"error": "Teacher with this email already exists"})
			return
		}

		// Parse joining date
		joiningDate, err := time.Parse("2006-01-02", teacherData.JoiningDate)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid joining date format. Use YYYY-MM-DD"})
			return
		}

		// Create new teacher
		newTeacher := models.Teacher{
			Name:              teacherData.Name,
			Email:             teacherData.Email,
			PhoneNumber:       teacherData.PhoneNumber,
			Address:           teacherData.Address,
			Department:        teacherData.Department,
			Designation:       teacherData.Designation,
			Subject:           teacherData.Subject,
			Qualifications:    teacherData.Qualifications,
			Experience:        teacherData.Experience,
			JoiningDate:       &joiningDate,
			Status:            "active", // Default status
			GoogleAccessToken: "",       // Will be set when teacher logs in via Google
			RefreshToken:      "",       // Will be set when teacher logs in
		}

		if err := handler.DB.Create(&newTeacher).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to create teacher", "details": err.Error()})
			return
		}

		// Return created teacher without sensitive information
		ctx.JSON(201, gin.H{
			"id":             newTeacher.ID,
			"name":           newTeacher.Name,
			"email":          newTeacher.Email,
			"phoneNumber":    newTeacher.PhoneNumber,
			"address":        newTeacher.Address,
			"profilePicture": newTeacher.ProfilePicture,
			"department":     newTeacher.Department,
			"designation":    newTeacher.Designation,
			"subject":        newTeacher.Subject,
			"qualifications": newTeacher.Qualifications,
			"experience":     newTeacher.Experience,
			"joiningDate":    newTeacher.JoiningDate,
			"status":         newTeacher.Status,
			"createdAt":      newTeacher.CreatedAt,
			"updatedAt":      newTeacher.UpdatedAt,
		})
	}
}

func UpdateTeacher(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		teacherID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Teacher not found"})
			return
		}

		var updateData struct {
			Name           string `json:"name"`
			Email          string `json:"email"`
			PhoneNumber    string `json:"phoneNumber"`
			Address        string `json:"address"`
			Department     string `json:"department"`
			Designation    string `json:"designation"`
			Subject        string `json:"subject"`
			Qualifications string `json:"qualifications"`
			Experience     string `json:"experience"`
			JoiningDate    string `json:"joiningDate"`
			Status         string `json:"status"`
		}

		if err := ctx.ShouldBindJSON(&updateData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		// Update only provided fields
		if updateData.Name != "" {
			teacher.Name = updateData.Name
		}
		if updateData.Email != "" {
			// Check if another teacher has this email
			var existingTeacher models.Teacher
			if err := handler.DB.Where("email = ? AND id != ?", updateData.Email, teacherID).First(&existingTeacher).Error; err == nil {
				ctx.JSON(409, gin.H{"error": "Another teacher with this email already exists"})
				return
			}
			teacher.Email = updateData.Email
		}
		if updateData.PhoneNumber != "" {
			teacher.PhoneNumber = updateData.PhoneNumber
		}
		if updateData.Address != "" {
			teacher.Address = updateData.Address
		}
		if updateData.Department != "" {
			teacher.Department = updateData.Department
		}
		if updateData.Designation != "" {
			teacher.Designation = updateData.Designation
		}
		if updateData.Subject != "" {
			teacher.Subject = updateData.Subject
		}
		if updateData.Qualifications != "" {
			teacher.Qualifications = updateData.Qualifications
		}
		if updateData.Experience != "" {
			teacher.Experience = updateData.Experience
		}
		if updateData.JoiningDate != "" {
			joiningDate, err := time.Parse("2006-01-02", updateData.JoiningDate)
			if err != nil {
				ctx.JSON(400, gin.H{"error": "Invalid joining date format. Use YYYY-MM-DD"})
				return
			}
			teacher.JoiningDate = &joiningDate
		}
		if updateData.Status != "" {
			teacher.Status = updateData.Status
		}

		if err := handler.DB.Save(&teacher).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update teacher", "details": err.Error()})
			return
		}

		// Return updated teacher without sensitive information
		ctx.JSON(200, gin.H{
			"id":             teacher.ID,
			"name":           teacher.Name,
			"email":          teacher.Email,
			"phoneNumber":    teacher.PhoneNumber,
			"address":        teacher.Address,
			"profilePicture": teacher.ProfilePicture,
			"department":     teacher.Department,
			"designation":    teacher.Designation,
			"subject":        teacher.Subject,
			"qualifications": teacher.Qualifications,
			"experience":     teacher.Experience,
			"joiningDate":    teacher.JoiningDate,
			"status":         teacher.Status,
			"lastLogin":      teacher.LastLogin,
			"createdAt":      teacher.CreatedAt,
			"updatedAt":      teacher.UpdatedAt,
		})
	}
}

func DeleteTeacher(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		teacherID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", teacherID).First(&teacher).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Teacher not found"})
			return
		}

		// Check if teacher has any associated sessions (including soft-deleted ones)
		var sessionCount int64
		if err := handler.DB.Unscoped().Model(&models.Session{}).Where("teacher_id = ?", teacherID).Count(&sessionCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to check teacher sessions", "details": err.Error()})
			return
		}

		log.Printf("Found %d sessions for teacher ID %d", sessionCount, teacherID)

		// If teacher has sessions, delete them first (cascade delete) - use Unscoped for hard delete
		if sessionCount > 0 {
			log.Printf("Deleting %d sessions for teacher ID %d", sessionCount, teacherID)
			if err := handler.DB.Unscoped().Where("teacher_id = ?", teacherID).Delete(&models.Session{}).Error; err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to delete teacher sessions", "details": err.Error()})
				return
			}
			log.Printf("Successfully deleted sessions for teacher ID %d", teacherID)
		}

		// Check if teacher has any associated courses
		var courseCount int64
		if err := handler.DB.Model(&models.Course{}).Where("teacher_id = ?", teacherID).Count(&courseCount).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to check teacher courses", "details": err.Error()})
			return
		}

		log.Printf("Found %d courses for teacher ID %d", courseCount, teacherID)

		// If teacher has courses, we need to delete enrollments first, then courses
		if courseCount > 0 {
			// Get all course IDs for this teacher
			var courseIDs []uint
			if err := handler.DB.Model(&models.Course{}).Where("teacher_id = ?", teacherID).Pluck("id", &courseIDs).Error; err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to get teacher course IDs", "details": err.Error()})
				return
			}

			// Delete all enrollments for these courses (including soft-deleted ones)
			if len(courseIDs) > 0 {
				var enrollmentCount int64
				if err := handler.DB.Unscoped().Model(&models.Enrollment{}).Where("course_id IN ?", courseIDs).Count(&enrollmentCount).Error; err != nil {
					ctx.JSON(500, gin.H{"error": "Failed to check course enrollments", "details": err.Error()})
					return
				}

				if enrollmentCount > 0 {
					log.Printf("Deleting %d enrollments for teacher ID %d courses", enrollmentCount, teacherID)
					if err := handler.DB.Unscoped().Where("course_id IN ?", courseIDs).Delete(&models.Enrollment{}).Error; err != nil {
						ctx.JSON(500, gin.H{"error": "Failed to delete course enrollments", "details": err.Error()})
						return
					}
					log.Printf("Successfully deleted %d enrollments", enrollmentCount)
				}
			}

			// Now delete the courses
			log.Printf("Deleting %d courses for teacher ID %d", courseCount, teacherID)
			if err := handler.DB.Where("teacher_id = ?", teacherID).Delete(&models.Course{}).Error; err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to delete teacher courses", "details": err.Error()})
				return
			}
			log.Printf("Successfully deleted courses for teacher ID %d", teacherID)
		}

		// Now delete the teacher
		if err := handler.DB.Delete(&teacher).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete teacher", "details": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Teacher deleted successfully"})
	}
} // Session CRUD Operations

// GetAllSessions - Get all sessions (admin only)
func GetAllSessions(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Automatically update past scheduled/active sessions to completed
		_ = handler.DB.Model(&models.Session{}).
			Where("end_time < ? AND status IN ('scheduled', 'active')", time.Now()).
			Update("status", "completed").Error

		var sessions []models.Session
		// Only preload minimal Teacher and Course data - avoid loading full objects with nested relations
		if err := handler.DB.
			Select("sessions.*").
			Preload("Teacher", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, name")
			}).
			Preload("Course", func(db *gorm.DB) *gorm.DB {
				return db.Select("id, name")
			}).
			Order("start_time desc").
			Find(&sessions).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch sessions", "details": err.Error()})
			return
		}

		// Count attendance per session
		type AttendanceCountResult struct {
			SessionID uint `gorm:"column:session_id"`
			Count     int  `gorm:"column:count"`
		}
		var attendanceCounts []AttendanceCountResult
		handler.DB.Model(&models.Attendance{}).
			Select("session_id, count(DISTINCT student_id) as count").
			Where("status IN ('present', 'late')").
			Group("session_id").
			Scan(&attendanceCounts)

		attendanceMap := make(map[uint]int)
		for _, r := range attendanceCounts {
			attendanceMap[r.SessionID] = r.Count
		}

		// Count active enrollments per course
		type CourseCountResult struct {
			CourseID uint `gorm:"column:course_id"`
			Count    int  `gorm:"column:count"`
		}
		var courseEnrollmentCounts []CourseCountResult
		handler.DB.Model(&models.Enrollment{}).
			Select("course_id, count(DISTINCT student_id) as count").
			Where("LOWER(COALESCE(status, 'active')) = 'active' OR status = '' OR status IS NULL").
			Group("course_id").
			Scan(&courseEnrollmentCounts)

		enrollmentMap := make(map[uint]int)
		for _, r := range courseEnrollmentCounts {
			enrollmentMap[r.CourseID] = r.Count
		}

		// Convert to response format
		response := []models.SessionResponse{} // Initialize as empty slice instead of nil
		for _, session := range sessions {
			resp := createSessionResponse(session)
			if count, ok := attendanceMap[session.ID]; ok {
				resp.StudentsAttended = count
			}
			if session.CourseID != nil {
				if count, ok := enrollmentMap[*session.CourseID]; ok {
					resp.EnrolledStudents = count
					// If studentsAttended is 0 for a session with enrollments, use enrolled students as fallback
					if resp.StudentsAttended == 0 && count > 0 {
						resp.StudentsAttended = count
					}
				}
			}
			response = append(response, resp)
		}

		ctx.JSON(200, response)
	}
}

// GetSessionByID - Get session by ID (admin only)
func GetSessionByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		sessionID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid session ID"})
			return
		}

		var session models.Session
		if err := handler.DB.Preload("Teacher").Preload("Course").Where("id = ?", sessionID).First(&session).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		// Ensure channel_name is set
		if session.ChannelName == "" {
			session.ChannelName = fmt.Sprintf("session_%d", session.ID)
			handler.DB.Model(&session).Update("channel_name", session.ChannelName)
		}

		var attendanceCount int64
		handler.DB.Model(&models.Attendance{}).Where("session_id = ? AND status IN ('present', 'late')", session.ID).Distinct("student_id").Count(&attendanceCount)

		var enrollmentCount int64
		if session.CourseID != nil {
			handler.DB.Model(&models.Enrollment{}).Where("course_id = ? AND (LOWER(COALESCE(status, 'active')) = 'active' OR status = '' OR status IS NULL)", *session.CourseID).Distinct("student_id").Count(&enrollmentCount)
		}

		resp := createSessionResponse(session)
		resp.StudentsAttended = int(attendanceCount)
		resp.EnrolledStudents = int(enrollmentCount)
		if resp.StudentsAttended == 0 && enrollmentCount > 0 {
			resp.StudentsAttended = int(enrollmentCount)
		}

		ctx.JSON(200, resp)
	}
}

// CreateSession - Create new session (admin only)
func CreateSession(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sessionData struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			TeacherID   uint   `json:"teacher_id" binding:"required"`
			CourseID    uint   `json:"course_id"`
			StartTime   string `json:"start_time" binding:"required"`
			EndTime     string `json:"end_time" binding:"required"`
			Location    string `json:"location"`
			MaxStudents int    `json:"max_students"`
			Status      string `json:"status"`
		}

		if err := ctx.ShouldBindJSON(&sessionData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		// Parse time strings
		startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sessionData.StartTime)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid start time format"})
			return
		}

		endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sessionData.EndTime)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid end time format"})
			return
		}

		// Calculate duration in minutes
		duration := int(endTime.Sub(startTime).Minutes())

		// Set defaults
		if sessionData.MaxStudents == 0 {
			sessionData.MaxStudents = 30
		}
		if sessionData.Status == "" {
			sessionData.Status = "scheduled"
		}

		// Verify teacher exists
		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", sessionData.TeacherID).First(&teacher).Error; err != nil {
			ctx.JSON(400, gin.H{"error": "Teacher not found"})
			return
		}

		// Verify course exists if provided
		var courseID *uint
		if sessionData.CourseID != 0 {
			var course models.Course
			if err := handler.DB.Where("id = ?", sessionData.CourseID).First(&course).Error; err != nil {
				ctx.JSON(400, gin.H{"error": "Course not found"})
				return
			}
			courseID = &sessionData.CourseID
		}

		// Generate unique channel name for Agora
		channelName := generateChannelName()

		session := models.Session{
			Title:       sessionData.Title,
			Description: sessionData.Description,
			TeacherID:   sessionData.TeacherID,
			CourseID:    courseID,
			StartTime:   startTime,
			EndTime:     endTime,
			Duration:    duration,
			Location:    sessionData.Location,
			ChannelName: channelName,
			MaxStudents: sessionData.MaxStudents,
			Status:      sessionData.Status,
		}

		if err := handler.DB.Create(&session).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to create session", "details": err.Error()})
			return
		}

		// Load the created session with relationships
		if err := handler.DB.Preload("Teacher").Preload("Course").Where("id = ?", session.ID).First(&session).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to load created session"})
			return
		}

		teacherName := ""
		if session.Teacher.ID != 0 {
			teacherName = session.Teacher.Name
		}

		courseName := ""
		if session.Course.ID != 0 {
			courseName = session.Course.Name
		}

		response := models.SessionResponse{
			ID:          session.ID,
			Title:       session.Title,
			Description: session.Description,
			TeacherID:   session.TeacherID,
			TeacherName: teacherName,
			CourseID:    session.CourseID,
			CourseName:  courseName,
			StartTime:   session.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			EndTime:     session.EndTime.Format("2006-01-02T15:04:05Z07:00"),
			Duration:    session.Duration,
			Location:    session.Location,
			ChannelName: session.ChannelName,
			MaxStudents: session.MaxStudents,
			Status:      session.Status,
			CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		ctx.JSON(201, response)

		// Trigger instant class reminder email to teacher and enrolled students if it's a live/instant session
		isInstant := session.Status == "live" || session.StartTime.Before(time.Now().Add(5*time.Minute))
		if isInstant {
			go utils.SendClassReminders(handler.DB, session, true)
		}
	}
}

// UpdateSession - Update session (admin only)
func UpdateSession(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		sessionID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid session ID"})
			return
		}

		var sessionData struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			TeacherID   uint   `json:"teacher_id"`
			CourseID    uint   `json:"course_id"`
			StartTime   string `json:"start_time"`
			EndTime     string `json:"end_time"`
			Location    string `json:"location"`
			MaxStudents int    `json:"max_students"`
			Status      string `json:"status"`
		}

		if err := ctx.ShouldBindJSON(&sessionData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid input", "details": err.Error()})
			return
		}

		var session models.Session
		if err := handler.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		// Update fields if provided
		if sessionData.Title != "" {
			session.Title = sessionData.Title
		}
		if sessionData.Description != "" {
			session.Description = sessionData.Description
		}
		if sessionData.TeacherID != 0 {
			// Verify teacher exists
			var teacher models.Teacher
			if err := handler.DB.Where("id = ?", sessionData.TeacherID).First(&teacher).Error; err != nil {
				ctx.JSON(400, gin.H{"error": "Teacher not found"})
				return
			}
			session.TeacherID = sessionData.TeacherID
		}
		if sessionData.CourseID != 0 {
			// Verify course exists
			var course models.Course
			if err := handler.DB.Where("id = ?", sessionData.CourseID).First(&course).Error; err != nil {
				ctx.JSON(400, gin.H{"error": "Course not found"})
				return
			}
			session.CourseID = &sessionData.CourseID
		}
		if sessionData.StartTime != "" {
			startTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sessionData.StartTime)
			if err != nil {
				ctx.JSON(400, gin.H{"error": "Invalid start time format"})
				return
			}
			session.StartTime = startTime
		}
		if sessionData.EndTime != "" {
			endTime, err := time.Parse("2006-01-02T15:04:05Z07:00", sessionData.EndTime)
			if err != nil {
				ctx.JSON(400, gin.H{"error": "Invalid end time format"})
				return
			}
			session.EndTime = endTime
		}
		if sessionData.Location != "" {
			session.Location = sessionData.Location
		}
		if sessionData.MaxStudents > 0 {
			session.MaxStudents = sessionData.MaxStudents
		}
		if sessionData.Status != "" {
			session.Status = sessionData.Status
		}

		if err := handler.DB.Save(&session).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update session", "details": err.Error()})
			return
		}

		// Load updated session with relationships
		if err := handler.DB.Preload("Teacher").Preload("Course").Where("id = ?", session.ID).First(&session).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to load updated session"})
			return
		}

		teacherName := ""
		if session.Teacher.ID != 0 {
			teacherName = session.Teacher.Name
		}

		courseName := ""
		if session.Course.ID != 0 {
			courseName = session.Course.Name
		}

		response := models.SessionResponse{
			ID:          session.ID,
			Title:       session.Title,
			Description: session.Description,
			TeacherID:   session.TeacherID,
			TeacherName: teacherName,
			CourseID:    session.CourseID,
			CourseName:  courseName,
			StartTime:   session.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			EndTime:     session.EndTime.Format("2006-01-02T15:04:05Z07:00"),
			Location:    session.Location,
			MaxStudents: session.MaxStudents,
			Status:      session.Status,
			CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		ctx.JSON(200, response)

		// Trigger instant class reminder email to teacher and students if status was updated to live
		if session.Status == "live" {
			go utils.SendClassReminders(handler.DB, session, true)
		}
	}
}

// DeleteSession - Delete session (admin only)
func DeleteSession(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		sessionID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid session ID"})
			return
		}

		var session models.Session
		if err := handler.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		if err := handler.DB.Delete(&session).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete session", "details": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Session deleted successfully"})
	}
}

// RTC Token Generation for Agora Live Sessions

// GenerateRTCToken - Generate Agora RTC token for live sessions
func GenerateRTCToken(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelName := ctx.Query("channelName")
		uidStr := ctx.Query("uid")
		roleStr := ctx.Query("role") // "teacher" or "student"

		if channelName == "" || uidStr == "" || roleStr == "" {
			ctx.JSON(400, gin.H{"error": "Missing parameters: channelName, uid, and role are required"})
			return
		}

		// Verify that the session with this channel name exists
		var session models.Session
		if err := handler.DB.Where("channel_name = ?", channelName).First(&session).Error; err != nil {
			// Fallback: check if channelName is formatted as session_{id}
			var sessionID uint64
			if _, scanErr := fmt.Sscanf(channelName, "session_%d", &sessionID); scanErr == nil && sessionID > 0 {
				if err2 := handler.DB.Where("id = ?", sessionID).First(&session).Error; err2 == nil {
					if session.ChannelName == "" {
						session.ChannelName = channelName
						handler.DB.Model(&session).Update("channel_name", channelName)
					}
				} else {
					ctx.JSON(404, gin.H{"error": "Session not found"})
					return
				}
			} else {
				ctx.JSON(404, gin.H{"error": "Session not found"})
				return
			}
		}

		// Use the utility function to generate token by creating a mock request
		utils.GenerateAgoraToken(ctx.Writer, ctx.Request)
	}
}

// GetAgoraAppID - Get Agora App ID for frontend
func GetAgoraAppID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appID := utils.GetAgoraAppID()
		ctx.JSON(200, gin.H{"appId": appID})
	}
}

// Whiteboard Token Generation

// GenerateWhiteboardToken - Generate whiteboard room token
func GenerateWhiteboardToken(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionIDStr := ctx.Query("sessionId")
		role := ctx.Query("role") // "admin", "writer", or "reader"

		if sessionIDStr == "" {
			ctx.JSON(400, gin.H{"error": "Missing parameter: sessionId is required"})
			return
		}

		if role == "" {
			role = "writer" // Default role
		}

		sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid session ID"})
			return
		}

		// Get session from database
		var session models.Session
		if err := handler.DB.Where("id = ?", sessionID).First(&session).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		// Create whiteboard room if not exists
		if session.WhiteboardUUID == "" {
			roomUUID, err := utils.CreateWhiteboardRoom(false) // not recording by default
			if err != nil {
				log.Printf("Failed to create whiteboard room: %v", err)
				ctx.JSON(500, gin.H{"error": "Failed to create whiteboard room", "details": err.Error()})
				return
			}

			// Update session with whiteboard UUID
			session.WhiteboardUUID = roomUUID
			if err := handler.DB.Save(&session).Error; err != nil {
				ctx.JSON(500, gin.H{"error": "Failed to save whiteboard UUID"})
				return
			}
		}

		// Generate room token
		token, err := utils.GenerateRoomToken(session.WhiteboardUUID, role, 3600000) // 1 hour
		if err != nil {
			log.Printf("Failed to generate whiteboard token: %v", err)
			ctx.JSON(500, gin.H{"error": "Failed to generate whiteboard token", "details": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{
			"roomUUID":  session.WhiteboardUUID,
			"roomToken": token,
			"role":      role,
		})
	}
}

// GetNetlessAppID - Get Netless (Whiteboard) App ID for frontend
func GetNetlessAppID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appID := os.Getenv("NETLESS_APP_ID")
		ctx.JSON(200, gin.H{"appId": appID})
	}
}

// GetStudentsByTeacher - Get all students enrolled in courses taught by a specific teacher
func GetStudentsByTeacher(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		teacherID := ctx.Param("teacherId")
		teacherIDUint, err := strconv.ParseUint(teacherID, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		// First, verify teacher exists
		var teacher models.Teacher
		if err := handler.DB.Where("id = ?", teacherIDUint).First(&teacher).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Teacher not found"})
			return
		}

		// Find all courses taught by this teacher
		var courses []models.Course
		if err := handler.DB.Where("teacher_id = ?", teacherIDUint).Find(&courses).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch courses"})
			return
		}

		// Get course IDs
		var courseIDs []uint
		for _, course := range courses {
			courseIDs = append(courseIDs, course.ID)
		}

		// If no courses, return empty array
		if len(courseIDs) == 0 {
			ctx.JSON(200, []gin.H{})
			return
		}

		// Find all enrollments for these courses
		var enrollments []models.Enrollment
		if err := handler.DB.Preload("Student").Preload("Course").Where("course_id IN ?", courseIDs).Find(&enrollments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch enrollments"})
			return
		}

		// Transform enrollments to student data with course information
		var response []gin.H
		for _, enrollment := range enrollments {
			studentData := gin.H{
				"id":             enrollment.Student.ID,
				"name":           enrollment.Student.Name,
				"email":          enrollment.Student.Email,
				"phoneNumber":    enrollment.Student.PhoneNumber,
				"address":        enrollment.Student.Address,
				"profilePicture": enrollment.Student.ProfilePicture,
				"courseId":       enrollment.Course.ID,
				"courseName":     enrollment.Course.Name,
				"enrollmentDate": enrollment.EnrolledAt,
				"status":         enrollment.Status,
				"progress":       enrollment.Progress,
			}
			response = append(response, studentData)
		}

		ctx.JSON(200, response)
	}
}

// GetStudentsByCourse - Get all students enrolled in a specific course
func GetStudentsByCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		courseID := ctx.Param("courseId")
		courseIDUint, err := strconv.ParseUint(courseID, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid course ID"})
			return
		}

		// First, verify course exists
		var course models.Course
		if err := handler.DB.Preload("Teacher").Where("id = ?", courseIDUint).First(&course).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found"})
			return
		}

		// Find all enrollments for this course
		var enrollments []models.Enrollment
		if err := handler.DB.Preload("Student").Where("course_id = ?", courseIDUint).Find(&enrollments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch enrollments"})
			return
		}

		// Transform enrollments to student data
		var response []gin.H
		for _, enrollment := range enrollments {
			studentData := gin.H{
				"id":             enrollment.Student.ID,
				"name":           enrollment.Student.Name,
				"email":          enrollment.Student.Email,
				"phoneNumber":    enrollment.Student.PhoneNumber,
				"address":        enrollment.Student.Address,
				"profilePicture": enrollment.Student.ProfilePicture,
				"enrollmentDate": enrollment.EnrolledAt,
				"status":         enrollment.Status,
				"progress":       enrollment.Progress,
			}
			response = append(response, studentData)
		}

		ctx.JSON(200, response)
	}
}

// GetCoursesByStudent - Get all courses a student is enrolled in
func GetCoursesByStudent(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		studentID := ctx.Param("studentId")
		studentIDUint, err := strconv.ParseUint(studentID, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid student ID"})
			return
		}

		// First, verify student exists
		var student models.Student
		if err := handler.DB.Where("id = ?", studentIDUint).First(&student).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Find all enrollments for this student
		var enrollments []models.Enrollment
		if err := handler.DB.Preload("Course").Preload("Course.Teacher").Where("student_id = ?", studentIDUint).Find(&enrollments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch enrollments"})
			return
		}

		// Transform enrollments to course data
		var response []gin.H
		for _, enrollment := range enrollments {
			courseData := gin.H{
				"id":             enrollment.Course.ID,
				"name":           enrollment.Course.Name,
				"code":           enrollment.Course.Code,
				"description":    enrollment.Course.Description,
				"credits":        enrollment.Course.Credits,
				"teacherId":      enrollment.Course.TeacherID,
				"teacherName":    enrollment.Course.Teacher.Name,
				"enrollmentDate": enrollment.EnrolledAt,
				"status":         enrollment.Status,
				"progress":       enrollment.Progress,
			}
			response = append(response, courseData)
		}

		ctx.JSON(200, response)
	}
}

// GetStats - Get platform statistics for admin dashboard
func GetStats(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var stats struct {
			TotalTeachers int64 `json:"totalTeachers"`
			TotalCourses  int64 `json:"totalCourses"`
			TotalStudents int64 `json:"totalStudents"`
			TotalAdmins   int64 `json:"totalAdmins"`
		}

		// Count total teachers
		if err := handler.DB.Model(&models.Teacher{}).Count(&stats.TotalTeachers).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count teachers"})
			return
		}

		// Count total courses
		if err := handler.DB.Model(&models.Course{}).Count(&stats.TotalCourses).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count courses"})
			return
		}

		// Count total students
		if err := handler.DB.Model(&models.Student{}).Count(&stats.TotalStudents).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count students"})
			return
		}

		// Count total admins
		if err := handler.DB.Model(&models.Admin{}).Count(&stats.TotalAdmins).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to count admins"})
			return
		}

		ctx.JSON(200, stats)
	}
}

// Admin CRUD Operations

// GetAllAdmins - Get all admins
func GetAllAdmins(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var admins []models.Admin
		if err := handler.DB.Find(&admins).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch admins"})
			return
		}

		// Return admins without sensitive information
		var response []gin.H
		for _, admin := range admins {
			response = append(response, gin.H{
				"ID":          admin.ID,
				"Name":        admin.Name,
				"Email":       admin.Email,
				"PhoneNumber": admin.PhoneNumber,
				"Address":     admin.Address,
				"Role":        admin.Role,
				"CreatedAt":   admin.CreatedAt,
				"LastLogin":   admin.LastLogin,
			})
		}

		ctx.JSON(200, response)
	}
}

// GetAdminByID - Get admin by ID
func GetAdminByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var admin models.Admin
		if err := handler.DB.Where("id = ?", id).First(&admin).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Admin not found"})
			return
		}

		ctx.JSON(200, gin.H{
			"ID":          admin.ID,
			"Name":        admin.Name,
			"Email":       admin.Email,
			"PhoneNumber": admin.PhoneNumber,
			"Address":     admin.Address,
			"Role":        admin.Role,
			"CreatedAt":   admin.CreatedAt,
			"LastLogin":   admin.LastLogin,
		})
	}
}

// CreateAdmin - Create new admin
func CreateAdmin(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var adminData struct {
			Name        string `json:"name" binding:"required"`
			Email       string `json:"email" binding:"required,email"`
			PhoneNumber string `json:"phoneNumber"`
			Address     string `json:"address"`
		}

		if err := ctx.ShouldBindJSON(&adminData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		// Check if admin with same email already exists
		var existingAdmin models.Admin
		if err := handler.DB.Where("email = ?", adminData.Email).First(&existingAdmin).Error; err == nil {
			ctx.JSON(400, gin.H{"error": "Admin with this email already exists"})
			return
		}

		admin := models.Admin{
			Name:        adminData.Name,
			Email:       adminData.Email,
			PhoneNumber: adminData.PhoneNumber,
			Address:     adminData.Address,
			Role:        "admin",
		}

		if err := handler.DB.Create(&admin).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to create admin", "details": err.Error()})
			return
		}

		ctx.JSON(201, gin.H{
			"message": "Admin created successfully",
			"admin": gin.H{
				"ID":          admin.ID,
				"Name":        admin.Name,
				"Email":       admin.Email,
				"PhoneNumber": admin.PhoneNumber,
				"Address":     admin.Address,
				"Role":        admin.Role,
			},
		})
	}
}

// UpdateAdmin - Update admin
func UpdateAdmin(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var admin models.Admin
		if err := handler.DB.Where("id = ?", id).First(&admin).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Admin not found"})
			return
		}

		var updateData struct {
			Name        string `json:"name"`
			PhoneNumber string `json:"phoneNumber"`
			Address     string `json:"address"`
		}

		if err := ctx.ShouldBindJSON(&updateData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		// Update allowed fields
		if updateData.Name != "" {
			admin.Name = updateData.Name
		}
		if updateData.PhoneNumber != "" {
			admin.PhoneNumber = updateData.PhoneNumber
		}
		if updateData.Address != "" {
			admin.Address = updateData.Address
		}

		if err := handler.DB.Save(&admin).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update admin"})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "Admin updated successfully",
			"admin": gin.H{
				"ID":          admin.ID,
				"Name":        admin.Name,
				"Email":       admin.Email,
				"PhoneNumber": admin.PhoneNumber,
				"Address":     admin.Address,
			},
		})
	}
}

// DeleteAdmin - Delete admin
func DeleteAdmin(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")

		// Get current admin ID from context
		currentAdminID, exists := ctx.Get("adminID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Admin ID not found in context"})
			return
		}

		// Prevent admin from deleting themselves
		if fmt.Sprintf("%v", currentAdminID) == id {
			ctx.JSON(400, gin.H{"error": "You cannot delete yourself"})
			return
		}

		if err := handler.DB.Delete(&models.Admin{}, id).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to delete admin"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Admin deleted successfully"})
	}
}
