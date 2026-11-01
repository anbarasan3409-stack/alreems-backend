package controllers

import (
	"net/http"
	"strconv"
	"time"

	"backend/database"
	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

// Helper function to create session response for teachers
func createTeacherSessionResponse(session models.Session) models.SessionResponse {
	status := session.Status
	if !session.EndTime.IsZero() && time.Now().After(session.EndTime) && status != "ongoing" && status != "live" {
		status = "completed"
	}

	sessionResponse := models.SessionResponse{
		ID:          session.ID,
		Title:       session.Title,
		Description: session.Description,
		TeacherID:   session.TeacherID,
		CourseID:    session.CourseID,
		Duration:    session.Duration,
		Location:    session.Location,
		ChannelName: session.ChannelName,
		MaxStudents: session.MaxStudents,
		Status:      status,
		StartTime:   session.StartTime.Format("2006-01-02T15:04:05Z07:00"),
		EndTime:     session.EndTime.Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if session.Teacher.ID != 0 {
		sessionResponse.TeacherName = session.Teacher.Name
	}

	if session.Course.ID != 0 {
		sessionResponse.CourseName = session.Course.Name
	}

	return sessionResponse
}

// GetTeacherSessions gets all sessions for a specific teacher
func GetTeacherSessions(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherIDInterface, exists := c.Get("teacherID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Teacher not authenticated",
			})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid teacher ID",
			})
			return
		}

		var sessions []models.Session
		result := handler.DB.Where("teacher_id = ?", teacherID).
			Preload("Teacher").
			Preload("Course").
			Order("start_time DESC").
			Find(&sessions)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch sessions",
			})
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
		var sessionResponses []models.SessionResponse
		for _, session := range sessions {
			sessionResponse := createTeacherSessionResponse(session)
			if count, ok := attendanceMap[session.ID]; ok {
				sessionResponse.StudentsAttended = count
			}
			if session.CourseID != nil {
				if count, ok := enrollmentMap[*session.CourseID]; ok {
					sessionResponse.EnrolledStudents = count
					if sessionResponse.StudentsAttended == 0 && count > 0 {
						sessionResponse.StudentsAttended = count
					}
				}
			}
			sessionResponses = append(sessionResponses, sessionResponse)
		}

		c.JSON(http.StatusOK, gin.H{
			"sessions": sessionResponses,
		})
	}
}

// GetTeacherSessionByID gets a specific session by ID for a teacher
func GetTeacherSessionByID(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherIDInterface, exists := c.Get("teacherID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Teacher not authenticated",
			})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid teacher ID",
			})
			return
		}

		sessionID := c.Param("id")
		id, err := strconv.Atoi(sessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid session ID",
			})
			return
		}

		var session models.Session
		result := handler.DB.Where("id = ? AND teacher_id = ?", id, teacherID).
			Preload("Teacher").
			Preload("Course").
			First(&session)

		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
			return
		}

		var attendanceCount int64
		handler.DB.Model(&models.Attendance{}).Where("session_id = ? AND status IN ('present', 'late')", session.ID).Distinct("student_id").Count(&attendanceCount)

		var enrollmentCount int64
		if session.CourseID != nil {
			handler.DB.Model(&models.Enrollment{}).Where("course_id = ? AND (LOWER(COALESCE(status, 'active')) = 'active' OR status = '' OR status IS NULL)", *session.CourseID).Distinct("student_id").Count(&enrollmentCount)
		}

		sessionResponse := createTeacherSessionResponse(session)
		sessionResponse.StudentsAttended = int(attendanceCount)
		sessionResponse.EnrolledStudents = int(enrollmentCount)
		if sessionResponse.StudentsAttended == 0 && enrollmentCount > 0 {
			sessionResponse.StudentsAttended = int(enrollmentCount)
		}

		c.JSON(http.StatusOK, gin.H{
			"session": sessionResponse,
		})
	}
}

