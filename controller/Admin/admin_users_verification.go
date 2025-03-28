package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
)

// check pending landlords
func GetPendingUsers(c *fiber.Ctx) error {
	var pendingUsers []model.User

	// Fetch users where status is "Pending"
	result := middleware.DBConn.Where("account_status = ?", "Pending").Find(&pendingUsers)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pending landlords",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Pending users retrieved successfully",
		"users":   pendingUsers,
	})
}

type VerifyLandlordRequest struct {
	Status string `json:"account_status"` // Expected values: "Approved" or "Rejected"
}

// ✅ Verify (Approve/Reject) a user
func VerifyUsers(c *fiber.Ctx) error {
	UserID := c.Params("id") // Get user ID from the URL
	var req VerifyLandlordRequest

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

	// Update landlord status in the database
	result := middleware.DBConn.Model(&model.User{}).
		Where("id = ?", UserID).
		Update("account_status", req.Status)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update User account status",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "User account status updated successfully",
		"users_id": UserID,
		"status":   req.Status,
	})
}
