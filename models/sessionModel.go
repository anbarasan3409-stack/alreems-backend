package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	Title              string         `json:"title" gorm:"not null"`
	Description        string         `json:"description"`
	TeacherID          uint           `json:"teacher_id" gorm:"not null"`
	Teacher            Teacher        `json:"teacher" gorm:"foreignKey:TeacherID"`
	CourseID           *uint          `json:"course_id"`
	Course             Course         `json:"course" gorm:"foreignKey:CourseID"`
	StartTime          time.Time      `json:"start_time" gorm:"not null"`
	EndTime            time.Time      `json:"end_time" gorm:"not null"`
	Duration           int            `json:"duration"` // Duration in minutes
	Location           string         `json:"location"`
	ChannelName        string         `json:"channel_name" gorm:"uniqueIndex"` // Agora channel name
	WhiteboardUUID     string         `json:"whiteboard_uuid"`                 // Netless whiteboard room UUID
	MaxStudents        int            `json:"max_students" gorm:"default:30"`
	Status             string         `json:"status" gorm:"default:scheduled"`       // scheduled, live, completed, cancelled
	RecordingEnabled   bool           `json:"recording_enabled" gorm:"default:true"` // Enable cloud recording
	RecordingID        string         `json:"recording_id"`                          // Agora cloud recording resource ID
	RecordingSID       string         `json:"recording_sid"`                         // Agora cloud recording session ID
	RecordingURL       string         `json:"recording_url"`                         // URL to recorded video
	RecordingStatus    string         `json:"recording_status"`                      // idle, recording, uploading, uploaded, failed
	RecordingStartedAt *time.Time     `json:"recording_started_at"`                  // When recording started
	RecordingStoppedAt *time.Time     `json:"recording_stopped_at"`                  // When recording stopped
	ReminderSent       bool           `json:"reminder_sent" gorm:"default:false"`    // Whether 30-min reminder email was sent
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}

// SessionResponse represents the session data sent to frontend
type SessionResponse struct {
	ID                 uint   `json:"id"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	TeacherID          uint   `json:"teacher_id"`
	TeacherName        string `json:"teacher_name"`
	CourseID           *uint  `json:"course_id"`
	CourseName         string `json:"course_name"`
	StartTime          string `json:"start_time"`
	EndTime            string `json:"end_time"`
	Duration           int    `json:"duration"`
	Location           string `json:"location"`
	ChannelName        string `json:"channel_name"`
	WhiteboardUUID     string `json:"whiteboard_uuid"`
	MaxStudents        int    `json:"max_students"`
	Status             string `json:"status"`
	RecordingEnabled   bool   `json:"recording_enabled"`
	RecordingURL       string `json:"recording_url"`
	RecordingStatus    string `json:"recording_status"`
	RecordingStartedAt string `json:"recording_started_at,omitempty"`
	RecordingStoppedAt string `json:"recording_stopped_at,omitempty"`
	EnrolledStudents   int    `json:"enrolled_students"`
	StudentsAttended   int    `json:"students_attended"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}
