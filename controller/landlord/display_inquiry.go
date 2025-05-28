package controller

import (
	"time"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"
	"github.com/Conding-Student/backend/model/response"

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

// Response struct for inquiries with tenant and property info
type InquiryResponse struct {
	ID             uint       `json:"id"`
	TenantUID      string     `json:"tenant_uid"`
	TenantName     string     `json:"tenant_name"`
	TenantEmail    string     `json:"tenant_email"`
	TenantPhotoURL string     `json:"tenant_photo_url"`
	PropertyID     uint       `json:"property_id"`
	PropertyName   string     `json:"property_name"`
	InitialMessage string     `json:"initial_message"`
	PreferredVisit *time.Time `json:"preferred_visit,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
	Status         string     `json:"status"`
}

// ✅ Fetch inquiries with tenant full name and property name
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

	var inquiries []InquiryResponse

	// Query to join users and properties
	err = middleware.DBConn.Model(&model.Inquiry{}).
		Select(`
	inquiries.id,
	inquiries.tenant_uid,
	users.fullname AS tenant_name,
	users.email AS tenant_email,
	users.photo_url AS tenant_photo_url,
	inquiries.property_id,
	apartments.property_name,
	inquiries.initial_message,
	inquiries.preferred_visit,
	inquiries.created_at,
	inquiries.expires_at
`).
		Joins("JOIN users ON users.uid = inquiries.tenant_uid").
		Joins("JOIN apartments ON apartments.id = inquiries.property_id").
		Where("apartments.uid = ?", uid).
		Order("inquiries.created_at DESC").
		Find(&inquiries).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch inquiries: " + err.Error(),
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
