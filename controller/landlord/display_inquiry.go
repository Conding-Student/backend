package controller

import (
	"time"

	"intern_template_v1/middleware"
	"intern_template_v1/model/response"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// 🔐 Extract UID from JWT Token
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

// ✅ Fetch inquiries with tenant full name and apartment name
func FetchInquiriesByLandlord(c *fiber.Ctx) error {
	// Get landlord UID from JWT token
	uid, err := GetUIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Unauthorized: Missing or invalid token",
			Data:    nil,
		})
	}

	// Struct to hold inquiry + tenant and apartment info
	var inquiries []struct {
		ID            uint      `json:"id"`
		UID           string    `json:"uid"`
		ApartmentID   uint      `json:"apartment_id"`
		ApartmentName string    `json:"apartment_name"`
		Message       string    `json:"message"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
		ExpiresAt     time.Time `json:"expires_at"`
		Notified      bool      `json:"notified"`
		TenantEmail   string    `json:"tenant_email"`
		FullName      string    `json:"full_name"`
	}

	// Query to join users and apartments
	err = middleware.DBConn.Table("inquiries").
		Select("inquiries.id, inquiries.uid, inquiries.apartment_id, apartments.property_name AS apartment_name, inquiries.message, inquiries.status, inquiries.created_at, inquiries.expires_at, inquiries.notified, users.email AS tenant_email, users.fullname AS full_name").
		Joins("JOIN users ON users.uid = inquiries.uid").
		Joins("JOIN apartments ON apartments.id = inquiries.apartment_id").
		Where("apartments.uid = ?", uid).
		Find(&inquiries).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch inquiries",
			Data:    nil,
		})
	}

	// Success response
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Inquiries retrieved successfully",
		Data: fiber.Map{
			"inquiries": inquiries,
		},
	})
}
