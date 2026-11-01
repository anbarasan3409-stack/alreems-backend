package models

import (
	"time"

	"gorm.io/gorm"
)

type LearningMaterial struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	FileType    string         `json:"file_type"` // e.g., pdf, video, link, document
	FileURL     string         `gorm:"not null" json:"file_url"` // Path or external link
	CourseID    uint           `gorm:"index;not null" json:"course_id"`
	TeacherID   uint           `gorm:"index;not null" json:"teacher_id"`
	
	Course      Course         `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Teacher     Teacher        `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
}
