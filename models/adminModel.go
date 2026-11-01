package models

import (
	"time"
)

type Admin struct {
	ID                uint      `gorm:"primaryKey"`
	CreatedAt         time.Time `gorm:"autoCreateTime"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime"`
	Email             string    `gorm:"uniqueIndex;not null"`
	Name              string    `gorm:"not null"`
	PhoneNumber       string    `gorm:"not null"`
	Address           string    `gorm:"not null"`
	ProfilePicture    string    `gorm:"not null"`
	Role              string    `gorm:"default:'admin';not null"`
	GoogleAccessToken string    `gorm:"not null"`
	RefreshToken      string    `gorm:"not null"`
	LastLogin         *time.Time
}
