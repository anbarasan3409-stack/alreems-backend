package models

import (
	"time"
)

// PaymentReminderSettings stores global system settings and payment reminder scheduler settings
type PaymentReminderSettings struct {
	ID                         uint       `gorm:"primaryKey" json:"id"`
	CreatedAt                  time.Time  `json:"createdAt"`
	UpdatedAt                  time.Time  `json:"updatedAt"`

	// General Registration Settings
	StudentRegistrationEnabled bool       `gorm:"default:true" json:"studentRegistrationEnabled"`
	TeacherRegistrationEnabled bool       `gorm:"default:true" json:"teacherRegistrationEnabled"`

	// Payment Reminder Scheduler Settings
	Enabled                    bool       `gorm:"default:false" json:"enabled"`
	ScheduleType               string     `gorm:"default:'monthly'" json:"scheduleType"` // "monthly" or "specific_date"
	DayOfMonth                 int        `gorm:"default:1" json:"dayOfMonth"`          // 1 - 31
	SpecificDate               string     `gorm:"default:''" json:"specificDate"`        // YYYY-MM-DD
	SendHour                   int        `gorm:"default:9" json:"sendHour"`             // 0 - 23 (e.g. 9 for 09:00 AM)
	SendMinute                 int        `gorm:"default:0" json:"sendMinute"`           // 0 - 59
	EmailSubject               string     `gorm:"type:varchar(255);default:'Payment Reminder: Upcoming Course Tuition Fee - Al Reems Academy'" json:"emailSubject"`
	CustomMessage              string     `gorm:"type:text" json:"customMessage"`

	// Execution tracking
	LastRunAt                  *time.Time `json:"lastRunAt"`
	LastRunStatus              string     `gorm:"type:text;default:''" json:"lastRunStatus"`
}
