package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for Inquiry request
type InquiryRequest struct {
	LandlordID  uint   `json:"landlord_id"`
	ApartmentID uint   `json:"apartment_id"`
	Message     string `json:"message"`
}

// ✅ Function to create an inquiry
func CreateInquiry(c *fiber.Ctx) error {
	// 🔍 Extract user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// 🆔 Extract Tenant ID safely from JWT claims
	tenantIDFloat, ok := userClaims["uid"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid user ID in token",
		})
	}
	tenantID := uint(tenantIDFloat) // Convert to uint

	// 📩 Parse request body
	var req InquiryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// 📌 Validate required fields
	if req.LandlordID == 0 || req.ApartmentID == 0 || req.Message == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required fields: landlord_id, apartment_id, or message",
		})
	}

	// 📅 Set expiration time to one week from the current time
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 1 week from now

	// 📝 Create inquiry entry
	inquiry := model.Inquiry{
		TenantID:    tenantID,
		ApartmentID: req.ApartmentID,
		Message:     req.Message,
		Status:      "Pending", // Default status
		CreatedAt:   time.Now(),
		ExpiresAt:   expirationTime, // Set the expiration date
	}

	// 🛠 Save inquiry in DB
	if err := middleware.DBConn.Create(&inquiry).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to create inquiry",
			"error":   err.Error(),
		})
	}

	// ✅ Return success response
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Inquiry sent successfully",
		"data": fiber.Map{
			"inquiry_id":   inquiry.ID,
			"tenant_id":    inquiry.TenantID,
			"landlord_id":  req.LandlordID,
			"apartment_id": inquiry.ApartmentID,
			"message":      inquiry.Message,
			"status":       inquiry.Status,
			"expires_at":   inquiry.ExpiresAt, // Return expiration date in response
		},
	})
}
