package controllers

import (
	"backend/database"
	"backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

type AttendanceInput struct {
	StudentID uint   `json:"student_id" binding:"required"`
	Status    string `json:"status" binding:"required"` // present, absent, late
}

type BulkAttendanceInput struct {
	SessionID   uint              `json:"session_id" binding:"required"`
	Attendances []AttendanceInput `json:"attendances" binding:"required"`
}

func MarkAttendance(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var markerID *uint
		if user, exists := ctx.Get("user"); exists {
			if uMap, ok := user.(map[string]interface{}); ok {
				if idFloat, ok := uMap["id"].(float64); ok {
					id := uint(idFloat)
					markerID = &id
				}
			}
		}

		var input BulkAttendanceInput
		if err := ctx.ShouldBindJSON(&input); err != nil {
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}

		var session models.Session
		if err := handler.DB.First(&session, input.SessionID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		var attendanceRecords []models.Attendance
		for _, att := range input.Attendances {
			attendanceRecords = append(attendanceRecords, models.Attendance{
				SessionID: input.SessionID,
				StudentID: att.StudentID,
				Status:    att.Status,
				MarkedBy:  markerID,
				Date:      session.StartTime,
			})
		}

		if len(attendanceRecords) == 0 {
			ctx.JSON(400, gin.H{"error": "No attendance records provided"})
			return
		}

		err := handler.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "student_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "marked_by", "updated_at"}),
		}).Create(&attendanceRecords).Error

		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to mark attendance"})
			return
		}

		ctx.JSON(200, gin.H{"message": "Attendance marked successfully"})
	}
}

// StudentSelfMarkAttendance records a student's attendance when they join a session
func StudentSelfMarkAttendance(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req struct {
			SessionID uint `json:"session_id" binding:"required"`
			StudentID uint `json:"student_id" binding:"required"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request: " + err.Error()})
			return
		}

		var session models.Session
		if err := handler.DB.First(&session, req.SessionID).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Session not found"})
			return
		}

		attDate := session.StartTime
		if attDate.IsZero() {
			attDate = time.Now()
		}

		attendance := models.Attendance{
			SessionID: req.SessionID,
			StudentID: req.StudentID,
			Status:    "present",
			Date:      attDate,
		}

		err := handler.DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "session_id"}, {Name: "student_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
		}).Create(&attendance).Error

		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to mark attendance: " + err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "Attendance recorded successfully"})
	}
}

func GetSessionAttendance(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sessionID := ctx.Param("sessionId")
		var attendances []models.Attendance

		if err := handler.DB.Preload("Student").Where("session_id = ?", sessionID).Find(&attendances).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch attendance"})
			return
		}

		ctx.JSON(200, attendances)
	}
}

func GetStudentAttendance(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		studentID := ctx.Param("studentId")
		if studentID == "" {
			studentID = ctx.Param("id")
		}
		
		// If accessing as student, ensure they only get their own or we just use context ID.
		// For simplicity, we assume routes are protected by middleware and param validation.
		if studentID == "" {
			if user, exists := ctx.Get("user"); exists {
				if uMap, ok := user.(map[string]interface{}); ok {
					if idFloat, ok := uMap["id"].(float64); ok {
						// Convert to string or use directly
						var attendances []models.Attendance
						if err := handler.DB.Preload("Session").Where("student_id = ?", uint(idFloat)).Order("date desc").Find(&attendances).Error; err != nil {
							ctx.JSON(500, gin.H{"error": "Failed to fetch attendance"})
							return
						}
						ctx.JSON(200, attendances)
						return
					}
				}
			}
			ctx.JSON(400, gin.H{"error": "Student ID required"})
			return
		}

		var attendances []models.Attendance
		if err := handler.DB.Preload("Session").Preload("Session.Course").Where("student_id = ?", studentID).Order("date desc").Find(&attendances).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch attendance"})
			return
		}

		ctx.JSON(200, attendances)
	}
}
