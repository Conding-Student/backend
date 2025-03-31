package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func FetchInquiriesByLandlord(c *fiber.Ctx) error {
	// ✅ Retrieve user claims from JWT token
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: User claims missing",
		})
	}

	// ✅ Extract user ID (landlord ID)
	landlordIDFloat, exists := userClaims["id"].(float64)
	if !exists {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid or missing user ID in token",
		})
	}
	landlordID := uint(landlordIDFloat) // Convert float64 to uint

	// ✅ Fetch inquiries linked to apartments owned by this landlord
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
