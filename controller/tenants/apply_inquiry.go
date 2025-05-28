// Controller
package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type CreateInquiryRequest struct {
	PropertyID     uint   `json:"property_id" validate:"required"`
	Message        string `json:"message" validate:"required,min=10"`
	PreferredVisit string `json:"preferred_visit,omitempty"` // Optional ISO8601
}

func CreateInquiry(c *fiber.Ctx) error {
	// 1. Authentication
	tenantUID, err := GetUIDFromToken(c)

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// 2. Request Validation
	var req CreateInquiryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// 3. Duplicate Check
	if exists, err := checkDuplicateInquiry(tenantUID, req.PropertyID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "System error",
		})
	} else if exists {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Existing inquiry for this property",
		})
	}

	// 4. Create Inquiry
	inquiry, err := createInquiryRecord(tenantUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create inquiry",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"data": fiber.Map{
			"id":          inquiry.ID,
			"property_id": inquiry.PropertyID,
			"expires_at":  inquiry.ExpiresAt.Format(time.RFC3339),
		},
	})
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

func checkDuplicateInquiry(tenantUID string, propertyID uint) (bool, error) {
	var count int64
	err := middleware.DBConn.Model(&model.Inquiry{}).
		Where("tenant_uid = ? AND property_id = ?", tenantUID, propertyID).
		Count(&count).Error
	return count > 0, err
}

func createInquiryRecord(tenantUID string, req CreateInquiryRequest) (*model.Inquiry, error) {
	// Get landlord UID from property
	var property model.Apartment
	if err := middleware.DBConn.First(&property, req.PropertyID).Error; err != nil {
		return nil, err
	}

	// Parse optional visit time
	var visitTime *time.Time
	if req.PreferredVisit != "" {
		if t, err := time.Parse(time.RFC3339, req.PreferredVisit); err == nil {
			visitTime = &t
		}
	}

	inquiry := &model.Inquiry{
		TenantUID:      tenantUID,
		PropertyID:     req.PropertyID,
		InitialMessage: req.Message,
		PreferredVisit: visitTime,
		CreatedAt:      time.Now(),
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 1 week expiration
	}

	return inquiry, middleware.DBConn.Create(inquiry).Error
}

func HasInquiryToApartment(tenantUID string, apartmentID uint) (bool, error) {
	var count int64
	err := middleware.DBConn.Model(&model.Inquiry{}).
		Where("tenant_uid = ? AND property_id = ?", tenantUID, apartmentID).
		Count(&count).Error
	return count > 0, err
}

func CheckHasInquiry(c *fiber.Ctx) error {
	fmt.Println("🔍 Checking for existing inquiry...")

	// Step 1: Auth check
	tenantUID, err := GetUIDFromToken(c)
	if err != nil {
		fmt.Println("❌ Unauthorized: Failed to extract UID from token:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}
	fmt.Println("✅ Tenant UID:", tenantUID)

	// Step 2: Parse query param
	apartmentIDStr := c.Query("apartment_id")
	if apartmentIDStr == "" {
		fmt.Println("❌ Missing apartment_id query param")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "apartment_id is required",
		})
	}

	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil || apartmentID <= 0 {
		fmt.Println("❌ Invalid apartment_id:", apartmentIDStr)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid apartment_id",
		})
	}
	fmt.Println("✅ Apartment ID:", apartmentID)

	// Step 3: Check inquiry in DB
	var count int64
	err = middleware.DBConn.Model(&model.Inquiry{}).
		Where("tenant_uid = ? AND property_id = ?", tenantUID, apartmentID).
		Count(&count).Error

	if err != nil {
		fmt.Println("❌ DB error while checking inquiry:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "System error",
		})
	}

	fmt.Printf("📊 Inquiry count for tenant %s and property %d: %d\n", tenantUID, apartmentID, count)

	// Step 4: Return result
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"has_inquiry": count > 0,
	})
}
