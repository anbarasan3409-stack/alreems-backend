package controllers

import (
	"backend/database"
	"backend/models"
	"backend/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// StartSessionRecording starts cloud recording for a session
func StartSessionRecording(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")
		id, err := strconv.ParseUint(sessionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		// Get session
		var session models.Session
		if err := handler.DB.Where("id = ?", id).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Check if recording is already in progress
		if session.RecordingStatus == "recording" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recording already in progress"})
			return
		}

		// Check if recording is enabled
		if !session.RecordingEnabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Recording is not enabled for this session"})
			return
		}

		// Generate recording UID (use a large number to avoid conflicts)
		recordingUID := "999999"

		// Step 1: Acquire resource
		resourceID, err := utils.AcquireRecordingResource(session.ChannelName, recordingUID)
		if err != nil {
			log.Printf("Failed to acquire recording resource: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acquire recording resource", "details": err.Error()})
			return
		}

		// Step 2: Generate token for recording bot
		token, err := utils.GenerateAgoraTokenForRecording(session.ChannelName, recordingUID)
		if err != nil {
			log.Printf("Failed to generate recording token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recording token", "details": err.Error()})
			return
		}

		// Step 3: Start recording
		sid, err := utils.StartCloudRecording(session.ChannelName, recordingUID, resourceID, token)
		if err != nil {
			log.Printf("Failed to start recording: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start recording", "details": err.Error()})
			return
		}

		// Update session with recording info
		now := time.Now()
		session.RecordingID = resourceID
		session.RecordingSID = sid
		session.RecordingStatus = "recording"
		session.RecordingStartedAt = &now

		if err := handler.DB.Save(&session).Error; err != nil {
			log.Printf("Failed to update session with recording info: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		log.Printf("Recording started for session %d: resourceID=%s, sid=%s", id, resourceID, sid)

		c.JSON(http.StatusOK, gin.H{
			"message":     "Recording started successfully",
			"resource_id": resourceID,
			"sid":         sid,
			"status":      "recording",
		})
	}
}

// StopSessionRecording stops cloud recording for a session
func StopSessionRecording(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")
		id, err := strconv.ParseUint(sessionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		// Get session
		var session models.Session
		if err := handler.DB.Where("id = ?", id).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Check if recording is in progress
		if session.RecordingStatus != "recording" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No recording in progress"})
			return
		}

		// Generate recording UID (must match the one used in start)
		recordingUID := "999999"

		// Stop recording
		err = utils.StopCloudRecording(session.ChannelName, recordingUID, session.RecordingID, session.RecordingSID)
		if err != nil {
			log.Printf("Failed to stop recording: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop recording", "details": err.Error()})
			return
		}

		// Update session
		now := time.Now()
		session.RecordingStatus = "uploading"
		session.RecordingStoppedAt = &now

		if err := handler.DB.Save(&session).Error; err != nil {
			log.Printf("Failed to update session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		log.Printf("Recording stopped for session %d", id)

		c.JSON(http.StatusOK, gin.H{
			"message": "Recording stopped successfully",
			"status":  "uploading",
		})
	}
}

// QuerySessionRecording queries the status of a recording
func QuerySessionRecording(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")
		id, err := strconv.ParseUint(sessionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		// Get session
		var session models.Session
		if err := handler.DB.Where("id = ?", id).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Check if recording exists
		if session.RecordingID == "" || session.RecordingSID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No recording found for this session"})
			return
		}

		// Query recording status
		queryResp, err := utils.QueryRecordingStatus(session.RecordingID, session.RecordingSID)
		if err != nil {
			log.Printf("Failed to query recording: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query recording", "details": err.Error()})
			return
		}

		// Extract file list if available
		var fileList []interface{}
		if serverResponse, ok := queryResp.ServerResponse["fileList"]; ok {
			if files, ok := serverResponse.([]interface{}); ok {
				fileList = files
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"resource_id":     queryResp.ResourceID,
			"sid":             queryResp.SID,
			"server_response": queryResp.ServerResponse,
			"file_list":       fileList,
			"status":          session.RecordingStatus,
		})
	}
}

// UpdateRecordingURL manually updates the recording URL (admin only)
func UpdateRecordingURL(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")
		id, err := strconv.ParseUint(sessionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		var request struct {
			RecordingURL string `json:"recording_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
			return
		}

		// Get session
		var session models.Session
		if err := handler.DB.Where("id = ?", id).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		// Update recording URL
		session.RecordingURL = request.RecordingURL
		session.RecordingStatus = "uploaded"

		if err := handler.DB.Save(&session).Error; err != nil {
			log.Printf("Failed to update session: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update session"})
			return
		}

		log.Printf("Recording URL updated for session %d", id)

		c.JSON(http.StatusOK, gin.H{
			"message":       "Recording URL updated successfully",
			"recording_url": request.RecordingURL,
		})
	}
}

// GetSessionRecording gets recording info for a session
func GetSessionRecording(handler database.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("id")
		id, err := strconv.ParseUint(sessionID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
			return
		}

		// Get session
		var session models.Session
		if err := handler.DB.Where("id = ?", id).First(&session).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
			return
		}

		var startedAt, stoppedAt string
		if session.RecordingStartedAt != nil {
			startedAt = session.RecordingStartedAt.Format(time.RFC3339)
		}
		if session.RecordingStoppedAt != nil {
			stoppedAt = session.RecordingStoppedAt.Format(time.RFC3339)
		}

		c.JSON(http.StatusOK, gin.H{
			"recording_enabled": session.RecordingEnabled,
			"recording_url":     session.RecordingURL,
			"recording_status":  session.RecordingStatus,
			"recording_id":      session.RecordingID,
			"recording_sid":     session.RecordingSID,
			"started_at":        startedAt,
			"stopped_at":        stoppedAt,
		})
	}
}
