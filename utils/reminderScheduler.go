package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"backend/models"
	"gorm.io/gorm"
)

// BuildPaymentReminderHTML generates a clean, responsive HTML email for payment reminders
func BuildPaymentReminderHTML(studentName string, courses []models.Course, customMessage string) string {
	if customMessage == "" {
		customMessage = "This is a friendly reminder regarding the upcoming tuition fee payment for the course(s) you are currently enrolled in at Al Reems Academy."
	}

	portalURL := os.Getenv("FRONTEND_URL")
	if portalURL == "" {
		portalURL = "http://localhost:3001"
	}

	var courseRows strings.Builder
	var totalAmount float64 = 0

	for _, c := range courses {
		price := c.Price
		if price <= 0 {
			price = 150.0 // Default 150 AED if not set
		}
		totalAmount += price

		courseRows.WriteString(fmt.Sprintf(`
			<tr style="border-bottom: 1px solid #E5E7EB;">
				<td style="padding: 12px 16px; font-weight: 500; color: #111827;">%s</td>
				<td style="padding: 12px 16px; color: #4B5563; font-family: monospace;">%s</td>
				<td style="padding: 12px 16px; text-align: right; font-weight: 600; color: #3B0200;">%.2f AED</td>
			</tr>
		`, c.Name, c.Code, price))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Payment Reminder - Al Reems Academy</title>
</head>
<body style="margin: 0; padding: 0; background-color: #F3F4F6; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1F2937;">
  <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F3F4F6; padding: 30px 15px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 600px; background-color: #FFFFFF; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);">
          
          <!-- Header Banner -->
          <tr>
            <td style="background-color: #3B0200; padding: 32px 24px; text-align: center;">
              <h1 style="margin: 0; color: #FFFFFF; font-size: 24px; font-weight: 700; letter-spacing: 0.5px;">AL REEMS ACADEMY</h1>
              <p style="margin: 6px 0 0 0; color: #FDE68A; font-size: 14px; font-weight: 500;">Tuition Fee Payment Reminder</p>
            </td>
          </tr>

          <!-- Main Content -->
          <tr>
            <td style="padding: 32px 28px;">
              <h2 style="margin: 0 0 16px 0; color: #111827; font-size: 20px;">Dear %s,</h2>
              <p style="margin: 0 0 20px 0; color: #4B5563; font-size: 15px; line-height: 1.6;">
                %s
              </p>

              <!-- Enrolled Courses Box -->
              <div style="background-color: #F9FAFB; border: 1px solid #E5E7EB; border-radius: 8px; overflow: hidden; margin-bottom: 24px;">
                <div style="background-color: #F3F4F6; padding: 12px 16px; border-bottom: 1px solid #E5E7EB;">
                  <strong style="color: #374151; font-size: 14px; text-transform: uppercase; letter-spacing: 0.5px;">Your Enrolled Courses</strong>
                </div>
                <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="font-size: 14px; border-collapse: collapse;">
                  <thead>
                    <tr style="background-color: #FAFAFA; border-bottom: 1px solid #E5E7EB; color: #6B7280; font-size: 12px; text-transform: uppercase;">
                      <th align="left" style="padding: 10px 16px;">Course</th>
                      <th align="left" style="padding: 10px 16px;">Code</th>
                      <th align="right" style="padding: 10px 16px;">Fee (AED)</th>
                    </tr>
                  </thead>
                  <tbody>
                    %s
                    <tr style="background-color: #FFFBEB;">
                      <td colspan="2" style="padding: 14px 16px; font-weight: 700; color: #92400E;">Total Estimated Due:</td>
                      <td style="padding: 14px 16px; text-align: right; font-weight: 700; color: #92400E; font-size: 16px;">%.2f AED</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <!-- Mandatory Ignore Notice (Required by user) -->
              <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #FEF3C7; border: 1px solid #F59E0B; border-left: 5px solid #D97706; border-radius: 8px; margin: 24px 0;">
                <tr>
                  <td style="padding: 16px 20px;">
                    <div style="display: flex; align-items: flex-start;">
                      <div>
                        <strong style="color: #92400E; font-size: 15px; display: block; margin-bottom: 4px;">⚠️ Important Notice</strong>
                        <span style="color: #78350F; font-size: 14px; line-height: 1.5; display: block;">
                          <strong>Please ignore this email if you have already completed your payment.</strong> If your payment was recently processed, no further action is required.
                        </span>
                      </div>
                    </div>
                  </td>
                </tr>
              </table>

              <!-- Action Button -->
              <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="margin: 28px 0 20px 0;">
                <tr>
                  <td align="center">
                    <a href="%s/dashboard" style="display: inline-block; background-color: #3B0200; color: #FFFFFF; font-weight: 600; font-size: 15px; padding: 14px 32px; border-radius: 8px; text-decoration: none; box-shadow: 0 2px 4px rgba(59, 2, 0, 0.2);">
                      Access Student Portal & Pay
                    </a>
                  </td>
                </tr>
              </table>

              <p style="margin: 24px 0 0 0; color: #6B7280; font-size: 13px; line-height: 1.6; text-align: center;">
                If you have any questions or require support, please feel free to reach out to our administration team.
              </p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="background-color: #F9FAFB; padding: 20px 24px; text-align: center; border-top: 1px solid #E5E7EB;">
              <p style="margin: 0; color: #9CA3AF; font-size: 12px;">
                &copy; 2026 Al Reems Academy. All rights reserved.
              </p>
              <p style="margin: 4px 0 0 0; color: #9CA3AF; font-size: 12px;">
                This automated reminder was sent to your registered email address.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, studentName, customMessage, courseRows.String(), totalAmount, portalURL)
}

