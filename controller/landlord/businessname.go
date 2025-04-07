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

// ✅ Function to insert/update business name of the landlord
func SetBusinessName(c *fiber.Ctx) error {
	// 🔍 Extract user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// 🆔 Extract Landlord Uid safely from JWT claims
	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// 📩 Parse request body
	var req BusinessNameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// 📌 Validate required fields
	if req.BusinessName == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required field: business_name",
		})
	}

	// 🏠 Fetch landlord profile based on UID
	var landlordProfile model.LandlordProfile
	if err := middleware.DBConn.Where("uid = ?", uid).First(&landlordProfile).Error; err != nil {
		// If landlord profile doesn't exist, create a new one
		if err.Error() == "record not found" {
			landlordProfile = model.LandlordProfile{
				Uid:          uid,
				BusinessName: req.BusinessName,
			}
			if err := middleware.DBConn.Create(&landlordProfile).Error; err != nil {
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"message": "Database error: Unable to create landlord profile",
					"error":   err.Error(),
				})
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to fetch landlord profile",
				"error":   err.Error(),
			})
		}
	} else {
		// If landlord profile exists, update business name
		landlordProfile.BusinessName = req.BusinessName
		if err := middleware.DBConn.Save(&landlordProfile).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to update landlord profile",
				"error":   err.Error(),
			})
		}
	}

	// 🎉 Success Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "Business name updated successfully",
		"business_name": landlordProfile.BusinessName,
	})
}
