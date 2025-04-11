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

// UpdateAccountStatus function to update the user's account status to 'Verified'
func UpdateAccountStatus(c *fiber.Ctx) error {
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

	// Update the account status to 'Verified'
	user.AccountStatus = "Verified"
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
			"account_status": user.AccountStatus, // Use the updated account_status field
		},
	})
}

func UpdateUserType(c *fiber.Ctx) error {
	// Extract the UID from the request (e.g., /user/type/:uid)
	uid := c.Params("uid")

	// Query the database for the user
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

	// Check if the user is already Verified
	if user.AccountStatus != "Verified" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message":        "User account must be verified to become a landlord",
			"current_status": user.AccountStatus,
		})
	}

	// Proceed to update the user type
	user.UserType = "Landlord"
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error updating user type",
			"error":   err.Error(),
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User type updated to Landlord successfully",
		"user": fiber.Map{
			"uid":            user.Uid,
			"user_type":      user.UserType,
			"account_status": user.AccountStatus,
		},
	})
}

// UpdateAccountStatus function to update the user's account status to 'Tenants'
func UpdateUserTypetenant(c *fiber.Ctx) error {
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
	user.UserType = "Tenants"
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