// StudentReminderGroup groups enrolled courses per student
type StudentReminderGroup struct {
	Student models.Student
	Courses []models.Course
}

// ExecutePaymentReminderJob sends reminder emails to all active enrolled students or a test email
func ExecutePaymentReminderJob(db *gorm.DB, testEmail string) (int, int, error) {
	var settings models.PaymentReminderSettings
	if err := db.First(&settings).Error; err != nil {
		// Create default if not found
		settings = models.PaymentReminderSettings{
			Enabled:                    false,
			ScheduleType:               "monthly",
			DayOfMonth:                 1,
			SendHour:                   9,
			SendMinute:                 0,
			EmailSubject:               "Payment Reminder: Upcoming Course Tuition Fee - Al Reems Academy",
			CustomMessage:              "This is a friendly reminder regarding the upcoming tuition fee payment for the course(s) you are currently enrolled in at Al Reems Academy.",
			StudentRegistrationEnabled: true,
			TeacherRegistrationEnabled: true,
		}
		db.Create(&settings)
	}

	subject := settings.EmailSubject
	if strings.TrimSpace(subject) == "" {
		subject = "Payment Reminder: Upcoming Course Tuition Fee - Al Reems Academy"
	}

	// Case 1: Test Email
	if testEmail != "" {
		log.Printf("Executing test payment reminder email to: %s", testEmail)
		// Fetch up to 2 sample courses or mock courses for preview
		var sampleCourses []models.Course
		db.Limit(2).Find(&sampleCourses)
		if len(sampleCourses) == 0 {
			sampleCourses = []models.Course{
				{
					Name:  "Quran Memorization & Tajweed",
					Code:  "QRN-101",
					Price: 150.0,
				},
				{
					Name:  "Islamic Studies & Arabic",
					Code:  "ARB-201",
					Price: 150.0,
				},
			}
		}

		body := BuildPaymentReminderHTML("Preview Admin / Student", sampleCourses, settings.CustomMessage)
		err := SendEmail([]string{testEmail}, subject+" [TEST PREVIEW]", body)
		if err != nil {
			return 0, 1, err
		}
		return 1, 0, nil
	}

	// Case 2: Send to all active enrolled students
	var enrollments []models.Enrollment
	if err := db.Preload("Student").Preload("Course").
		Where("LOWER(COALESCE(status, 'active')) = 'active' OR status = '' OR status IS NULL").
		Find(&enrollments).Error; err != nil {
		log.Printf("Failed to fetch enrollments for payment reminder: %v", err)
		return 0, 0, err
	}

	// Group enrollments by StudentID
	studentGroups := make(map[uint]*StudentReminderGroup)
	for _, e := range enrollments {
		if strings.TrimSpace(e.Student.Email) == "" {
			continue
		}
		if _, exists := studentGroups[e.StudentID]; !exists {
			studentGroups[e.StudentID] = &StudentReminderGroup{
				Student: e.Student,
				Courses: []models.Course{},
			}
		}
		studentGroups[e.StudentID].Courses = append(studentGroups[e.StudentID].Courses, e.Course)
	}

	sentCount := 0
	failCount := 0

	for _, group := range studentGroups {
		body := BuildPaymentReminderHTML(group.Student.Name, group.Courses, settings.CustomMessage)
		err := SendEmail([]string{group.Student.Email}, subject, body)
		if err != nil {
			log.Printf("Failed to send payment reminder to student %s (%s): %v", group.Student.Name, group.Student.Email, err)
			failCount++
		} else {
			sentCount++
		}
	}

	now := time.Now()
	settings.LastRunAt = &now
	uaeLoc, err := time.LoadLocation("Asia/Dubai")
	if err != nil {
		uaeLoc = time.FixedZone("GST", 4*3600)
	}
	nowInUAE := now.In(uaeLoc)
	settings.LastRunStatus = fmt.Sprintf("Completed: Sent to %d student(s) (%d failed) on %s", sentCount, failCount, nowInUAE.Format("Jan 02, 2006 at 15:04 GST"))
	db.Save(&settings)

	log.Printf("Payment Reminder Run Complete: %d sent, %d failed", sentCount, failCount)
	return sentCount, failCount, nil
}

