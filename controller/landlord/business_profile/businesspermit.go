package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for parsing business permit image URL request
type BusinessPermitRequest struct {
	BusinessPermit string `json:"business_permit_image_url"` // URL of the business permit image
}

// ✅ Function to insert/update business permit image URL of the landlord
func SetUpdateBusinessPermitImage(c *fiber.Ctx) error {
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
	var req BusinessPermitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// 📌 Validate required fields
	if req.BusinessPermit == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required field: business_permit_image_url",
		})
	}

	// 🏠 Find or create landlord profile based on UID
	var landlordProfile model.LandlordProfile
	if err := middleware.DBConn.Where("uid = ?", uid).First(&landlordProfile).Error; err != nil {
		// If no profile exists, create a new one
		if err := middleware.DBConn.Create(&model.LandlordProfile{
			Uid:            uid,
			BusinessPermit: req.BusinessPermit,
		}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to create landlord profile",
				"error":   err.Error(),
			})
		}
	} else {
		// If profile exists, update the business permit image URL
		landlordProfile.BusinessPermit = req.BusinessPermit
		if err := middleware.DBConn.Save(&landlordProfile).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to update landlord profile",
				"error":   err.Error(),
			})
		}
	}

	// 🎉 Success Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Business permit image URL updated successfully",
		"data": fiber.Map{
			"uid":             uid,
			"business_permit": req.BusinessPermit,
		},
	})
}
