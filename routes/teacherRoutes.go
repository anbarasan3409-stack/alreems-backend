package routes

import (
	"backend/controllers"
	"backend/database"
	"backend/middlewares"

	"github.com/gin-gonic/gin"
)

func TeacherRoutes(router *gin.Engine, handler database.Handler) {
	// Teacher auth routes (don't require authentication)
	router.POST("/api/teacher/auth/refresh", controllers.TeacherRefreshTokenController(handler))
	router.POST("/api/teacher/auth/logout", controllers.TeacherLogoutController(handler))

	// Protected teacher routes
	teacherGroup := router.Group("/api/teacher")
	teacherGroup.Use(middlewares.TeacherAuthMiddleware())

	// Teacher profile
	teacherGroup.GET("/profile", func(c *gin.Context) {
		controllers.GetTeacherProfile(c, handler)
	})

	// Teacher course and student management
	teacherGroup.GET("/courses", controllers.GetTeacherCourses(handler))
	teacherGroup.GET("/courses/:courseId/students", controllers.GetTeacherCourseStudents(handler))
	teacherGroup.GET("/students", controllers.GetAllTeacherStudents(handler))

	// Teacher session management
	teacherGroup.GET("/sessions", controllers.GetTeacherSessions(handler))
	teacherGroup.GET("/sessions/:id", controllers.GetTeacherSessionByID(handler))
	teacherGroup.POST("/sessions", controllers.CreateTeacherSession(handler))
	teacherGroup.PUT("/sessions/:id", controllers.UpdateTeacherSession(handler))
	teacherGroup.DELETE("/sessions/:id", controllers.DeleteTeacherSession(handler))

	// Attendance routes
	teacherGroup.POST("/attendance", controllers.MarkAttendance(handler))
	teacherGroup.GET("/attendance/session/:sessionId", controllers.GetSessionAttendance(handler))

	// Material routes
	teacherGroup.POST("/materials", controllers.CreateMaterial(handler))
	teacherGroup.GET("/courses/:courseId/materials", controllers.GetMaterialsByCourse(handler))
	teacherGroup.DELETE("/materials/:id", controllers.DeleteMaterial(handler))

	// Assignment routes
	teacherGroup.POST("/assignments", controllers.CreateAssignment(handler))
	teacherGroup.GET("/courses/:courseId/assignments", controllers.GetAssignmentsByCourse(handler))
	teacherGroup.DELETE("/assignments/:id", controllers.DeleteAssignment(handler))
	teacherGroup.GET("/assignments/:assignmentId/submissions", controllers.GetAssignmentSubmissions(handler))
	teacherGroup.POST("/assignments/submissions/:submissionId/grade", controllers.GradeAssignment(handler))

	// Recording routes for teachers
	teacherGroup.POST("/sessions/:id/recording/start", controllers.StartSessionRecording(handler))
	teacherGroup.POST("/sessions/:id/recording/stop", controllers.StopSessionRecording(handler))
	teacherGroup.GET("/sessions/:id/recording/query", controllers.QuerySessionRecording(handler))
	teacherGroup.GET("/sessions/:id/recording", controllers.GetSessionRecording(handler))

	// Agora RTC and Whiteboard token routes
	teacherGroup.GET("/rtc-token", controllers.GenerateRTCToken(handler))
	teacherGroup.GET("/agora/app-id", controllers.GetAgoraAppID(handler))
	teacherGroup.GET("/whiteboard-token", controllers.GenerateWhiteboardToken(handler))
	teacherGroup.GET("/netless/app-id", controllers.GetNetlessAppID(handler))
}
