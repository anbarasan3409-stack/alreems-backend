package models

import (
	"time"
)

type Course struct {
	ID          uint         `gorm:"primaryKey"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
	Name        string       `gorm:"not null"`
	Code        string       `gorm:"uniqueIndex;not null"`
	Description string       `gorm:"type:text"`
	Credits     int          `gorm:"not null"`
	Price       float64      `gorm:"default:0" json:"price"`
	TeacherID   uint         `gorm:"not null"`
	Teacher     Teacher      `gorm:"foreignKey:TeacherID"`
	Enrollments []Enrollment `gorm:"foreignKey:CourseID"`
}
