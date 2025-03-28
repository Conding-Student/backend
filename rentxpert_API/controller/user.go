package controller

import (
	"log"

	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"intern_template_v1/model/response"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// GetUserProfile retrieves user profile based on JWT token and stores UserID in context
func GetUserProfile(c *fiber.Ctx) error {
	// Get user claims from JWT stored in middleware
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Extract user ID from JWT claims
	userID := uint(userClaims["id"].(float64))

	// ✅ Store UserID in context
	c.Locals("userID", userID)

	// Fetch user details from the database
	var user model.User
	result := middleware.DBConn.Where("id = ?", userID).First(&user)

	if result.Error != nil {
		log.Println("[ERROR] Failed to fetch user profile:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Could not retrieve user profile",
			Data:    nil,
		})
	}

	// ✅ Return user profile
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User profile retrieved successfully",
		Data: fiber.Map{
			"id":         user.ID,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
			"address":    user.Address,
			"user_type":  user.UserType,
		},
	})
}
