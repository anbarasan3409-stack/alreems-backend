package models

import (
	"time"

	"gorm.io/gorm"
)

type StudyActivity struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	StudentID      uint           `json:"student_id" gorm:"not null;uniqueIndex:idx_student_date"`
	Student        Student        `json:"-" gorm:"foreignKey:StudentID"`
	Date           string         `json:"date" gorm:"not null;uniqueIndex:idx_student_date"` // Format: YYYY-MM-DD
	MinutesStudied int            `json:"minutes_studied" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
