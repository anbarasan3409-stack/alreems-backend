package controllers

import (
	"fmt"
	"strconv"

	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// GetTeacherCourses - Get all courses taught by the authenticated teacher
func GetTeacherCourses(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get teacher ID from context (set by auth middleware)
		teacherIDInterface, exists := ctx.Get("teacherID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Teacher not authenticated"})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		var courses []models.Course
		if err := handler.DB.Where("teacher_id = ?", teacherID).Find(&courses).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch courses"})
			return
		}

		// Get enrollment count for each course
		var response []gin.H
		for _, course := range courses {
			var enrollmentCount int64
			handler.DB.Model(&models.Enrollment{}).Where("course_id = ?", course.ID).Count(&enrollmentCount)

			courseData := gin.H{
				"ID":              course.ID,
				"Name":            course.Name,
				"Code":            course.Code,
				"Description":     course.Description,
				"Credits":         course.Credits,
				"CreatedAt":       course.CreatedAt,
				"UpdatedAt":       course.UpdatedAt,
				"enrollmentCount": enrollmentCount,
			}
			response = append(response, courseData)
		}

		ctx.JSON(200, response)
	}
}

// GetTeacherCourseStudents - Get all students enrolled in a specific course taught by the teacher
func GetTeacherCourseStudents(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get teacher ID from context
		teacherIDInterface, exists := ctx.Get("teacherID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Teacher not authenticated"})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		// Get course ID from URL parameter
		courseIDStr := ctx.Param("courseId")
		courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
		if err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid course ID"})
			return
		}

		// Verify that the course belongs to this teacher
		var course models.Course
		if err := handler.DB.Where("id = ? AND teacher_id = ?", courseID, teacherID).First(&course).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found or you don't have permission to access it"})
			return
		}

		// Get enrollments with student information
		var enrollments []models.Enrollment
		if err := handler.DB.Preload("Student").Where("course_id = ?", courseID).Find(&enrollments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch enrolled students"})
			return
		}

		// Format response
		var students []gin.H
		for _, enrollment := range enrollments {
			studentData := gin.H{
				"enrollmentID": enrollment.ID,
				"studentID":    enrollment.Student.ID,
				"name":         enrollment.Student.Name,
				"email":        enrollment.Student.Email,
				"phoneNumber":  enrollment.Student.PhoneNumber,
				"address":      enrollment.Student.Address,
				"progress":     enrollment.Progress,
				"status":       enrollment.Status,
				"enrolledAt":   enrollment.EnrolledAt,
				"completedAt":  enrollment.CompletedAt,
			}
			students = append(students, studentData)
		}

		response := gin.H{
			"course": gin.H{
				"ID":          course.ID,
				"Name":        course.Name,
				"Code":        course.Code,
				"Description": course.Description,
				"Credits":     course.Credits,
			},
			"students":   students,
			"totalCount": len(students),
		}

		ctx.JSON(200, response)
	}
}

// GetAllTeacherStudents - Get all students enrolled in any course taught by the teacher
func GetAllTeacherStudents(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get teacher ID from context
		teacherIDInterface, exists := ctx.Get("teacherID")
		if !exists {
			ctx.JSON(401, gin.H{"error": "Teacher not authenticated"})
			return
		}

		teacherID, ok := teacherIDInterface.(uint)
		if !ok {
			ctx.JSON(400, gin.H{"error": "Invalid teacher ID"})
			return
		}

		// Get all courses taught by this teacher
		var courses []models.Course
		if err := handler.DB.Where("teacher_id = ?", teacherID).Find(&courses).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch courses"})
			return
		}

		// Debug logging
		fmt.Printf("Teacher ID: %d, Found courses: %+v\n", teacherID, courses)

		// Extract course IDs
		var courseIDs []uint
		for _, course := range courses {
			courseIDs = append(courseIDs, course.ID)
		}

		fmt.Printf("Course IDs: %+v\n", courseIDs)

		if len(courseIDs) == 0 {
			ctx.JSON(200, gin.H{
				"students":   []gin.H{},
				"totalCount": 0,
			})
			return
		}

		// Get all enrollments for these courses with student and course information
		var enrollments []models.Enrollment
		if err := handler.DB.Preload("Student").Preload("Course").Where("course_id IN ?", courseIDs).Find(&enrollments).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch enrolled students"})
			return
		}

		fmt.Printf("Found enrollments: %d\n", len(enrollments))
		for i, enrollment := range enrollments {
			fmt.Printf("Enrollment %d: StudentID=%d, CourseID=%d, Student=%+v\n", i, enrollment.StudentID, enrollment.CourseID, enrollment.Student)
		}

		// Group students by course and format response
		studentMap := make(map[uint]gin.H)
		courseEnrollments := make(map[uint][]gin.H)

		for _, enrollment := range enrollments {
			// Add student to unique students map
			if _, exists := studentMap[enrollment.Student.ID]; !exists {
				studentMap[enrollment.Student.ID] = gin.H{
					"studentID":   enrollment.Student.ID,
					"name":        enrollment.Student.Name,
					"email":       enrollment.Student.Email,
					"phoneNumber": enrollment.Student.PhoneNumber,
					"address":     enrollment.Student.Address,
				}
			}

			// Add enrollment to course-specific list
			enrollmentData := gin.H{
				"enrollmentID": enrollment.ID,
				"studentID":    enrollment.Student.ID,
				"studentName":  enrollment.Student.Name,
				"progress":     enrollment.Progress,
				"status":       enrollment.Status,
				"enrolledAt":   enrollment.EnrolledAt,
				"completedAt":  enrollment.CompletedAt,
			}
			courseEnrollments[enrollment.Course.ID] = append(courseEnrollments[enrollment.Course.ID], enrollmentData)
		}

		// Convert map to slice for unique students
		var uniqueStudents []gin.H
		for _, student := range studentMap {
			uniqueStudents = append(uniqueStudents, student)
		}

		// Format courses with their enrollments
		var coursesWithStudents []gin.H
		for _, course := range courses {
			courseData := gin.H{
				"courseID":     course.ID,
				"name":         course.Name,
				"code":         course.Code,
				"description":  course.Description,
				"credits":      course.Credits,
				"enrollments":  courseEnrollments[course.ID],
				"studentCount": len(courseEnrollments[course.ID]),
			}
			coursesWithStudents = append(coursesWithStudents, courseData)
		}

		response := gin.H{
			"students":         uniqueStudents,
			"totalCount":       len(uniqueStudents),
			"courses":          coursesWithStudents,
			"totalEnrollments": len(enrollments),
		}

		ctx.JSON(200, response)
	}
}
