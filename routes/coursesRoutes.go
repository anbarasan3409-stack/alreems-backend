package routes

import (
	"backend/controllers"
	"backend/database"

	"github.com/gin-gonic/gin"
)

func CoursesRoutes(router *gin.Engine, handler database.Handler) {
	coursesGroup := router.Group("/api/courses")
	coursesGroup.GET("/", controllers.GetAllCourses(handler))
	coursesGroup.GET("/:id", controllers.GetCourseByID(handler))
}
