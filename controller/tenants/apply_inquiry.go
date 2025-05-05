// Controller
package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"time"

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
            "id":           inquiry.ID,
            "property_id":  inquiry.PropertyID,
            "expires_at":   inquiry.ExpiresAt.Format(time.RFC3339),
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
        LandlordUID:    property.Uid,
        PropertyID:     req.PropertyID,
        InitialMessage: req.Message,
        PreferredVisit: visitTime,
        CreatedAt:      time.Now(),
        ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // 1 week expiration
    }

    return inquiry, middleware.DBConn.Create(inquiry).Error
}