package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 🆕 Updated struct to include birthday
type ContactInfoRequest struct {
	PhoneNumber *string `json:"phone_number"` // Optional
	Address     *string `json:"address"`      // Optional
	Fullname    *string `json:"fullname"`     // Optional
	Birthday    *string `json:"birthday"`     // Optional, in YYYY-MM-DD format (as a string)
}

// ✅ Function to update phone number, address, fullname, and birthday
func UpdateContactInfo(c *fiber.Ctx) error {
	// 🔐 Extract JWT claims
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid UID",
		})
	}

	// 📥 Parse request body
	var req ContactInfoRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// 🚫 Ensure at least one field is provided
	if req.PhoneNumber == nil && req.Address == nil && req.Fullname == nil && req.Birthday == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "At least one field (phone_number, address, fullname, or birthday) must be provided",
		})
	}

	// 🔍 Fetch the existing user
	var user model.User
	if err := middleware.DBConn.Where("uid = ?", uid).First(&user).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "User not found",
			"error":   err.Error(),
		})
	}

	// 🔄 Update only the provided fields
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}
	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.Fullname != nil {
		user.Fullname = *req.Fullname
	}
	if req.Birthday != nil {
		// Validate and parse the birthday field (if provided)
		parsedBirthday, err := time.Parse("2006-01-02", *req.Birthday)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid birthday format. Use YYYY-MM-DD.",
			})
		}
		user.Birthday = parsedBirthday
	}

	// 💾 Save updates
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user contact info",
			"error":   err.Error(),
		})
	}

	// ✅ Response with updated info
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Contact information updated successfully",
		"phone_number": user.PhoneNumber,
		"address":      user.Address,
		"fullname":     user.Fullname,
		"birthday":     user.Birthday.Format("2006-01-02"), // Format birthday back to string
	})
}