// StartReminderScheduler starts a background ticker that checks schedule criteria every minute
func StartReminderScheduler(db *gorm.DB) {
	globalSchedulerDB = db
	log.Println("Starting automated Payment Reminder Mail Scheduler...")
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			checkAndTriggerScheduler(db)
		}
	}()
}

func checkAndTriggerScheduler(db *gorm.DB) {
	var settings models.PaymentReminderSettings
	if err := db.First(&settings).Error; err != nil {
		return
	}

	if !settings.Enabled {
		return
	}

	now := time.Now()

	// Check if already executed today
	if settings.LastRunAt != nil {
		if settings.LastRunAt.Year() == now.Year() && settings.LastRunAt.YearDay() == now.YearDay() {
			return
		}
	}

	// Check time match: current hour == sendHour and minute >= sendMinute
	if now.Hour() != settings.SendHour || now.Minute() < settings.SendMinute {
		return
	}

	shouldTrigger := false

	if settings.ScheduleType == "monthly" {
		targetDay := settings.DayOfMonth
		if targetDay < 1 {
			targetDay = 1
		}
		// Handle months with fewer days (e.g. Feb 28/29, or 30-day months)
		lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, now.Location()).Day()
		effectiveDay := targetDay
		if effectiveDay > lastDayOfMonth {
			effectiveDay = lastDayOfMonth
		}

		if now.Day() == effectiveDay {
			shouldTrigger = true
		}
	} else if settings.ScheduleType == "specific_date" {
		if strings.TrimSpace(settings.SpecificDate) != "" {
			todayStr := now.Format("2006-01-02")
			if todayStr == strings.TrimSpace(settings.SpecificDate) {
				shouldTrigger = true
			}
		}
	}

	if shouldTrigger {
		log.Printf("Scheduler criteria met on %s! Triggering payment reminder emails...", now.Format("2006-01-02 15:04:05"))
		go func() {
			_, _, err := ExecutePaymentReminderJob(db, "")
			if err != nil {
				log.Printf("Error during scheduled payment reminder run: %v", err)
			}
		}()
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CLASS REMINDER EMAIL SYSTEM
// ─────────────────────────────────────────────────────────────────────────────

// BuildClassReminderHTML generates a clean, responsive HTML email for class reminders
func BuildClassReminderHTML(teacherName, sessionTitle, courseName, startTime, endTime, location, joinURL string, isInstant bool) string {
	badgeColor := "#2563EB"
	badgeText := "📅 Class Starting in 30 Minutes"
	headerSubtitle := "Upcoming Class Reminder"
	urgencyNote := "Your class is scheduled to begin in approximately 30 minutes. Please prepare your materials and join on time."

	if isInstant {
		badgeColor = "#DC2626"
		badgeText = "🔴 Class Starting Now"
		headerSubtitle = "Instant Class Notification"
		urgencyNote = "A class session has been created and is starting immediately. Please join the session now."
	}

	courseRow := ""
	if courseName != "" {
		courseRow = fmt.Sprintf(`
		<tr style="border-bottom: 1px solid #E5E7EB;">
			<td style="padding: 10px 16px; color: #6B7280; font-size: 13px; width: 120px;">Course</td>
			<td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
		</tr>`, courseName)
	}

	locationRow := ""
	if location != "" {
		locationRow = fmt.Sprintf(`
		<tr style="border-bottom: 1px solid #E5E7EB;">
			<td style="padding: 10px 16px; color: #6B7280; font-size: 13px;">Location</td>
			<td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
		</tr>`, location)
	}

	joinButton := ""
	if joinURL != "" {
		joinButton = fmt.Sprintf(`
		<table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="margin: 28px 0 20px 0;">
			<tr>
				<td align="center">
					<a href="%s" style="display: inline-block; background-color: #3B0200; color: #FFFFFF; font-weight: 600; font-size: 15px; padding: 14px 36px; border-radius: 8px; text-decoration: none; box-shadow: 0 2px 4px rgba(59, 2, 0, 0.25);">
						Join Class Now →
					</a>
				</td>
			</tr>
		</table>`, joinURL)
	}

	portalURL := os.Getenv("TEACHERS_URL")
	if portalURL == "" {
		portalURL = os.Getenv("FRONTEND_URL")
		if portalURL == "" {
			portalURL = "http://localhost:3002"
		}
	}
	if joinURL == "" {
		joinButton = fmt.Sprintf(`
		<table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="margin: 28px 0 20px 0;">
			<tr>
				<td align="center">
					<a href="%s/sessions" style="display: inline-block; background-color: #3B0200; color: #FFFFFF; font-weight: 600; font-size: 15px; padding: 14px 36px; border-radius: 8px; text-decoration: none; box-shadow: 0 2px 4px rgba(59, 2, 0, 0.25);">
						Go to My Sessions →
					</a>
				</td>
			</tr>
		</table>`, portalURL)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Class Reminder - Al Reems Academy</title>
</head>
<body style="margin: 0; padding: 0; background-color: #F3F4F6; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1F2937;">
  <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F3F4F6; padding: 30px 15px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 600px; background-color: #FFFFFF; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);">

          <!-- Header Banner -->
          <tr>
            <td style="background-color: #3B0200; padding: 32px 24px; text-align: center;">
              <h1 style="margin: 0; color: #FFFFFF; font-size: 24px; font-weight: 700; letter-spacing: 0.5px;">AL REEMS ACADEMY</h1>
              <p style="margin: 6px 0 0 0; color: #FDE68A; font-size: 14px; font-weight: 500;">%s</p>
            </td>
          </tr>

          <!-- Badge Row -->
          <tr>
            <td style="padding: 0; text-align: center; background-color: %s;">
              <p style="margin: 0; padding: 10px 24px; color: #FFFFFF; font-size: 13px; font-weight: 600; letter-spacing: 0.5px;">%s</p>
            </td>
          </tr>

          <!-- Main Content -->
          <tr>
            <td style="padding: 32px 28px;">
              <h2 style="margin: 0 0 8px 0; color: #111827; font-size: 20px;">Dear %s,</h2>
              <p style="margin: 0 0 24px 0; color: #4B5563; font-size: 15px; line-height: 1.6;">%s</p>

              <!-- Session Details Box -->
              <div style="background-color: #F9FAFB; border: 1px solid #E5E7EB; border-radius: 8px; overflow: hidden; margin-bottom: 24px;">
                <div style="background-color: #F3F4F6; padding: 12px 16px; border-bottom: 1px solid #E5E7EB;">
                  <strong style="color: #374151; font-size: 14px; text-transform: uppercase; letter-spacing: 0.5px;">Session Details</strong>
                </div>
                <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="font-size: 14px; border-collapse: collapse;">
                  <tbody>
                    <tr style="border-bottom: 1px solid #E5E7EB;">
                      <td style="padding: 10px 16px; color: #6B7280; font-size: 13px; width: 120px;">Session</td>
                      <td style="padding: 10px 16px; font-weight: 600; color: #111827;">%s</td>
                    </tr>
                    %s
                    <tr style="border-bottom: 1px solid #E5E7EB;">
                      <td style="padding: 10px 16px; color: #6B7280; font-size: 13px;">Start Time</td>
                      <td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
                    </tr>
                    <tr style="border-bottom: 1px solid #E5E7EB;">
                      <td style="padding: 10px 16px; color: #6B7280; font-size: 13px;">End Time</td>
                      <td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
                    </tr>
                    %s
                  </tbody>
                </table>
              </div>

              %s

              <p style="margin: 24px 0 0 0; color: #6B7280; font-size: 13px; line-height: 1.6; text-align: center;">
                If you have any questions, please contact the academy administration.
              </p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="background-color: #F9FAFB; padding: 20px 24px; text-align: center; border-top: 1px solid #E5E7EB;">
              <p style="margin: 0; color: #9CA3AF; font-size: 12px;">
                &copy; 2026 Al Reems Academy. All rights reserved.
              </p>
              <p style="margin: 4px 0 0 0; color: #9CA3AF; font-size: 12px;">
                This automated class reminder was sent to your registered email address.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		headerSubtitle,
		badgeColor, badgeText,
		teacherName,
		urgencyNote,
		sessionTitle,
		courseRow,
		startTime,
		endTime,
		locationRow,
		joinButton,
	)
}

var globalSchedulerDB *gorm.DB

// SetSchedulerDB registers the global DB instance for the reminder schedulers
func SetSchedulerDB(db *gorm.DB) {
	globalSchedulerDB = db
}

// BuildStudentClassReminderHTML generates a clean, responsive HTML email for enrolled students
func BuildStudentClassReminderHTML(studentName, teacherName, sessionTitle, courseName, startTime, endTime, location, joinURL string, isInstant bool) string {
	badgeColor := "#2563EB"
	badgeText := "📅 Live Class Starting in 30 Minutes"
	headerSubtitle := "Upcoming Class Reminder"
	urgencyNote := fmt.Sprintf("Your class session <strong>%s</strong> is scheduled to begin in approximately 30 minutes. Please prepare your materials and join on time.", sessionTitle)

	if isInstant {
		badgeColor = "#DC2626"
		badgeText = "🔴 Class Starting Now — Join Live"
		headerSubtitle = "Live Class Notification"
		urgencyNote = fmt.Sprintf("Your class session <strong>%s</strong> is starting right now! Click the button below to join the classroom immediately.", sessionTitle)
	}

	courseRow := ""
	if courseName != "" {
		courseRow = fmt.Sprintf(`
		<tr style="border-bottom: 1px solid #E5E7EB;">
			<td style="padding: 10px 16px; color: #6B7280; font-size: 13px; width: 120px;">Course</td>
			<td style="padding: 10px 16px; font-weight: 600; color: #111827;">%s</td>
		</tr>`, courseName)
	}

	instructorRow := ""
	if teacherName != "" {
		instructorRow = fmt.Sprintf(`
		<tr style="border-bottom: 1px solid #E5E7EB;">
			<td style="padding: 10px 16px; color: #6B7280; font-size: 13px; width: 120px;">Instructor</td>
			<td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
		</tr>`, teacherName)
	}

	locationRow := ""
	if location != "" {
		locationRow = fmt.Sprintf(`
		<tr style="border-bottom: 1px solid #E5E7EB;">
			<td style="padding: 10px 16px; color: #6B7280; font-size: 13px;">Location</td>
			<td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
		</tr>`, location)
	}

	joinButton := fmt.Sprintf(`
	<table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="margin: 28px 0 20px 0;">
		<tr>
			<td align="center">
				<a href="%s" style="display: inline-block; background-color: #3B0200; color: #FFFFFF; font-weight: 600; font-size: 15px; padding: 14px 36px; border-radius: 8px; text-decoration: none; box-shadow: 0 2px 4px rgba(59, 2, 0, 0.25);">
					Join Live Class Now →
				</a>
			</td>
		</tr>
	</table>`, joinURL)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Class Reminder - Al Reems Academy</title>
</head>
<body style="margin: 0; padding: 0; background-color: #F3F4F6; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; color: #1F2937;">
  <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #F3F4F6; padding: 30px 15px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 600px; background-color: #FFFFFF; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);">

          <!-- Header Banner -->
          <tr>
            <td style="background-color: #3B0200; padding: 32px 24px; text-align: center;">
              <h1 style="margin: 0; color: #FFFFFF; font-size: 24px; font-weight: 700; letter-spacing: 0.5px;">AL REEMS ACADEMY</h1>
              <p style="margin: 6px 0 0 0; color: #FDE68A; font-size: 14px; font-weight: 500;">%s</p>
            </td>
          </tr>

          <!-- Badge Row -->
          <tr>
            <td style="padding: 0; text-align: center; background-color: %s;">
              <p style="margin: 0; padding: 10px 24px; color: #FFFFFF; font-size: 13px; font-weight: 600; letter-spacing: 0.5px;">%s</p>
            </td>
          </tr>

          <!-- Main Content -->
          <tr>
            <td style="padding: 32px 28px;">
              <h2 style="margin: 0 0 8px 0; color: #111827; font-size: 20px;">Dear %s,</h2>
              <p style="margin: 0 0 24px 0; color: #4B5563; font-size: 15px; line-height: 1.6;">%s</p>

              <!-- Session Details Box -->
              <div style="background-color: #F9FAFB; border: 1px solid #E5E7EB; border-radius: 8px; overflow: hidden; margin-bottom: 24px;">
                <div style="background-color: #F3F4F6; padding: 12px 16px; border-bottom: 1px solid #E5E7EB;">
                  <strong style="color: #374151; font-size: 14px; text-transform: uppercase; letter-spacing: 0.5px;">Session Details</strong>
                </div>
                <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="font-size: 14px; border-collapse: collapse;">
                  <tbody>
                    <tr style="border-bottom: 1px solid #E5E7EB;">
                      <td style="padding: 10px 16px; color: #6B7280; font-size: 13px; width: 120px;">Session</td>
                      <td style="padding: 10px 16px; font-weight: 600; color: #111827;">%s</td>
                    </tr>
                    %s
                    %s
                    <tr style="border-bottom: 1px solid #E5E7EB;">
                      <td style="padding: 10px 16px; color: #6B7280; font-size: 13px;">Start Time</td>
                      <td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
                    </tr>
                    <tr style="border-bottom: 1px solid #E5E7EB;">
                      <td style="padding: 10px 16px; color: #6B7280; font-size: 13px;">End Time</td>
                      <td style="padding: 10px 16px; font-weight: 500; color: #111827;">%s</td>
                    </tr>
                    %s
                  </tbody>
                </table>
              </div>

              %s

              <p style="margin: 24px 0 0 0; color: #6B7280; font-size: 13px; line-height: 1.6; text-align: center;">
                Need assistance? Feel free to reach out to our academy administration.
              </p>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="background-color: #F9FAFB; padding: 20px 24px; text-align: center; border-top: 1px solid #E5E7EB;">
              <p style="margin: 0; color: #9CA3AF; font-size: 12px;">
                &copy; 2026 Al Reems Academy. All rights reserved.
              </p>
              <p style="margin: 4px 0 0 0; color: #9CA3AF; font-size: 12px;">
                This automated reminder was sent to your registered student email address.
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		headerSubtitle,
		badgeColor, badgeText,
		studentName,
		urgencyNote,
		sessionTitle,
		courseRow,
		instructorRow,
		startTime,
		endTime,
		locationRow,
		joinButton,
	)
}

