package controllers

import (
	"backend/database"
	"backend/models"

	"github.com/gin-gonic/gin"
)

func GetAllCourses(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var courses []models.Course
		if err := handler.DB.Preload("Teacher").Find(&courses).Error; err != nil {
			ctx.JSON(500, gin.H{"error": "Failed to fetch courses"})
			return
		}

		// Create custom response with teacher name
		var response []gin.H
		for _, course := range courses {
			courseResponse := gin.H{
				"ID":          course.ID,
				"CreatedAt":   course.CreatedAt,
				"UpdatedAt":   course.UpdatedAt,
				"Name":        course.Name,
				"Code":        course.Code,
				"Description": course.Description,
				"Credits":     course.Credits,
				"Price":       course.Price,
				"TeacherID":   course.TeacherID,
				"TeacherName": course.Teacher.Name,
				"Enrollments": course.Enrollments,
			}
			response = append(response, courseResponse)
		}

		ctx.JSON(200, response)
	}
}

func GetCourseByID(handler database.Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		if id == "" {
			id = ctx.Param("courseId")
		}
		var course models.Course
		if err := handler.DB.Preload("Teacher").First(&course, id).Error; err != nil {
			ctx.JSON(404, gin.H{"error": "Course not found"})
			return
		}

		// Create custom response with teacher name
		courseResponse := gin.H{
			"ID":          course.ID,
			"CreatedAt":   course.CreatedAt,
			"UpdatedAt":   course.UpdatedAt,
			"Name":        course.Name,
			"Code":        course.Code,
			"Description": course.Description,
			"Credits":     course.Credits,
			"Price":       course.Price,
			"TeacherID":   course.TeacherID,
			"TeacherName": course.Teacher.Name,
			"Enrollments": course.Enrollments,
		}

		ctx.JSON(200, courseResponse)
	}
}
