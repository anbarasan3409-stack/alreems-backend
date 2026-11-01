package main

import (
	database "backend/database"
	models "backend/models"
	routes "backend/routes"
	controllers "backend/controllers"
	utils "backend/utils"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file")

	}

	port := os.Getenv("PORT")
	router := gin.New()
	router.Use(gin.Logger())

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{os.Getenv("FRONTEND_URL"), os.Getenv("ADMIN_URL"), os.Getenv("TEACHERS_URL")}
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	db := database.Init()

	// Seed Admin user
	var admin models.Admin
	if err := db.Where("email = ?", "sanjaiexplores@gmail.com").First(&admin).Error; err != nil {
		fmt.Println("Admin not found, creating one...")
		admin = models.Admin{
			Name:              "Sanjai",
			Email:             "sanjaiexplores@gmail.com",
			Role:              "admin",
			PhoneNumber:       "",
			Address:           "",
			ProfilePicture:    "",
			GoogleAccessToken: "",
			RefreshToken:      "",
		}
		if createErr := db.Create(&admin).Error; createErr != nil {
			fmt.Println("Failed to create admin:", createErr)
		} else {
			fmt.Println("Admin user created successfully")
		}
	} else {
		fmt.Println("Admin user already exists")
	}

	// Seed student enrollments and payment records for VED-101
	database.SeedStudentEnrollments(db)

	handler := database.Connect(db)

	// Start automated payment reminder + class reminder background schedulers
	utils.StartAllSchedulers(db)

	routes.AuthRoutes(router, handler)
	routes.StudentRoutes(router, handler)
	routes.CoursesRoutes(router, handler)
	routes.AdminRoutes(router, handler)
	routes.TeacherRoutes(router, handler)
	routes.SetupAgoraRoutes(router)

	// Stripe Webhook (must be raw body, handled in controller)
	router.POST("/api/webhook/stripe", controllers.StripeWebhook)

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"Message": "Alreems Backend is Running",
		})
	})

	fmt.Println("Alreems Backend is running on port", port)
	err = router.Run(":" + port)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
