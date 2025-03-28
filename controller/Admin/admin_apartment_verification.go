package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
)

// Get pending status
func GetPendingApartments(c *fiber.Ctx) error {
	var pendingApartments []model.Apartment

	// Fetch apartments where status is "Pending"
	result := middleware.DBConn.Where("status = ?", "Pending").Find(&pendingApartments)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pending apartments",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "Pending apartments retrieved successfully",
		"apartments": pendingApartments,
	})
}

type VerifyApartmentRequest struct {
	Status string `json:"status"` // Expected values: "Approved" or "Rejected"
}

// ✅ Verify (Approve/Reject) an Apartment
func VerifyApartment(c *fiber.Ctx) error {
	apartmentID := c.Params("id") // Get apartment ID from the URL
	var req VerifyApartmentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Check if the provided status is valid
	if req.Status != "Approved" && req.Status != "Rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Use 'Approved' or 'Rejected'.",
		})
	}

	// Update apartment status in the database
	result := middleware.DBConn.Model(&model.Apartment{}).
		Where("id = ?", apartmentID).
		Update("status", req.Status)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update apartment status",
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Apartment status updated successfully",
		"apartment_id": apartmentID,
		"status":       req.Status,
	})
}
