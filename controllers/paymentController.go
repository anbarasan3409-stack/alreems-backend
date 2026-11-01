package controllers

import (
	"backend/database"
	"backend/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"github.com/stripe/stripe-go/v74/product"
	"github.com/stripe/stripe-go/v74/webhook"
)

// EnsureStripeCustomer checks if a student has a Stripe customer ID, if not or invalid creates one
func EnsureStripeCustomer(student *models.Student) error {
	stripeKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeKey != "" {
		stripe.Key = stripeKey
	}

	if student.StripeCustomerID != "" {
		// Verify customer exists in the active Stripe account (e.g. not a leftover test-mode customer ID)
		c, err := customer.Get(student.StripeCustomerID, nil)
		if err == nil && c != nil && !c.Deleted {
			return nil
		}
		log.Printf("Stripe customer %s not found in active Stripe account (%v). Recreating customer...", student.StripeCustomerID, err)
		student.StripeCustomerID = ""
	}

	params := &stripe.CustomerParams{
		Email: stripe.String(student.Email),
		Name:  stripe.String(student.Name),
	}
	c, err := customer.New(params)
	if err != nil {
		log.Printf("Error creating Stripe customer for student %d: %v", student.ID, err)
		return err
	}

	student.StripeCustomerID = c.ID
	db := database.Init()
	db.Save(student)
	return nil
}

// GetOrCreateCoursePrice creates a Stripe product and price for the course if needed in AED
func GetOrCreateCoursePrice(course *models.Course) (string, error) {
	productName := fmt.Sprintf("Course %d: %s", course.ID, course.Name)

	// Price in AED (default to 150 AED if not set or <= 0)
	coursePriceAED := course.Price
	if coursePriceAED <= 0 {
		coursePriceAED = 150.0
	}
	// 1 AED = 100 fils
	targetUnitAmount := int64(coursePriceAED * 100)

	params := &stripe.ProductListParams{}
	params.Filters.AddFilter("active", "", "true")
	i := product.List(params)
	var foundProductID string
	for i.Next() {
		p := i.Product()
		if p.Name == productName {
			foundProductID = p.ID
			// Check if there is an active price with AED and matching unit amount
			priceParams := &stripe.PriceListParams{Product: stripe.String(p.ID)}
			priceParams.Filters.AddFilter("active", "", "true")
			pi := price.List(priceParams)
			for pi.Next() {
				pr := pi.Price()
				if strings.ToLower(string(pr.Currency)) == "aed" && pr.UnitAmount == targetUnitAmount {
					return pr.ID, nil
				}
			}
			break
		}
	}

	// Create Product if it didn't exist
	if foundProductID == "" {
		pParams := &stripe.ProductParams{
			Name: stripe.String(productName),
		}
		p, err := product.New(pParams)
		if err != nil {
			log.Printf("Error creating Stripe product: %v", err)
			return "", err
		}
		foundProductID = p.ID
	}

	// Create Price: in AED / month
	priceParams := &stripe.PriceParams{
		Currency:   stripe.String(string(stripe.CurrencyAED)),
		UnitAmount: stripe.Int64(targetUnitAmount),
		Recurring: &stripe.PriceRecurringParams{
			Interval: stripe.String("month"),
		},
		Product: stripe.String(foundProductID),
	}
	pr, err := price.New(priceParams)
	if err != nil {
		log.Printf("Error creating Stripe price: %v", err)
		return "", err
	}

	return pr.ID, nil
}

