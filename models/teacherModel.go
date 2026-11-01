package models

import (
	"time"
)

type Teacher struct {
	ID                uint      `gorm:"primaryKey"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
	Email             string    `gorm:"uniqueIndex;not null"`
	Name              string    `gorm:"not null"`
	PhoneNumber       string    `gorm:"not null"`
	Address           string    `gorm:"not null"`
	ProfilePicture    string    `gorm:"not null"`
	Department        string    `gorm:"not null"`
	Designation       string    `gorm:"not null"`
	Subject           string    `gorm:"not null"`
	Qualifications    string    `gorm:"not null"`
	Experience        string    `gorm:"not null"`
	JoiningDate       *time.Time
	Status            string `gorm:"default:active"`
	GoogleAccessToken string `gorm:"not null"`
	RefreshToken      string `gorm:"not null"`
	LastLogin         *time.Time
}
