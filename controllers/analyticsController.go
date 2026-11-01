package controllers

import (
	"backend/database"
	"backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetStudentAnalytics returns analytics like streak and total study hours
func GetStudentAnalytics(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		
		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Calculate total study hours based on StudyActivity
		var totalMinutes int64
		handler.DB.Model(&models.StudyActivity{}).Where("student_id = ?", student.ID).Select("COALESCE(SUM(minutes_studied), 0)").Scan(&totalMinutes)

		// Check if streak is still valid (if LastActivityDate was before yesterday, reset current streak)
		currentStreak := student.CurrentStreak
		if student.LastActivityDate != nil {
			now := time.Now()
			lastActivity := *student.LastActivityDate
			
			// Get start of today and start of yesterday for comparison
			startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			startOfYesterday := startOfToday.AddDate(0, 0, -1)
			
			// If last activity is strictly before yesterday's start, streak is broken
			if lastActivity.Before(startOfYesterday) {
				currentStreak = 0
			}
		}

		ctx.JSON(200, gin.H{
			"total_study_hours": float64(totalMinutes) / 60.0,
			"current_streak":    currentStreak,
			"longest_streak":    student.LongestStreak,
		})
	}
}

// GetStudentWeeklyActivity returns the last 7 days of study activity
func GetStudentWeeklyActivity(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		
		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		// Get last 7 days of activity
		var activities []models.StudyActivity
		
		// Determine the start date (7 days ago)
		now := time.Now()
		sevenDaysAgo := now.AddDate(0, 0, -6)
		startDateStr := sevenDaysAgo.Format("2006-01-02")
		
		handler.DB.Where("student_id = ? AND date >= ?", student.ID, startDateStr).Find(&activities)
		
		// Map activities by date for easy lookup
		activityMap := make(map[string]int)
		for _, act := range activities {
			activityMap[act.Date] = act.MinutesStudied
		}
		
		// Construct the array of 7 days
		var weeklyActivity []gin.H
		
		for i := 6; i >= 0; i-- {
			date := now.AddDate(0, 0, -i)
			dateStr := date.Format("2006-01-02")
			
			minutes := 0
			if val, exists := activityMap[dateStr]; exists {
				minutes = val
			}
			
			weeklyActivity = append(weeklyActivity, gin.H{
				"day":      date.Format("Mon"),
				"date":     dateStr,
				"minutes":  minutes,
				"hours":    float64(minutes) / 60.0,
			})
		}

		ctx.JSON(200, weeklyActivity)
	}
}

// LogStudyTime increments study minutes and updates the student's streak
func LogStudyTime(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		
		var req struct {
			Minutes int `json:"minutes"`
		}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid request body. Expected 'minutes'"})
			return
		}
		
		if req.Minutes <= 0 || req.Minutes > 300 {
			// Ignore invalid or unreasonably high single payloads
			ctx.JSON(400, gin.H{"error": "Invalid minutes amount"})
			return
		}

		var student models.Student
		if err := handler.DB.First(&student, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Student not found"})
			return
		}

		now := time.Now()
		todayStr := now.Format("2006-01-02")
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		startOfYesterday := startOfToday.AddDate(0, 0, -1)

		// Use a transaction to ensure data consistency
		err := handler.DB.Transaction(func(tx *gorm.DB) error {
			// 1. Update or create today's StudyActivity
			var activity models.StudyActivity
			err := tx.Where("student_id = ? AND date = ?", student.ID, todayStr).First(&activity).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					activity = models.StudyActivity{
						StudentID:      student.ID,
						Date:           todayStr,
						MinutesStudied: req.Minutes,
					}
					if err := tx.Create(&activity).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				activity.MinutesStudied += req.Minutes
				if err := tx.Save(&activity).Error; err != nil {
					return err
				}
			}

			// 2. Update Streak logic on Student model
			
			if student.LastActivityDate == nil {
				student.CurrentStreak = 1
				if student.CurrentStreak > student.LongestStreak {
					student.LongestStreak = student.CurrentStreak
				}
			} else {
				lastActivity := *student.LastActivityDate
				
				// If last activity was strictly before yesterday, streak reset
				if lastActivity.Before(startOfYesterday) {
					student.CurrentStreak = 1
				} else if lastActivity.Before(startOfToday) {
					// Last activity was yesterday, increment streak
					student.CurrentStreak += 1
					if student.CurrentStreak > student.LongestStreak {
						student.LongestStreak = student.CurrentStreak
					}
				}
				// If last activity is today, streak stays the same. 
			}

			// Always update LastActivityDate when logging time
			student.LastActivityDate = &now
			
			if err := tx.Save(&student).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to log study time"})
			return
		}

		ctx.JSON(200, gin.H{
			"message": "Study time logged successfully",
			"current_streak": student.CurrentStreak,
		})
	}
}
