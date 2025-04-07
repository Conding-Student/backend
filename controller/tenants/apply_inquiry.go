package controller

import (
	"intern_template_v1/middleware"
	//"intern_template_v1/model"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct to capture incoming inquiry data from the request body
type CreateInquiryRequest struct {
	ApartmentID int    `json:"apartment_id"`
	Message     string `json:"message"`
}

// Function to extract the UID from the JWT token
func GetUIDFromToken(c *fiber.Ctx) (string, error) {
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return "", fiber.ErrUnauthorized
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return "", fiber.ErrUnauthorized
	}

	return uid, nil
}

func CreateInquiry(c *fiber.Ctx) error {
	// ✅ Extract UID from JWT Token
	uid, err := GetUIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing or invalid JWT",
		})
	}

	// ✅ Parse request body
	var req CreateInquiryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ✅ Check if the tenant already submitted an inquiry for the same apartment
	var count int64
	err = middleware.DBConn.
		Table("inquiries").
		Where("uid = ? AND apartment_id = ?", uid, req.ApartmentID).
		Count(&count).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error during duplicate check",
			"error":   err.Error(),
		})
	}

	if count > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "You’ve already submitted an inquiry for this apartment.",
		})
	}

	// ✅ Create inquiry
	currentTime := time.Now()
	expirationTime := currentTime.Add(7 * 24 * time.Hour)

	query := `INSERT INTO inquiries (uid, apartment_id, message, status, created_at, expires_at, notified) 
	          VALUES (?, ?, ?, 'Pending', ?, ?, false)`

	if err := middleware.DBConn.Exec(query, uid, req.ApartmentID, req.Message, currentTime, expirationTime).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to create inquiry",
			"error":   err.Error(),
		})
	}

	// ✅ Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Inquiry created successfully",
	})
}
