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

// Verify (Approve/Reject) an Apartment
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

	var apartment model.Apartment
	result := middleware.DBConn.First(&apartment, apartmentID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Apartment not found",
		})
	}

	// 🔥 If status is "Rejected", DELETE the apartment
	if req.Status == "Rejected" {
		deleteResult := middleware.DBConn.Delete(&apartment)
		if deleteResult.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete apartment",
			})
		}
		return c.JSON(fiber.Map{
			"message": "Apartment rejected and removed successfully",
		})
	}

	// Otherwise, just update the status
	apartment.Status = req.Status
	middleware.DBConn.Save(&apartment)

	return c.JSON(fiber.Map{
		"message":      "Apartment status updated successfully",
		"apartment_id": apartmentID,
		"status":       req.Status,
	})
}
