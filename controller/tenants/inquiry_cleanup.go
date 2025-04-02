package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Inquiry notification function that sends reminders to tenants with pending inquiries
func NotifyPendingInquiries(c *fiber.Ctx) error {
	// Get the current time in Manila timezone (Asia/Manila)
	currentTime := time.Now().In(time.FixedZone("Asia/Manila", 8*60*60)) // UTC +8

	// Fetch inquiries that have been pending for exactly 1 week and have not been notified
	var inquiries []model.Inquiry
	if err := middleware.DBConn.
		Where("status = ? AND created_at <= ? AND notified = ?", "Pending", currentTime.Add(-7*24*time.Hour), false).
		Find(&inquiries).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch inquiries",
			"error":   err.Error(),
		})
	}

	// List to hold notifications to return
	var notifications []fiber.Map

	// Loop through the inquiries to send notifications
	for _, inquiry := range inquiries {
		// Fetch tenant details based on Tenant ID
		var tenant model.User
		if err := middleware.DBConn.
			Where("id = ?", inquiry.TenantID).
			First(&tenant).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to fetch tenant details",
				"error":   err.Error(),
			})
		}

		// Create the notification message for the tenant
		message := "Your inquiry for apartment ID " + strconv.Itoa(int(inquiry.ApartmentID)) + " is still pending due to the landlord's inactivity. We recommend you try applying for other available apartments."

		// Append notification for response
		notifications = append(notifications, fiber.Map{
			"tenant_id":    tenant.ID,
			"tenant_name":  tenant.FirstName + " " + tenant.LastName,
			"tenant_email": tenant.Email,
			"apartment_id": inquiry.ApartmentID,
			"message":      message,
		})

		// Update the `notified` column to `true` so the tenant won't be notified again for this inquiry
		if err := middleware.DBConn.Model(&inquiry).Update("notified", true).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error while updating notification status",
				"error":   err.Error(),
			})
		}
	}

	// Return the notifications as part of the response
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":       "Pending inquiry notifications sent successfully",
		"notifications": notifications,
	})
}
