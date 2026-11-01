package models

import (
	"time"

	"gorm.io/gorm"
)

type Attendance struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	SessionID uint           `json:"session_id" gorm:"not null;uniqueIndex:idx_session_student"`
	Session   Session        `json:"-" gorm:"foreignKey:SessionID"`
	StudentID uint           `json:"student_id" gorm:"not null;uniqueIndex:idx_session_student"`
	Student   Student        `json:"student" gorm:"foreignKey:StudentID"`
	Status    string         `json:"status" gorm:"not null;default:'absent'"` // present, absent, late
	MarkedBy  *uint          `json:"marked_by"` // Teacher or Admin who marked it
	Date      time.Time      `json:"date" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
