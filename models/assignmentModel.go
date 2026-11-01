package models

import (
	"time"

	"gorm.io/gorm"
)

type Assignment struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	DueDate     time.Time      `json:"due_date"`
	CourseID    uint           `gorm:"index;not null" json:"course_id"`
	TeacherID   uint           `gorm:"index;not null" json:"teacher_id"`
	TotalMarks  int            `json:"total_marks"`
	
	Course      Course         `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Teacher     Teacher        `gorm:"foreignKey:TeacherID" json:"teacher,omitempty"`
}

type AssignmentSubmission struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	AssignmentID uint           `gorm:"index;uniqueIndex:idx_assignment_student;not null" json:"assignment_id"`
	StudentID    uint           `gorm:"index;uniqueIndex:idx_assignment_student;not null" json:"student_id"`
	FileURL      string         `json:"file_url"`
	Status       string         `gorm:"default:'submitted'" json:"status"` // submitted, graded, late
	MarksObtained int           `json:"marks_obtained"`
	Feedback     string         `json:"feedback"`
	
	Assignment   Assignment     `gorm:"foreignKey:AssignmentID" json:"assignment,omitempty"`
	Student      Student        `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}