// CreateTeacherSession creates a new session for a teacher
func CreateTeacherSession(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherIDInterface, exists := c.Get("teacherID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Teacher not authenticated",
			})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid teacher ID",
			})
			return
		}

		var req struct {
			Title       string    `json:"title" binding:"required"`
			Description string    `json:"description"`
			CourseID    *uint     `json:"course_id"`
			StartTime   time.Time `json:"start_time" binding:"required"`
			EndTime     time.Time `json:"end_time" binding:"required"`
			Duration    int       `json:"duration" binding:"required"`
			Location    string    `json:"location"`
			ChannelName string    `json:"channel_name"`
			MaxStudents int       `json:"max_students"`
			Status      string    `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Validate that the end time is after start time
		if req.EndTime.Before(req.StartTime) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "End time must be after start time",
			})
			return
		}

		// If course_id is provided, verify it belongs to the teacher
		if req.CourseID != nil {
			var course models.Course
			result := handler.DB.Where("id = ? AND teacher_id = ?", *req.CourseID, teacherID).First(&course)
			if result.Error != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Course not found or does not belong to teacher",
				})
				return
			}
		}

		// Generate unique channel name if not provided
		if req.ChannelName == "" {
			req.ChannelName = "session_" + strconv.FormatInt(time.Now().Unix(), 10)
		}

		// Set default max students if not provided
		if req.MaxStudents == 0 {
			req.MaxStudents = 30
		}

		// Set default status if not provided
		status := req.Status
		if status == "" {
			status = "scheduled"
		}

		session := models.Session{
			Title:       req.Title,
			Description: req.Description,
			TeacherID:   teacherID,
			CourseID:    req.CourseID,
			StartTime:   req.StartTime,
			EndTime:     req.EndTime,
			Duration:    req.Duration,
			Location:    req.Location,
			ChannelName: req.ChannelName,
			MaxStudents: req.MaxStudents,
			Status:      status,
		}

		result := handler.DB.Create(&session)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create session",
			})
			return
		}

		// Return the created session with preloaded data
		handler.DB.Preload("Teacher").Preload("Course").First(&session, session.ID)

		sessionResponse := createTeacherSessionResponse(session)

		c.JSON(http.StatusCreated, gin.H{
			"message": "Session created successfully",
			"session": sessionResponse,
		})

		// Trigger instant class reminder email to teacher and enrolled students if it's a live/instant session
		isInstant := session.Status == "live" || session.StartTime.Before(time.Now().Add(5*time.Minute))
		if isInstant {
			go utils.SendClassReminders(handler.DB, session, true)
		}
	}
}

// UpdateTeacherSession updates an existing session for a teacher
func UpdateTeacherSession(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherIDInterface, exists := c.Get("teacherID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Teacher not authenticated",
			})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid teacher ID",
			})
			return
		}

		sessionID := c.Param("id")
		id, err := strconv.Atoi(sessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid session ID",
			})
			return
		}

		var req struct {
			Title       string    `json:"title"`
			Description string    `json:"description"`
			CourseID    *uint     `json:"course_id"`
			StartTime   time.Time `json:"start_time"`
			EndTime     time.Time `json:"end_time"`
			Duration    int       `json:"duration"`
			Location    string    `json:"location"`
			ChannelName string    `json:"channel_name"`
			MaxStudents int       `json:"max_students"`
			Status      string    `json:"status"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		// Find the session
		var session models.Session
		result := handler.DB.Where("id = ? AND teacher_id = ?", id, teacherID).First(&session)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
			return
		}

		// Validate that the end time is after start time if both are provided
		if !req.StartTime.IsZero() && !req.EndTime.IsZero() && req.EndTime.Before(req.StartTime) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "End time must be after start time",
			})
			return
		}

		// If course_id is provided, verify it belongs to the teacher
		if req.CourseID != nil {
			var course models.Course
			result := handler.DB.Where("id = ? AND teacher_id = ?", *req.CourseID, teacherID).First(&course)
			if result.Error != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Course not found or does not belong to teacher",
				})
				return
			}
		}

		// Update fields if provided
		if req.Title != "" {
			session.Title = req.Title
		}
		if req.Description != "" {
			session.Description = req.Description
		}
		if req.CourseID != nil {
			session.CourseID = req.CourseID
		}
		if !req.StartTime.IsZero() {
			session.StartTime = req.StartTime
		}
		if !req.EndTime.IsZero() {
			session.EndTime = req.EndTime
		}
		if req.Duration > 0 {
			session.Duration = req.Duration
		}
		if req.Location != "" {
			session.Location = req.Location
		}
		if req.ChannelName != "" {
			session.ChannelName = req.ChannelName
		}
		if req.MaxStudents > 0 {
			session.MaxStudents = req.MaxStudents
		}
		if req.Status != "" {
			session.Status = req.Status
		}

		result = handler.DB.Save(&session)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update session",
			})
			return
		}

		// Return the updated session with preloaded data
		handler.DB.Preload("Teacher").Preload("Course").First(&session, session.ID)

		sessionResponse := createTeacherSessionResponse(session)

		c.JSON(http.StatusOK, gin.H{
			"message": "Session updated successfully",
			"session": sessionResponse,
		})

		// Trigger instant class reminder email to teacher and students if status was updated to live
		if session.Status == "live" {
			go utils.SendClassReminders(handler.DB, session, true)
		}
	}
}

// DeleteTeacherSession deletes a session for a teacher
func DeleteTeacherSession(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		teacherIDInterface, exists := c.Get("teacherID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Teacher not authenticated",
			})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid teacher ID",
			})
			return
		}

		sessionID := c.Param("id")
		id, err := strconv.Atoi(sessionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid session ID",
			})
			return
		}

		// Find the session
		var session models.Session
		result := handler.DB.Where("id = ? AND teacher_id = ?", id, teacherID).First(&session)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Session not found",
			})
			return
		}

		// Delete the session
		result = handler.DB.Delete(&session)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete session",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Session deleted successfully",
		})
	}
}
