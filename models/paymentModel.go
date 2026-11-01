package models

import (
	"time"
)

type Payment struct {
	ID            uint       `gorm:"primaryKey" json:"ID"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	StudentID     uint       `gorm:"not null" json:"studentID"`
	CourseID      uint       `gorm:"not null" json:"courseID"`
	Amount        float64    `gorm:"not null" json:"amount"`
	Status        string     `gorm:"not null" json:"status"`
	PaymentDate   *time.Time `json:"paymentDate"`
	Method        string     `gorm:"not null" json:"method"`
	TransactionID string     `gorm:"uniqueIndex" json:"transactionID"`
	Student       *Student   `gorm:"foreignKey:StudentID" json:"student,omitempty"`
	Course        *Course    `gorm:"foreignKey:CourseID" json:"course,omitempty"`
}
