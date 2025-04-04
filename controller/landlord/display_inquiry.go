package controller

import (
	"intern_template_v1/middleware"
	//"intern_template_v1/model"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 🔹 Function to Extract the Landlord UID from JWT Token
func GetUIDFromToken(c *fiber.Ctx) (string, error) {
	// This retrieves the claims from the JWT token
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		// If the user claims can't be extracted, return an error
		return "", fiber.ErrUnauthorized
	}

	// Extract the "uid" value from the claims
	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		// If the "uid" is missing or not valid, return an error
		return "", fiber.ErrUnauthorized
	}

	// Return the landlord's UID from the token
	return uid, nil
}

// / ✅ Fetch inquiries for a landlord
func FetchInquiriesByLandlord(c *fiber.Ctx) error {
	// 1. **Retrieve the Landlord UID from the JWT Token**
	uid, err := GetUIDFromToken(c)
	if err != nil {
		// If the UID is missing or invalid in the token, respond with unauthorized error
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing or invalid JWT",
		})
	}

	// 2. **Define a Structure for Inquiry Data** (Only include tenant_email)
	var inquiries []struct {
		ID          uint      `json:"id"`
		UID         string    `json:"uid"`
		ApartmentID uint      `json:"apartment_id"`
		Message     string    `json:"message"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		ExpiresAt   time.Time `json:"expires_at"`
		Notified    bool      `json:"notified"`
		TenantEmail string    `json:"tenant_email"`
	}

	// 3. **Fetch Inquiries Linked to the Landlord's Apartments**
	if err := middleware.DBConn.Table("inquiries").
		Select("inquiries.id, inquiries.uid, inquiries.apartment_id, inquiries.message, inquiries.status, inquiries.created_at, inquiries.expires_at, inquiries.notified, users.email AS tenant_email").
		Joins("JOIN users ON users.uid = inquiries.uid").
		Joins("JOIN apartments ON apartments.id = inquiries.apartment_id").
		Where("apartments.uid = ?", uid).
		Find(&inquiries).Error; err != nil {
		// If there's an error in fetching the inquiries, return an internal server error with the error message
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch inquiries",
			"error":   err.Error(),
		})
	}

	// 4. **Respond with the Retrieved Inquiries**
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Inquiries retrieved successfully",
		"inquiries": inquiries,
	})
}

// ✅ Struct to parse inquiry status update request
type UpdateInquiryStatusRequest struct {
	Status string `json:"status"` // Expected values: "Responded" or "Expired"
}

// ✅ Update inquiry status for a landlord based on their UID (from URL)
func UpdateInquiryStatusByLandlord(c *fiber.Ctx) error {
	// ✅ Retrieve landlord UID from the URL (e.g., /update-inquiry-status/:uid)
	landlordUID := c.Params("uid")
	if landlordUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Landlord UID is required in the URL",
		})
	}

	// ✅ Get the new status from the request body
	var req struct {
		Status string `json:"status"` // Expected values: 'Responded' or 'Expired'
	}

	// Parse request body to get the status
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate that the status is either 'Responded' or 'Expired'
	if req.Status != "Responded" && req.Status != "Expired" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Use 'Responded' or 'Expired'.",
		})
	}

	// ✅ Construct the SQL query to update the status based on landlord's UID
	query := `
		UPDATE inquiries
		SET status = ?
		WHERE inquiries.uid IN (
			SELECT inquiries.uid
			FROM inquiries
			JOIN apartments ON apartments.id = inquiries.apartment_id
			WHERE apartments.uid = ?
		)
	`

	// Execute the SQL query with the status and landlord's UID
	if err := middleware.DBConn.Exec(query, req.Status, landlordUID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update inquiry status",
			"error":   err.Error(),
		})
	}

	// ✅ Respond with success message
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Inquiry status updated successfully",
	})
}
