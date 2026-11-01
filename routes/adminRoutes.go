package routes

import (
	controllers "backend/controllers"
	"backend/database"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

func AdminRoutes(router *gin.Engine, handler database.Handler) {
	// Admin auth routes (don't require authentication)
	router.POST("/api/auth/admin/refresh", controllers.AdminRefreshTokenController(handler))
	router.POST("/api/auth/admin/logout", controllers.AdminLogoutController(handler))

	adminGroup := router.Group("/api/admin")
	adminGroup.Use(middlewares.AdminAuthMiddleware())
	adminGroup.GET("/profile", func(c *gin.Context) {
		controllers.GetAdminProfile(c, handler)
	})

	// Admin management routes
	adminGroup.GET("/admins", controllers.GetAllAdmins(handler))
	adminGroup.GET("/admins/:id", controllers.GetAdminByID(handler))
	adminGroup.POST("/admins", controllers.CreateAdmin(handler))
	adminGroup.PUT("/admins/:id", controllers.UpdateAdmin(handler))
	adminGroup.DELETE("/admins/:id", controllers.DeleteAdmin(handler))

	adminGroup.GET("/students", controllers.GetAllStudents(handler))
	adminGroup.POST("/students", controllers.CreateStudentByAdmin(handler))
	adminGroup.GET("/students/:id", controllers.GetAdminStudentByID(handler))
	adminGroup.DELETE("/students/:id", controllers.DeleteStudent(handler))
	adminGroup.GET("/students/profile-status", controllers.GetAllStudentsProfileStatus(handler))
	adminGroup.GET("/students/:id/profile-status", controllers.CheckProfileCompletionStatus(handler))
	adminGroup.GET("/students/teacher/:teacherId", controllers.GetStudentsByTeacher(handler))
	adminGroup.GET("/students/course/:courseId", controllers.GetStudentsByCourse(handler))
	adminGroup.GET("/courses/student/:studentId", controllers.GetCoursesByStudent(handler))
	adminGroup.GET("/teachers", controllers.GetAllTeachers(handler))
	adminGroup.GET("/teachers/:id", controllers.GetTeacherByID(handler))
	adminGroup.POST("/teachers", controllers.CreateTeacher(handler))
	adminGroup.PUT("/teachers/:id", controllers.UpdateTeacher(handler))
	adminGroup.DELETE("/teachers/:id", controllers.DeleteTeacher(handler))
	adminGroup.GET("/courses", controllers.GetAllCourses(handler))
	adminGroup.GET("/courses/:id", controllers.GetCourseByID(handler))
	adminGroup.GET("/analytics", controllers.GetAnalytics(handler))

	// Stats route for dashboard
	adminGroup.GET("/stats", controllers.GetStats(handler))

	adminGroup.POST("/courses", controllers.CreateCourse(handler))
	adminGroup.PUT("/courses/:id", controllers.EditCourse(handler))
	adminGroup.DELETE("/courses/:id", controllers.DeleteCourse(handler))

	adminGroup.GET("/payments", controllers.GetAllPayments(handler))
	adminGroup.POST("/payments", controllers.CreatePayment(handler))
	adminGroup.PUT("/payments/:id/status", controllers.UpdatePaymentStatus(handler))
	adminGroup.DELETE("/payments/:id", controllers.DeletePayment(handler))
	adminGroup.GET("/subscriptions", controllers.GetAllSubscriptions)

	adminGroup.GET("/settings", controllers.GetSystemSettings(handler))
	adminGroup.PUT("/settings", controllers.UpdateSystemSettings(handler))
	adminGroup.POST("/settings/payment-reminder/send-now", controllers.SendPaymentRemindersNow(handler))
	adminGroup.POST("/settings/payment-reminder/test", controllers.SendTestPaymentReminder(handler))
	adminGroup.GET("/login-activities", controllers.GetLoginActivities(handler))
	adminGroup.PUT("/profile/update", controllers.UpdateAdminProfile(handler))

	// Session routes
	adminGroup.GET("/sessions", controllers.GetAllSessions(handler))
	adminGroup.GET("/sessions/:id", controllers.GetSessionByID(handler))
	adminGroup.POST("/sessions", controllers.CreateSession(handler))
	adminGroup.PUT("/sessions/:id", controllers.UpdateSession(handler))
	adminGroup.DELETE("/sessions/:id", controllers.DeleteSession(handler))
	adminGroup.POST("/sessions/:id/send-reminder", controllers.SendSessionReminderNow(handler))

	// Enquiry routes
	adminGroup.GET("/enquiries", controllers.GetAllEnquiries(handler))
	adminGroup.PUT("/enquiries/:id/status", controllers.UpdateEnquiryStatus(handler))
	adminGroup.DELETE("/enquiries/:id", controllers.DeleteEnquiry(handler))

	// Feedback routes
	adminGroup.GET("/feedbacks", controllers.GetAllFeedbacks(handler))
	adminGroup.PUT("/feedbacks/:id/toggle-featured", controllers.ToggleFeaturedFeedback(handler))
	adminGroup.DELETE("/feedbacks/:id", controllers.DeleteFeedback(handler))

	// Export routes
	adminGroup.GET("/export/students", controllers.ExportStudentsCSV(handler))
	adminGroup.GET("/export/enquiries", controllers.ExportEnquiriesCSV(handler))

	// Attendance routes
	adminGroup.POST("/attendance", controllers.MarkAttendance(handler))
	adminGroup.GET("/attendance/session/:sessionId", controllers.GetSessionAttendance(handler))
	adminGroup.GET("/attendance/student/:studentId", controllers.GetStudentAttendance(handler))

	// Recording routes
	adminGroup.POST("/sessions/:id/recording/start", controllers.StartSessionRecording(handler))
	adminGroup.POST("/sessions/:id/recording/stop", controllers.StopSessionRecording(handler))
	adminGroup.GET("/sessions/:id/recording/query", controllers.QuerySessionRecording(handler))
	adminGroup.GET("/sessions/:id/recording", controllers.GetSessionRecording(handler))
	adminGroup.PUT("/sessions/:id/recording/url", controllers.UpdateRecordingURL(handler))

	// Agora RTC token routes
	adminGroup.GET("/rtc-token", controllers.GenerateRTCToken(handler))
	adminGroup.GET("/agora/app-id", controllers.GetAgoraAppID(handler))

	// Whiteboard token routes
	adminGroup.GET("/whiteboard-token", controllers.GenerateWhiteboardToken(handler))
	adminGroup.GET("/netless/app-id", controllers.GetNetlessAppID(handler))
}
