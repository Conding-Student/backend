package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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

// UpdateAccountStatus function to update the user's account status to 'Landlord'
func UpdateUserType(c *fiber.Ctx) error {
	// Extract the UID from the request (you can modify this part based on your request)
	uid := c.Params("uid") // Assuming UID is passed in the URL as a parameter

	// Query the database to find the user by UID
	var user model.User
	if err := middleware.DBConn.Where("uid = ?", uid).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	// Update the account status to 'Landlord'
	user.UserType = "Landlord"
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating account status",
			"error":   err.Error(),
		})
	}

	// Return success response with updated user data
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Account status updated successfully",
		"user": fiber.Map{
			"uid":            user.Uid,
			"account_status": user.UserType,
		},
	})
}
