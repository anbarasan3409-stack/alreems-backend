package database

import (
	"log"
	"os"

	"backend/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init() *gorm.DB {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	log.Println("Connecting to database...")
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatalln("Failed to connect to database:", err)
	}
	log.Println("Database connected successfully")
	// Auto-migrate tables
	if err := db.AutoMigrate(
		&models.Admin{},
		&models.Teacher{},
		&models.Student{},
		&models.Course{},
		&models.Session{},
		&models.Enrollment{},
		&models.Subscription{},
		&models.Payment{},
		&models.PaymentReminderSettings{},
		&models.Feedback{},
		&models.Enquiry{},
		&models.Assignment{},
		&models.Attendance{},
		&models.LearningMaterial{},
		&models.StudyActivity{},
	); err != nil {
		log.Println("Warning: Failed to migrate tables:", err)
	}

	// Ensure sessions table has required columns and constraints
	_ = db.Exec("ALTER TABLE sessions ADD COLUMN IF NOT EXISTS reminder_sent BOOLEAN DEFAULT false;").Error
	_ = db.Exec("ALTER TABLE sessions ALTER COLUMN course_id DROP NOT NULL;").Error

	return db
}
