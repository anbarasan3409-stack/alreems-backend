package models

import (
	"time"

	"gorm.io/gorm"
)

type Feedback struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	StudentID   uint           `json:"student_id"`
	StudentName string         `json:"student_name"`
	Course      string         `json:"course"`
	Rating      int            `gorm:"check:rating >= 1 AND rating <= 5" json:"rating"` // 1-5 stars
	Message     string         `gorm:"not null" json:"message"`
	Featured    bool           `gorm:"default:false" json:"featured"` // Admin can feature it as testimonial
}
