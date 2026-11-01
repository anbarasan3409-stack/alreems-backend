package controllers

import (
	"backend/database"
	"backend/models"
	"backend/utils"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

func GetStudentByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}
		ctx.JSON(200, student)
	}
}

func UpdateStudentProfile(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Bind the request body to update data
		var updateData struct {
			Name           string `json:"name"`
			PhoneNumber    string `json:"phoneNumber"`
			Phone_Number   string `json:"phone_number"` // Support snake_case
			Address        string `json:"address"`
			ProfilePicture string `json:"profilePicture"`
		}

		if err := ctx.ShouldBindJSON(&updateData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		// Update only the provided fields
		if updateData.Name != "" {
			student.Name = updateData.Name
		}
		// Support both phoneNumber and phone_number
		if updateData.PhoneNumber != "" {
			student.PhoneNumber = updateData.PhoneNumber
		} else if updateData.Phone_Number != "" {
			student.PhoneNumber = updateData.Phone_Number
		}
		if updateData.Address != "" {
			student.Address = updateData.Address
		}
		if updateData.ProfilePicture != "" {
			student.ProfilePicture = updateData.ProfilePicture
		}

		if err := handler.DB.Save(&student).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update student"})
			return
		}

		// Return updated profile
		ctx.JSON(200, gin.H{
			"message": "Student profile updated successfully",
			"profile": gin.H{
				"ID":             student.ID,
				"name":           student.Name,
				"email":          student.Email,
				"phoneNumber":    student.PhoneNumber,
				"address":        student.Address,
				"profilePicture": student.ProfilePicture,
				"lastLogin":      student.LastLogin,
				"createdAt":      student.CreatedAt,
				"updatedAt":      student.UpdatedAt,
			},
		})
	}
}

