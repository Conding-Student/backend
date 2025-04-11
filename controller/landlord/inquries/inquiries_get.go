package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// UpdateInquiryStatusByLandlord updates an inquiry's status if the landlord owns the apartment
func UpdateInquiryStatusByLandlord(c *fiber.Ctx) error {
	// 🔐 Extract JWT claims
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// 🆔 Extract landlord UID
	landlordUID, ok := userClaims["uid"].(string)
	if !ok || landlordUID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// 📥 Request body structure
	type Request struct {
		InquiryID uint   `json:"inquiry_id"`
		Status    string `json:"status"` // should be "Accepted" or "Rejected"
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if req.Status != "Accepted" && req.Status != "Rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid status: must be 'Accepted' or 'Rejected'",
		})
	}

	// 🧠 Check if the inquiry belongs to a property owned by the landlord
	var inquiry model.Inquiry
	if err := middleware.DBConn.
		Joins("JOIN apartments ON inquiries.apartment_id = apartments.id").
		Where("inquiries.id = ? AND apartments.uid = ?", req.InquiryID, landlordUID).
		First(&inquiry).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Inquiry not found or does not belong to your property",
			"error":   err.Error(),
		})
	}

	// ✅ Update the inquiry status
	inquiry.Status = req.Status
	if err := middleware.DBConn.Save(&inquiry).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error while updating inquiry status",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Inquiry status updated successfully",
		"status":  req.Status,
	})
}