func CreateCheckoutSession(c *gin.Context) {
	// Set Stripe API key from environment variables
	stripeKey := strings.TrimSpace(os.Getenv("STRIPE_SECRET_KEY"))
	if stripeKey == "" {
		log.Printf("STRIPE_SECRET_KEY is empty")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment gateway is not configured (missing STRIPE_SECRET_KEY). Please configure a valid secret key."})
		return
	}
	if strings.HasPrefix(stripeKey, "pk_") {
		log.Printf("STRIPE_SECRET_KEY starts with 'pk_', which is a Publishable Key. Stripe requires a Secret Key starting with 'sk_' (e.g. sk_live_... or sk_test_...) on the server.")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stripe misconfiguration: STRIPE_SECRET_KEY must be a secret key (sk_...), but a publishable key (pk_...) was provided."})
		return
	}
	stripe.Key = stripeKey

	studentID, _ := strconv.Atoi(c.Param("id"))
	
	// Read request body for course ID
	var req struct {
		CourseID uint `json:"course_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	db := database.Init()

	var student models.Student
	if err := db.First(&student, studentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	var course models.Course
	if err := db.First(&course, req.CourseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}

	// Ensure Stripe Customer
	if err := EnsureStripeCustomer(&student); err != nil {
		log.Printf("EnsureStripeCustomer failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Stripe customer: %v", err)})
		return
	}

	// Ensure Stripe Price
	priceID, err := GetOrCreateCoursePrice(&course)
	if err != nil {
		log.Printf("GetOrCreateCoursePrice failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create Stripe price: %v", err)})
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "https://reems.co.in"
	}

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(student.StripeCustomerID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(fmt.Sprintf("%s/dashboard?session_id={CHECKOUT_SESSION_ID}&course_id=%d", frontendURL, course.ID)),
		CancelURL:  stripe.String(fmt.Sprintf("%s/dashboard/course/%d", frontendURL, course.ID)),
		ClientReferenceID: stripe.String(fmt.Sprintf("%d_%d", student.ID, course.ID)),
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("Stripe session.New failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create checkout session: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": s.URL})
}

func StripeWebhook(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	endpointSecret := strings.TrimSpace(os.Getenv("STRIPE_WEBHOOK_SECRET"))
	signatureHeader := c.GetHeader("Stripe-Signature")

	event, err := webhook.ConstructEventWithOptions(payload, signatureHeader, endpointSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		log.Printf("⚠️  Webhook signature verification failed. %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature verification failed"})
		return
	}

	db := database.Init()

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			c.Status(http.StatusBadRequest)
			return
		}

		// ClientReferenceID is "StudentID_CourseID"
		var studentID, courseID uint
		fmt.Sscanf(session.ClientReferenceID, "%d_%d", &studentID, &courseID)

		// Create Subscription Record
		sub := models.Subscription{
			StudentID:            studentID,
			CourseID:             courseID,
			StripeSubscriptionID: session.Subscription.ID,
			StripeCustomerID:     session.Customer.ID,
			Status:               "active",
			CurrentPeriodStart:   time.Now(),
			CurrentPeriodEnd:     time.Now().AddDate(0, 1, 0), // Default 1 month, will be updated by invoice
		}
		db.Create(&sub)

		// Also register the student in the Enrollment table if not already
		var count int64
		db.Model(&models.Enrollment{}).Where("student_id = ? AND course_id = ?", studentID, courseID).Count(&count)
		if count == 0 {
			enrollment := models.Enrollment{
				StudentID: studentID,
				CourseID:  courseID,
				Progress:  0,
			}
			db.Create(&enrollment)
		}

	case "invoice.payment_succeeded":
		var invoice stripe.Invoice
		err := json.Unmarshal(event.Data.Raw, &invoice)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			c.Status(http.StatusBadRequest)
			return
		}

		// Update subscription end date
		if invoice.Subscription != nil {
			var sub models.Subscription
			if err := db.Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&sub).Error; err == nil {
				sub.Status = "active"
				// Invoice period end is unix timestamp
				sub.CurrentPeriodEnd = time.Unix(invoice.PeriodEnd, 0)
				db.Save(&sub)
			}
		}

	case "customer.subscription.deleted", "customer.subscription.updated":
		var stripeSub stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &stripeSub)
		if err != nil {
			log.Printf("Error parsing webhook JSON: %v\n", err)
			c.Status(http.StatusBadRequest)
			return
		}

		var sub models.Subscription
		if err := db.Where("stripe_subscription_id = ?", stripeSub.ID).First(&sub).Error; err == nil {
			sub.Status = string(stripeSub.Status)
			sub.CurrentPeriodEnd = time.Unix(stripeSub.CurrentPeriodEnd, 0)
			sub.CancelAtPeriodEnd = stripeSub.CancelAtPeriodEnd
			db.Save(&sub)
		}
	}

	c.Status(http.StatusOK)
}

// GetStudentSubscriptions returns active subscriptions for a student
func GetStudentSubscriptions(c *gin.Context) {
	studentID, _ := strconv.Atoi(c.Param("id"))
	db := database.Init()

	var subs []models.Subscription
	db.Where("student_id = ?", studentID).Find(&subs)

	c.JSON(http.StatusOK, subs)
}

// GetAllSubscriptions returns all subscriptions (for Admin)
func GetAllSubscriptions(c *gin.Context) {
	db := database.Init()
	var subs []models.Subscription
	db.Find(&subs)
	c.JSON(http.StatusOK, subs)
}