// SendClassReminders dispatches email reminders to BOTH the teacher AND all enrolled students for a session
func SendClassReminders(db *gorm.DB, session models.Session, isInstant bool) {
	if db == nil {
		db = globalSchedulerDB
	}

	// 1. Convert session times to UAE GST (Gulf Standard Time, UTC+4)
	uaeLoc, err := time.LoadLocation("Asia/Dubai")
	if err != nil {
		uaeLoc = time.FixedZone("GST", 4*3600)
	}
	startInUAE := session.StartTime.In(uaeLoc)
	endInUAE := session.EndTime.In(uaeLoc)
	startFormatted := startInUAE.Format("Mon, 02 Jan 2006 at 03:04 PM GST")
	endFormatted := endInUAE.Format("03:04 PM GST")

	courseName := ""
	if session.Course.ID != 0 {
		courseName = session.Course.Name
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://reems.co.in"
	}
	studentJoinURL := fmt.Sprintf("%s/dashboard/session/%d", frontendURL, session.ID)

	teachersURL := os.Getenv("TEACHERS_URL")
	if teachersURL == "" {
		teachersURL = "https://teachers.reems.co.in"
	}
	teacherJoinURL := fmt.Sprintf("%s/sessions", teachersURL)

	subjectPrefix := "⏰ Class Reminder (30 Min)"
	if isInstant {
		subjectPrefix = "🔴 Class Starting Now"
	}
	subject := fmt.Sprintf("%s: %s — Al Reems Academy", subjectPrefix, session.Title)

	// 2. Dispatch reminder to Teacher
	if strings.TrimSpace(session.Teacher.Email) != "" {
		teacherBody := BuildClassReminderHTML(
			session.Teacher.Name,
			session.Title,
			courseName,
			startFormatted,
			endFormatted,
			session.Location,
			teacherJoinURL,
			isInstant,
		)
		if err := SendEmail([]string{session.Teacher.Email}, subject, teacherBody); err != nil {
			log.Printf("[ClassReminder] Failed to send reminder to teacher %s (%s): %v", session.Teacher.Name, session.Teacher.Email, err)
		} else {
			log.Printf("[ClassReminder] Reminder sent to teacher %s (%s) for session '%s'", session.Teacher.Name, session.Teacher.Email, session.Title)
		}
	}

	// 3. Dispatch reminder to all Enrolled Students for this course
	var courseID uint
	if session.CourseID != nil && *session.CourseID != 0 {
		courseID = *session.CourseID
	} else if session.Course.ID != 0 {
		courseID = session.Course.ID
	}

	if courseID != 0 && db != nil {
		var enrollments []models.Enrollment
		err := db.Preload("Student").
			Where("course_id = ? AND (LOWER(COALESCE(status, 'active')) = 'active' OR status = '' OR status IS NULL)", courseID).
			Find(&enrollments).Error

		if err != nil {
			log.Printf("[ClassReminder] Failed to query enrolled students for course %d: %v", courseID, err)
		} else {
			sentCount := 0
			failCount := 0
			for _, e := range enrollments {
				studentEmail := strings.TrimSpace(e.Student.Email)
				if studentEmail == "" {
					continue
				}

				studentName := e.Student.Name
				if studentName == "" {
					studentName = "Student"
				}

				teacherDisplayName := session.Teacher.Name
				if teacherDisplayName == "" {
					teacherDisplayName = "Your Instructor"
				}

				studentBody := BuildStudentClassReminderHTML(
					studentName,
					teacherDisplayName,
					session.Title,
					courseName,
					startFormatted,
					endFormatted,
					session.Location,
					studentJoinURL,
					isInstant,
				)

				if err := SendEmail([]string{studentEmail}, subject, studentBody); err != nil {
					log.Printf("[ClassReminder] Failed to send class reminder to student %s (%s): %v", studentName, studentEmail, err)
					failCount++
				} else {
					log.Printf("[ClassReminder] Class reminder sent to student %s (%s) for session '%s'", studentName, studentEmail, session.Title)
					sentCount++
				}

				// Optional WhatsApp notification
				if strings.TrimSpace(e.Student.PhoneNumber) != "" {
					waUrgency := "is starting now"
					if !isInstant {
						waUrgency = "is starting in 30 minutes"
					}
					waMsg := fmt.Sprintf("Hi %s, your class '%s' for %s %s. Join here: %s",
						studentName, session.Title, courseName, waUrgency, studentJoinURL)
					_ = SendWhatsAppMessage(e.Student.PhoneNumber, waMsg)
				}
			}
			log.Printf("[ClassReminder] Student reminder dispatch complete for session %d: %d sent, %d failed", session.ID, sentCount, failCount)
		}
	} else if courseID == 0 {
		log.Printf("[ClassReminder] Session %d has no associated course; skipping student reminder dispatch", session.ID)
	}

	// 4. Mark reminder_sent in DB if scheduled
	if db != nil && !isInstant {
		db.Model(&models.Session{}).Where("id = ?", session.ID).Update("reminder_sent", true)
	}
}