func RegisterStudentForCourse(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var req struct {
			CourseID uint `json:"course_id"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		var course models.Course
		if err := handler.DB.First(&course, req.CourseID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found"})
			return
		}

		// Check if student is already enrolled
		var existingEnrollment models.Enrollment
		if err := handler.DB.Where("student_id = ? AND course_id = ?", student.ID, course.ID).First(&existingEnrollment).Error; err == nil {
			ctx.JSON(409, gin.H{"error": "Student is already enrolled in this course"})
			return
		}

		// Create new enrollment
		enrollment := models.Enrollment{
			StudentID: student.ID,
			CourseID:  course.ID,
			Status:    "active",
			Progress:  0,
		}

		if err := handler.DB.Create(&enrollment).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to register for course"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Student registered for course successfully"})
	}
}

func GetEnrolledCourses(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		var enrollments []models.Enrollment
		if err := handler.DB.Preload("Course.Teacher").Where("student_id = ?", student.ID).Find(&enrollments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch enrolled courses"})
			return
		}

		// Transform enrollments to course data with progress
		var enrolledCourses []gin.H
		for _, enrollment := range enrollments {
			courseData := gin.H{
				"id":           enrollment.Course.ID,
				"name":         enrollment.Course.Name,
				"code":         enrollment.Course.Code,
				"description":  enrollment.Course.Description,
				"credits":      enrollment.Course.Credits,
				"teacher_id":   enrollment.Course.TeacherID,
				"teacher_name": enrollment.Course.Teacher.Name,
				"progress":     enrollment.Progress,
				"status":       enrollment.Status,
				"enrolled_at":  enrollment.EnrolledAt,
				"completed_at": enrollment.CompletedAt,
				"created_at":   enrollment.Course.CreatedAt,
				"updated_at":   enrollment.Course.UpdatedAt,
			}
			enrolledCourses = append(enrolledCourses, courseData)
		}

		ctx.JSON(200, enrolledCourses)
	}
}

// GetStudentSessions returns all sessions that a student can view
func GetStudentSessions(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var sessions []models.Session

		// Get all sessions with preloaded teacher and course data
		if err := handler.DB.Preload("Teacher").Preload("Course").Find(&sessions).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch sessions"})
			return
		}

		// Transform sessions to include teacher and course names
		var response []gin.H
		for _, session := range sessions {
			status := session.Status
			if !session.EndTime.IsZero() && time.Now().After(session.EndTime) && status != "ongoing" && status != "live" {
				status = "completed"
			}

			sessionData := gin.H{
				"id":           session.ID,
				"title":        session.Title,
				"description":  session.Description,
				"course_id":    session.CourseID,
				"teacher_id":   session.TeacherID,
				"start_time":   session.StartTime,
				"end_time":     session.EndTime,
				"max_students": session.MaxStudents,
				"channel_name": session.ChannelName,
				"status":       status,
				"location":     session.Location,
				"duration":     session.Duration,
				"created_at":   session.CreatedAt,
				"updated_at":   session.UpdatedAt,
			}

			// Add teacher info
			sessionData["teacher_name"] = session.Teacher.Name
			sessionData["teacher_email"] = session.Teacher.Email

			// Add course info if course is linked
			if session.CourseID != nil {
				sessionData["course_name"] = session.Course.Name
				sessionData["course_description"] = session.Course.Description
			}

			response = append(response, sessionData)
		}

		ctx.JSON(200, response)
	}
}

// GetStudentSessionByID returns a specific session that a student can view
func GetStudentSessionByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var session models.Session

		// Get session with preloaded teacher and course data
		if err := handler.DB.Preload("Teacher").Preload("Course").First(&session, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		// Transform session data
		status := session.Status
		if !session.EndTime.IsZero() && time.Now().After(session.EndTime) && status != "ongoing" && status != "live" {
			status = "completed"
		}

		sessionData := gin.H{
			"id":           session.ID,
			"title":        session.Title,
			"description":  session.Description,
			"course_id":    session.CourseID,
			"teacher_id":   session.TeacherID,
			"start_time":   session.StartTime,
			"end_time":     session.EndTime,
			"max_students": session.MaxStudents,
			"channel_name": session.ChannelName,
			"status":       status,
			"location":     session.Location,
			"duration":     session.Duration,
			"created_at":   session.CreatedAt,
			"updated_at":   session.UpdatedAt,
		}

		// Add teacher info
		sessionData["teacher_name"] = session.Teacher.Name
		sessionData["teacher_email"] = session.Teacher.Email

		// Add course info if course is linked
		if session.CourseID != nil {
			sessionData["course_name"] = session.Course.Name
			sessionData["course_description"] = session.Course.Description
		}

		ctx.JSON(200, sessionData)
	}
}

// GenerateStudentRTCToken generates Agora RTC token for students to join video sessions
func GenerateStudentRTCToken(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		channelName := ctx.Query("channelName")
		uidStr := ctx.Query("uid")
		roleStr := ctx.Query("role")

		if channelName == "" {
			ctx.JSON(400, gin.H{"error": "Missing parameter: channelName is required"})
			return
		}

		if uidStr == "" {
			uidStr = "0"
		}

		// Default role for students is subscriber
		if roleStr == "" {
			roleStr = "student"
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

		// Generate Agora token
		token, err := utils.GenerateStudentAgoraToken(channelName, uidStr, roleStr)
		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to generate token: " + err.Error()})
			return
		}

		// Automatically record student attendance when joining session
		studentIDToRecord := ctx.Query("studentId")
		if studentIDToRecord == "" && uidStr != "" && uidStr != "0" {
			studentIDToRecord = uidStr
		}
		if studentIDToRecord != "" && session.ID > 0 {
			if sID, sErr := strconv.ParseUint(studentIDToRecord, 10, 32); sErr == nil && sID > 0 {
				attDate := session.StartTime
				if attDate.IsZero() {
					attDate = time.Now()
				}
				att := models.Attendance{
					SessionID: session.ID,
					StudentID: uint(sID),
					Status:    "present",
					Date:      attDate,
				}
				_ = handler.DB.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "session_id"}, {Name: "student_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
				}).Create(&att).Error
			}
		}

		ctx.JSON(200, gin.H{
			"token":       token,
			"appId":       utils.GetAgoraAppID(),
			"channelName": channelName,
			"uid":         uidStr,
		})
	}
}

// CheckProfileCompletionStatus checks if a student's profile is complete
func CheckProfileCompletionStatus(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Define required fields for a complete profile
		missingFields := []string{}
		completionPercentage := 0

		// Check each required field
		if student.Name == "" {
			missingFields = append(missingFields, "name")
		} else {
			completionPercentage += 20
		}

		if student.Email == "" {
			missingFields = append(missingFields, "email")
		} else {
			completionPercentage += 20
		}

		if student.PhoneNumber == "" {
			missingFields = append(missingFields, "phoneNumber")
		} else {
			completionPercentage += 20
		}

		if student.Address == "" {
			missingFields = append(missingFields, "address")
		} else {
			completionPercentage += 20
		}

		if student.ProfilePicture == "" {
			missingFields = append(missingFields, "profilePicture")
		} else {
			completionPercentage += 20
		}

		isComplete := len(missingFields) == 0

		ctx.JSON(200, gin.H{
			"studentId":            student.ID,
			"isComplete":           isComplete,
			"completionPercentage": completionPercentage,
			"missingFields":        missingFields,
			"profile": gin.H{
				"id":             student.ID,
				"name":           student.Name,
				"email":          student.Email,
				"phoneNumber":    student.PhoneNumber,
				"address":        student.Address,
				"profilePicture": student.ProfilePicture,
				"lastLogin":      student.LastLogin,
				"createdAt":      student.CreatedAt,
				"updatedAt":      student.UpdatedAt,
			},
		})
	}
}

// GetAllStudentsProfileStatus gets profile completion status for all students
func GetAllStudentsProfileStatus(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var students []models.Student
		if err := handler.DB.Find(&students).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch students"})
			return
		}

		var response []gin.H
		for _, student := range students {
			// Calculate completion status
			missingFields := []string{}
			completionPercentage := 0

			if student.Name == "" {
				missingFields = append(missingFields, "name")
			} else {
				completionPercentage += 20
			}

			if student.Email == "" {
				missingFields = append(missingFields, "email")
			} else {
				completionPercentage += 20
			}

			if student.PhoneNumber == "" {
				missingFields = append(missingFields, "phoneNumber")
			} else {
				completionPercentage += 20
			}

			if student.Address == "" {
				missingFields = append(missingFields, "address")
			} else {
				completionPercentage += 20
			}

			if student.ProfilePicture == "" {
				missingFields = append(missingFields, "profilePicture")
			} else {
				completionPercentage += 20
			}

			isComplete := len(missingFields) == 0

			studentData := gin.H{
				"id":                   student.ID,
				"name":                 student.Name,
				"email":                student.Email,
				"phoneNumber":          student.PhoneNumber,
				"address":              student.Address,
				"profilePicture":       student.ProfilePicture,
				"lastLogin":            student.LastLogin,
				"createdAt":            student.CreatedAt,
				"updatedAt":            student.UpdatedAt,
				"isComplete":           isComplete,
				"completionPercentage": completionPercentage,
				"missingFields":        missingFields,
			}

			response = append(response, studentData)
		}

		ctx.JSON(200, response)
	}
}

// CheckLoggedInStudentProfile checks if the currently logged-in student's profile is complete
func CheckLoggedInStudentProfile(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get student ID from context (set by middleware)
		studentID, exists := ctx.Get("studentID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Student ID not found in context"})
			return
		}

		var student models.Student
		if err := handler.DB.First(&student, studentID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Define required fields for a complete profile
		missingFields := []string{}
		completionPercentage := 0

		// Check each required field
		if student.Name == "" {
			missingFields = append(missingFields, "name")
		} else {
			completionPercentage += 25
		}

		if student.Email == "" {
			missingFields = append(missingFields, "email")
		} else {
			completionPercentage += 25
		}

		if student.PhoneNumber == "" {
			missingFields = append(missingFields, "phoneNumber")
		} else {
			completionPercentage += 25
		}

		if student.Address == "" {
			missingFields = append(missingFields, "address")
		} else {
			completionPercentage += 25
		}

		isComplete := len(missingFields) == 0

		ctx.JSON(200, gin.H{
			"studentId":            student.ID,
			"isComplete":           isComplete,
			"completionPercentage": completionPercentage,
			"missingFields":        missingFields,
			"profile": gin.H{
				"id":             student.ID,
				"name":           student.Name,
				"email":          student.Email,
				"phoneNumber":    student.PhoneNumber,
				"address":        student.Address,
				"profilePicture": student.ProfilePicture,
				"lastLogin":      student.LastLogin,
				"createdAt":      student.CreatedAt,
				"updatedAt":      student.UpdatedAt,
			},
		})
	}
}

// UpdateLoggedInStudentProfile updates the currently logged-in student's profile
func UpdateLoggedInStudentProfile(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get student ID from context (set by middleware)
		studentID, exists := ctx.Get("studentID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Student ID not found in context"})
			return
		}

		var student models.Student
		if err := handler.DB.First(&student, studentID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Bind the request body to update data
		var updateData struct {
			Name           string `json:"name"`
			PhoneNumber    string `json:"phoneNumber"`
			Phone_Number   string `json:"phone_number"` // Support snake_case
			Address        string `json:"address"`
			ProfilePicture string `json:"profilePicture"`
		}

		if err := ctx.ShouldBindJSON(&updateData); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body"})
			return
		}

		log.Printf("Update data received - Name: %s, PhoneNumber: %s, Phone_Number: %s, Address: %s",
			updateData.Name, updateData.PhoneNumber, updateData.Phone_Number, updateData.Address)

		// Update only the provided fields
		if updateData.Name != "" {
			student.Name = updateData.Name
		}
		// Support both phoneNumber and phone_number
		if updateData.PhoneNumber != "" {
			log.Printf("Setting PhoneNumber from phoneNumber field: %s", updateData.PhoneNumber)
			student.PhoneNumber = updateData.PhoneNumber
		} else if updateData.Phone_Number != "" {
			log.Printf("Setting PhoneNumber from phone_number field: %s", updateData.Phone_Number)
			student.PhoneNumber = updateData.Phone_Number
		}
		if updateData.Address != "" {
			student.Address = updateData.Address
		}
		if updateData.ProfilePicture != "" {
			student.ProfilePicture = updateData.ProfilePicture
		}

		log.Printf("Student object before save - PhoneNumber: %s", student.PhoneNumber)

		if err := handler.DB.Save(&student).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to update student profile"})
			return
		}

		// Return updated profile without sensitive information
		ctx.JSON(200, gin.H{
			"message": "Profile updated successfully",
			"profile": gin.H{
				"id":             student.ID,
				"name":           student.Name,
				"email":          student.Email,
				"phoneNumber":    student.PhoneNumber,
				"address":        student.Address,
				"profilePicture": student.ProfilePicture,
				"lastLogin":      student.LastLogin,
				"createdAt":      student.CreatedAt,
				"updatedAt":      student.UpdatedAt,
			},
		})
	}
}
