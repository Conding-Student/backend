package controller

import (
	"fmt"
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// DeleteInquiryAfterViewingNotification deletes the inquiry after tenant views rejection notification
func DeleteInquiryAfterViewingNotification(c *fiber.Ctx) error {
	// Extract JWT claims
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	tenantUID, ok := userClaims["uid"].(string)
	if !ok || tenantUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid tenant UID",
		})
	}

	// 📥 Request body to handle inquiry ID
	type Request struct {
		InquiryID uint `json:"inquiry_id"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Retrieve the inquiry related to this tenant
	var inquiry model.Inquiry
	if err := middleware.DBConn.Where("id = ? AND uid = ?", req.InquiryID, tenantUID).First(&inquiry).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Inquiry not found or does not belong to this tenant",
			"error":   err.Error(),
		})
	}

	// Check if inquiry is either Rejected or Pending and expired
	// currentTime := time.Now()
	// if inquiry.Status == "Rejected" || (inquiry.Status == "Pending" && inquiry.ExpiresAt.Before(currentTime)) {
	// 	// Mark the inquiry as notified by setting "notified" to true
	// 	inquiry.Notified = true
	// 	if err := middleware.DBConn.Save(&inquiry).Error; err != nil {
	// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 			"message": "Error updating inquiry status",
	// 			"error":   err.Error(),
	// 		})
	// 	}

	// 	// Delete the inquiry from the database
	// 	if err := middleware.DBConn.Delete(&inquiry).Error; err != nil {
	// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 			"message": "Error deleting inquiry",
	// 			"error":   err.Error(),
	// 		})
	// 	}

	// 	// 🎉 Return success response
	// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
	// 		"message": "Inquiry successfully deleted after viewing rejection notification or expiration",
	// 	})
	// }

	// If inquiry is not Rejected or expired, it cannot be deleted
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "Only rejected or expired inquiries can be deleted",
	})
}

func DeleteExpiredInquiries() {
	// Run this function periodically, e.g., every hour
	for {
		// Get the current time
		currentTime := time.Now()

		// Query all inquiries where the expiration date has passed and the status is "Pending"
		var expiredInquiries []model.Inquiry
		if err := middleware.DBConn.Where("expires_at < ? AND status = ?", currentTime, "Pending").Find(&expiredInquiries).Error; err != nil {
			fmt.Printf("Error fetching expired inquiries: %v\n", err)
			return
		}

		// Loop through all expired inquiries
		// for _, inquiry := range expiredInquiries {
		// 	// If the inquiry is expired and notified is true, delete it
		// 	if inquiry.Notified == true {
		// 		// Delete the inquiry immediately
		// 		if err := middleware.DBConn.Delete(&inquiry).Error; err != nil {
		// 			fmt.Printf("Error deleting expired inquiry ID: %d, Error: %v\n", inquiry.ID, err)
		// 		} else {
		// 			fmt.Printf("Deleted expired inquiry ID: %d\n", inquiry.ID)
		// 		}
		// 	} else {
		// 		// If the inquiry is expired but not notified, don't delete it yet
		// 		fmt.Printf("Inquiry ID: %d is expired but not yet notified\n", inquiry.ID)
		// 	}
		// }

		// Wait for some time before running again, e.g., every 1 hour
		time.Sleep(1 * time.Hour)
	}
}

func CountAcceptedOrRejectedInquiries(c *fiber.Ctx) error {
	// Extract JWT claims to get tenant UID
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	tenantUID, ok := userClaims["uid"].(string)
	if !ok || tenantUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid tenant UID",
		})
	}

	// Count the number of inquiries with status "Accepted" or "Rejected" for this tenant
	var count int64
	if err := middleware.DBConn.
		Model(&model.Inquiry{}).
		Where("uid = ? AND status IN ?", tenantUID, []string{"Accepted", "Rejected"}).
		Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to count inquiries",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"uid":                  tenantUID,
		"accepted_or_rejected": count,
	})
}

// GetAllinquiries retrieves all inquiries for the tenant
// with status "Accepted", "Rejected", or "Pending"	
func GetAllinquiries(c *fiber.Ctx) error {
	// Extract JWT claims to get tenant UID
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	tenantUID, ok := userClaims["uid"].(string)
	if !ok || tenantUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid tenant UID",
		})
	}

	// Define a struct to hold the result of the query
	type InquiryResponse struct {
		LandlordName   string `json:"landlord_name"`
		LandlordPhoto  string `json:"landlord_photo"`
		InquiryMessage string `json:"inquiry_message"`
		InquiryStatus  string `json:"inquiry_status"`
	}

	// Execute the query to get inquiries with status "Accepted" or "Rejected"
	var inquiries []InquiryResponse
	if err := middleware.DBConn.
		Raw(`
			SELECT 
				u.fullname AS landlord_name,
				u.photo_url AS landlord_photo,
				i.message AS inquiry_message,
				i.status AS inquiry_status
			FROM inquiries i
			JOIN apartments a ON i.apartment_id = a.id
			JOIN users u ON a.uid = u.uid
			WHERE i.status IN ('Rejected', 'Accepted', 'Pending')
			  AND i.uid = ? 
			  AND u.user_type = 'Landlord'`, tenantUID).
		Scan(&inquiries).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve inquiries",
			"error":   err.Error(),
		})
	}

	// Static message for Approved and Rejected inquiries
	for i := range inquiries {
		if inquiries[i].InquiryStatus == "Accepted" {
			inquiries[i].InquiryMessage = "Your inquiry has been approved."
		} else if inquiries[i].InquiryStatus == "Rejected" {
			inquiries[i].InquiryMessage = "Your inquiry has been rejected."
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"tenant_uid": tenantUID,
		"inquiries":  inquiries,
	})
}