// SendClassReminderToTeacher sends an immediate class reminder to the teacher (for instant/live sessions)
func SendClassReminderToTeacher(session models.Session) {
	if globalSchedulerDB != nil {
		SendClassReminders(globalSchedulerDB, session, true)
		return
	}

	if session.Teacher.Email == "" {
		log.Printf("[ClassReminder] Skipping instant reminder for session %d — teacher email is empty", session.ID)
		return
	}

	uaeLoc, err := time.LoadLocation("Asia/Dubai")
	if err != nil {
		uaeLoc = time.FixedZone("GST", 4*3600)
	}
	startFormatted := session.StartTime.In(uaeLoc).Format("Mon, 02 Jan 2006 at 03:04 PM GST")
	endFormatted := session.EndTime.In(uaeLoc).Format("03:04 PM GST")

	courseName := ""
	if session.Course.ID != 0 {
		courseName = session.Course.Name
	}

	subject := fmt.Sprintf("🔴 Class Starting Now: %s — Al Reems Academy", session.Title)
	body := BuildClassReminderHTML(
		session.Teacher.Name,
		session.Title,
		courseName,
		startFormatted,
		endFormatted,
		session.Location,
		"", // join URL (portal will have session listing)
		true,
	)

	if err := SendEmail([]string{session.Teacher.Email}, subject, body); err != nil {
		log.Printf("[ClassReminder] Failed to send instant reminder to teacher %s (%s): %v", session.Teacher.Name, session.Teacher.Email, err)
	} else {
		log.Printf("[ClassReminder] Instant reminder sent to teacher %s (%s) for session '%s'", session.Teacher.Name, session.Teacher.Email, session.Title)
	}
}

