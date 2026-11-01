package database

import (
	"fmt"
	"log"
	"time"

	"backend/models"
	"gorm.io/gorm"
)

// SeedStudentEnrollments enrolls alreemsacademy@gmail.com and sanjaiijee@gmail.com in VED-101 and creates payment records
func SeedStudentEnrollments(db *gorm.DB) {
	log.Println("--- Seeding Student Enrollments and Payments for VED-101 ---")

	// 1. Locate Course VED-101
	var course models.Course
	if err := db.Where("code = ? OR LOWER(name) = ?", "VED-101", "vedic maths").First(&course).Error; err != nil {
		log.Printf("Course VED-101 not found: %v\n", err)
		return
	}

	price := course.Price
	if price <= 0 {
		price = 100.0
	}

	studentsToEnroll := []struct {
		Email string
		Name  string
		Phone string
	}{
		{Email: "alreemsacademy@gmail.com", Name: "Al Reems Academy", Phone: "+971500000000"},
		{Email: "sanjaiijee@gmail.com", Name: "Sanjai", Phone: "+971500000001"},
	}

	for _, sData := range studentsToEnroll {
		// Find or create student
		var student models.Student
		if err := db.Where("email = ?", sData.Email).First(&student).Error; err != nil {
			student = models.Student{
				Email:          sData.Email,
				Name:           sData.Name,
				PhoneNumber:    sData.Phone,
				Address:        "United Arab Emirates",
				Class:          "Active",
				ProfilePicture: "",
			}
			if err := db.Create(&student).Error; err != nil {
				log.Printf("Failed to create student %s: %v\n", sData.Email, err)
				continue
			}
			log.Printf("Created student: %s (ID: %d)\n", student.Email, student.ID)
		} else {
			log.Printf("Student %s found (ID: %d)\n", student.Email, student.ID)
		}

		// Ensure enrollment exists in VED-101
		var enrollment models.Enrollment
		if err := db.Where("student_id = ? AND course_id = ?", student.ID, course.ID).First(&enrollment).Error; err != nil {
			enrollment = models.Enrollment{
				StudentID: student.ID,
				CourseID:  course.ID,
				Status:    "active",
				Progress:  0,
			}
			if err := db.Create(&enrollment).Error; err != nil {
				log.Printf("Failed to enroll student %s in %s: %v\n", student.Email, course.Code, err)
			} else {
				log.Printf("Enrolled student %s in %s successfully\n", student.Email, course.Code)
			}
		} else {
			log.Printf("Student %s is already enrolled in %s\n", student.Email, course.Code)
		}

		// Ensure payment record exists for this course
		var payment models.Payment
		if err := db.Where("student_id = ? AND course_id = ?", student.ID, course.ID).First(&payment).Error; err != nil {
			now := time.Now()
			payment = models.Payment{
				StudentID:     student.ID,
				CourseID:      course.ID,
				Amount:        price,
				Status:        "completed",
				PaymentDate:   &now,
				Method:        "Stripe",
				TransactionID: fmt.Sprintf("txn_ved_%d_%d", student.ID, now.Unix()),
			}
			if err := db.Create(&payment).Error; err != nil {
				log.Printf("Failed to create payment record for student %s: %v\n", student.Email, err)
			} else {
				log.Printf("Payment record created for student %s: %.2f AED (Status: completed)\n", student.Email, price)
			}
		} else {
			log.Printf("Payment record already exists for student %s in %s\n", student.Email, course.Code)
		}
	}
	log.Println("--- VED-101 Student Enrollment & Payment Seeding Complete ---")
}
