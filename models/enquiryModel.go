package models

import (
	"time"

	"gorm.io/gorm"
)

type Enquiry struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Name             string         `gorm:"not null" json:"name"`
	Email            string         `gorm:"not null" json:"email"`
	Phone            string         `json:"phone"`
	Subject          string         `json:"subject"`
	Message          string         `gorm:"not null" json:"message"`
	Status           string         `gorm:"default:'new'" json:"status"` // new, in-progress, resolved, closed
	CourseOfInterest string         `json:"course_of_interest"`
}
