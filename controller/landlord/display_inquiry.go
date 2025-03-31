package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ✅ Fetch inquiries for a landlord
func FetchInquiriesByLandlord(c *fiber.Ctx) error {
	// 📌 Retrieve landlord ID from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: User claims missing",
		})
	}

	landlordIDFloat, exists := userClaims["id"].(float64)
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid or missing user ID in token",
		})
	}
	landlordID := uint(landlordIDFloat) // Convert float64 to uint

	// 📌 Fetch inquiries linked to apartments owned by this landlord
	var inquiries []struct {
		model.Inquiry
		TenantName  string `json:"tenant_name"`
		PhoneNumber string `json:"phone_number"`
	}

	if err := middleware.DBConn.Table("inquiries").
		Select("inquiries.*, users.first_name || ' ' || users.last_name AS tenant_name, users.phone_number").
		Joins("JOIN users ON users.id = inquiries.tenant_id").
		Joins("JOIN apartments ON apartments.id = inquiries.apartment_id").
		Where("apartments.user_id = ?", landlordID).
		Find(&inquiries).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch inquiries",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":   "Inquiries retrieved successfully",
		"inquiries": inquiries,
	})
}

// ✅ Get all inquiries with "Pending" status
func GetPendingInquiries(c *fiber.Ctx) error {
	var pendingInquiries []model.Inquiry

	// Fetch inquiries where status is "Pending"
	result := middleware.DBConn.Where("status = ?", "Pending").Find(&pendingInquiries)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pending inquiries",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Pending inquiries retrieved successfully",
		"inquiries": pendingInquiries,
	})
}

// ✅ Struct to parse inquiry status update request
type UpdateInquiryStatusRequest struct {
	Status string `json:"status"` // Expected values: "Responded" or "Expired"
}

// ✅ Update inquiry status (Responded/Expired)
func UpdateInquiryStatus(c *fiber.Ctx) error {
	inquiryID := c.Params("id") // Get inquiry ID from URL
	var req UpdateInquiryStatusRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate status input
	if req.Status != "Responded" && req.Status != "Expired" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Use 'Responded' or 'Expired'.",
		})
	}

	// Update inquiry status in database
	result := middleware.DBConn.Model(&model.Inquiry{}).
		Where("id = ?", inquiryID).
		Updates(map[string]interface{}{
			"status": req.Status,
		})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update inquiry status",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "Inquiry status updated successfully",
		"inquiry_id": inquiryID,
		"status":     req.Status,
	})
}
