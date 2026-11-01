package models

import (
	"time"
)

type Student struct {
	ID                uint      `gorm:"primaryKey"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
	Email             string    `gorm:"uniqueIndex;not null"`
	Name              string    `gorm:"not null"`
	Class             string    `gorm:"default:''"`
	PhoneNumber       string    `gorm:"not null"`
	Address           string    `gorm:"not null"`
	ProfilePicture    string    `gorm:"not null"`
	GoogleAccessToken string    `gorm:"not null"`
	RefreshToken      string    `gorm:"not null"`
	LastLogin         *time.Time
	CurrentStreak     int       `gorm:"default:0"`
	LongestStreak     int       `gorm:"default:0"`
	LastActivityDate  *time.Time
	StripeCustomerID  string       `gorm:"default:''"`
	Enrollments       []Enrollment `gorm:"foreignKey:StudentID"`
}
