package models

import (
	"time"
)

type Subscription struct {
	ID                   uint       `gorm:"primaryKey"`
	CreatedAt            time.Time  `gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime"`
	StudentID            uint       `gorm:"not null;index"`
	CourseID             uint       `gorm:"not null;index"`
	StripeSubscriptionID string     `gorm:"uniqueIndex;not null"`
	StripeCustomerID     string     `gorm:"not null"`
	Status               string     `gorm:"not null"` // e.g., active, past_due, canceled, trialing
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
	CancelAtPeriodEnd    bool       `gorm:"default:false"`
	CanceledAt           *time.Time
}
