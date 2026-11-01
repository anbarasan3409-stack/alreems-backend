package routes

import (
	"backend/controllers"
	"backend/database"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

func StudentRoutes(router *gin.Engine, handler database.Handler) {
	// Student auth routes (don't require authentication)
	router.POST("/api/student/auth/refresh", controllers.StudentRefreshTokenController(handler))
	router.POST("/api/student/auth/logout", controllers.StudentLogoutController(handler))

	// Protected student routes
	studentGroup := router.Group("/api/student")
	studentGroup.Use(middlewares.StudentAuthMiddleware())
	studentGroup.GET("/profile", func(c *gin.Context) {
		controllers.GetStudentProfile(c, handler)
	})
	// Add profile status check for logged-in student
	studentGroup.GET("/check-profile", controllers.CheckLoggedInStudentProfile(handler))
	// Add profile update for logged-in student
	studentGroup.PUT("/profile", controllers.UpdateLoggedInStudentProfile(handler))

	// Public student data routes (by ID)
	studentPublicGroup := router.Group("/api/students")
	studentPublicGroup.GET("/:id", controllers.GetStudentByID(handler))
	studentPublicGroup.PUT("/:id", controllers.UpdateStudentProfile(handler))
	studentPublicGroup.POST("/:id/register", controllers.RegisterStudentForCourse(handler))
	studentPublicGroup.GET("/:id/enrolled", controllers.GetEnrolledCourses(handler))
	studentPublicGroup.GET("/:id/analytics", controllers.GetStudentAnalytics(handler))
	studentPublicGroup.GET("/:id/activity", controllers.GetStudentWeeklyActivity(handler))
	studentPublicGroup.POST("/:id/activity/log", controllers.LogStudyTime(handler))
	studentPublicGroup.POST("/:id/subscribe", controllers.CreateCheckoutSession)
	studentPublicGroup.GET("/:id/subscriptions", controllers.GetStudentSubscriptions)

	// Attendance route
	studentPublicGroup.GET("/:id/attendance", controllers.GetStudentAttendance(handler))
	studentPublicGroup.POST("/attendance/mark", controllers.StudentSelfMarkAttendance(handler))

	// Materials
	studentPublicGroup.GET("/courses/:courseId/materials", controllers.GetMaterialsByCourse(handler))

	// Assignments
	studentPublicGroup.GET("/courses/:courseId/assignments", controllers.GetAssignmentsByCourse(handler))
	studentPublicGroup.POST("/assignments/submit", controllers.SubmitAssignment(handler))

	// Profile completion status routes
	studentPublicGroup.GET("/:id/profile-status", controllers.CheckProfileCompletionStatus(handler))
	studentPublicGroup.GET("/profile-status/all", controllers.GetAllStudentsProfileStatus(handler))

	// Student session routes (accessible to all students)
	studentPublicGroup.GET("/sessions", controllers.GetStudentSessions(handler))
	studentPublicGroup.GET("/sessions/:id", controllers.GetStudentSessionByID(handler))
	studentPublicGroup.GET("/rtc-token", controllers.GenerateStudentRTCToken(handler))

	// Course routes accessible to students
	studentPublicGroup.GET("/courses", controllers.GetAllCourses(handler))
	studentPublicGroup.GET("/courses/:courseId", controllers.GetCourseByID(handler))

	// Public Enquiry route
	router.POST("/api/enquiries", controllers.SubmitEnquiry)

	// Public Feedback routes
	router.POST("/api/feedbacks", controllers.SubmitFeedback)
	router.GET("/api/feedbacks/testimonials", controllers.GetFeaturedFeedbacks)
}