// StartClassReminderScheduler starts a background ticker that checks every minute
// for sessions starting in 25–35 minutes and sends a 30-minute-ahead reminder email.
func StartClassReminderScheduler(db *gorm.DB) {
	globalSchedulerDB = db
	log.Println("[ClassReminder] Starting Class Reminder Scheduler (30-min pre-class emails)...")
	ticker := time.NewTicker(1 * time.Minute)

	go func() {
		for range ticker.C {
			checkAndSendClassReminders(db)
		}
	}()
}

func checkAndSendClassReminders(db *gorm.DB) {
	now := time.Now()
	windowStart := now.Add(25 * time.Minute)
	windowEnd := now.Add(35 * time.Minute)

	// Find scheduled sessions that start within the 25–35 minute window and haven't had a reminder sent
	var sessions []models.Session
	if err := db.
		Preload("Teacher").
		Preload("Course").
		Where("status = ? AND start_time BETWEEN ? AND ? AND (reminder_sent = ? OR reminder_sent IS NULL)", "scheduled", windowStart, windowEnd, false).
		Find(&sessions).Error; err != nil {
		log.Printf("[ClassReminder] Failed to query upcoming sessions: %v", err)
		return
	}

	for _, session := range sessions {
		log.Printf("[ClassReminder] Processing 30-min scheduled reminder for session %d ('%s')...", session.ID, session.Title)
		SendClassReminders(db, session, false)
	}
}

// StartAllSchedulers launches both the payment reminder scheduler and the class reminder scheduler.
// Call this from main.go instead of StartReminderScheduler directly.
func StartAllSchedulers(db *gorm.DB) {
	globalSchedulerDB = db
	StartReminderScheduler(db)
	StartClassReminderScheduler(db)
}

