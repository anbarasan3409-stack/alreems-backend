package models

import (
	"time"

	"gorm.io/gorm"
)

type Enrollment struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	StudentID   uint           `json:"student_id" gorm:"not null"`
	Student     Student        `json:"student" gorm:"foreignKey:StudentID"`
	CourseID    uint           `json:"course_id" gorm:"not null"`
	Course      Course         `json:"course" gorm:"foreignKey:CourseID"`
	EnrolledAt  time.Time      `json:"enrolled_at" gorm:"autoCreateTime"`
	Progress    float32        `json:"progress" gorm:"default:0"`    // Percentage (0-100)
	Status      string         `json:"status" gorm:"default:active"` // active, completed, dropped
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// EnrollmentResponse represents the enrollment data sent to frontend
type EnrollmentResponse struct {
	ID          uint    `json:"id"`
	StudentID   uint    `json:"student_id"`
	CourseID    uint    `json:"course_id"`
	CourseName  string  `json:"course_name"`
	CourseCode  string  `json:"course_code"`
	Description string  `json:"description"`
	Credits     int     `json:"credits"`
	TeacherID   uint    `json:"teacher_id"`
	Progress    float32 `json:"progress"`
	Status      string  `json:"status"`
	EnrolledAt  string  `json:"enrolled_at"`
	CompletedAt string  `json:"completed_at,omitempty"`
}
