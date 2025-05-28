package controller

import (
	"net/http"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for parsing valid ID URL request
type ValidIDRequest struct {
	ValidID string `json:"valid_id"` // URL of the valid ID image
}

// ✅ Function to insert/update valid ID for the user based on UID
func SetValidID(c *fiber.Ctx) error {
	// 🔍 Extract user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// 🆔 Extract User UID safely from JWT claims
	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid user UID",
		})
	}

	// 📩 Parse request body
	var req ValidIDRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// 📌 Validate required fields
	if req.ValidID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required field: valid_id",
		})
	}

	// 🏠 Find the user by UID
	var user model.User
	if err := middleware.DBConn.Where("uid = ?", uid).First(&user).Error; err != nil {
		// If the user does not exist, return an error
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
			"error":   err.Error(),
		})
	}

	// 🔄 Update the user's valid ID image URL
	user.ValidID = req.ValidID
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to update valid ID",
			"error":   err.Error(),
		})
	}

	// 🎉 Success Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Valid ID updated successfully",
		"data": fiber.Map{
			"uid":      uid,
			"valid_id": req.ValidID,
		},
	})
}
