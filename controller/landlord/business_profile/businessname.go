package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for parsing business name request
type BusinessNameRequest struct {
	BusinessName string `json:"business_name"`
}

// ✅ Function to update the business name of the landlord
func UpdateBusinessName(c *fiber.Ctx) error {
	// 🔐 Extract user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// 🆔 Extract Landlord UID safely from JWT claims
	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// 📥 Parse request body for new business name
	var req BusinessNameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	if req.BusinessName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required field: business_name",
		})
	}

	// 🔄 Try updating the existing profile directly
	result := middleware.DBConn.Model(&model.LandlordProfile{}).
		Where("uid = ?", uid).
		Update("business_name", req.BusinessName)

	if result.RowsAffected == 0 {
		// ❗ No existing profile found, so create a new one
		newProfile := model.LandlordProfile{
			Uid:          uid,
			BusinessName: req.BusinessName,
		}
		if err := middleware.DBConn.Create(&newProfile).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to create landlord profile",
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":       "Business name created successfully",
			"business_name": newProfile.BusinessName,
		})
	} else if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to update landlord profile",
			"error":   result.Error.Error(),
		})
	}

	// ✅ Update successful
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Business name updated successfully",
		"business_name": req.BusinessName,
	})
}
