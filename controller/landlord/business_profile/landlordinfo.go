package controller

import (
	"net/http"
	"strings"
	"time"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// ✅ Extend request struct to include profile image URL
type ContactInfoRequest struct {
	PhoneNumber *string `json:"phone_number"` // Optional
	Address     *string `json:"address"`      // Optional
	Fullname    *string `json:"fullname"`     // Optional
	Birthday    *string `json:"birthday"`     // Optional, in YYYY-MM-DD format
	ProfilePic  *string `json:"profile_pic"`  // Optional, Cloudinary URL
}

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
	if req.PhoneNumber == nil && req.Address == nil && req.Fullname == nil && req.Birthday == nil && req.ProfilePic == nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "At least one field (phone_number, address, fullname, birthday, or profile_pic) must be provided",
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
		parsedBirthday, err := time.Parse("2006-01-02", *req.Birthday)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid birthday format. Use YYYY-MM-DD.",
			})
		}
		user.Birthday = parsedBirthday
	}
	if req.ProfilePic != nil {
		user.PhotoURL = *req.ProfilePic
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
		"birthday":     user.Birthday.Format("2006-01-02"),
		"profile_pic":  user.PhotoURL,
	})
}

func VerifyLandlordUsingAdmin(c *fiber.Ctx) error {
	// Get UID from params
	uid := c.Params("uid")
	if uid == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing user UID parameter",
		})
	}

	// Start database transaction
	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to start database transaction",
			"error":   tx.Error.Error(),
		})
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get latest landlord profile for the UID
	var landlordProfile model.LandlordProfile
	if err := tx.Where("uid = ?", uid).Order("created_at DESC").First(&landlordProfile).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "No landlord profile found for the provided user UID",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error retrieving landlord profile",
			"error":   err.Error(),
		})
	}

	// New: Prevent verification if profile has existing rejection reason
	if strings.TrimSpace(landlordProfile.RejectionReason) != "" {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "Cannot verify a previously rejected landlord profile",
		})
	}

	// Check if user exists and current status
	var user model.User
	if err := tx.Where("uid = ?", uid).First(&user).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "User not found for this UID",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error checking user",
			"error":   err.Error(),
		})
	}

	// Prevent verification if already a verified landlord
	if user.UserType == "Landlord" && user.AccountStatus == "Verified" {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "User is already a verified landlord",
		})
	}

	// Update user account status and type
	if err := tx.Model(&model.User{}).
		Where("uid = ?", uid).
		Updates(map[string]interface{}{
			"account_status": "Verified",
			"user_type":      "Landlord",
		}).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user status and type",
			"error":   err.Error(),
		})
	}

	// Update landlord profile using the retrieved ID
	updateData := map[string]interface{}{
		"verified_at":      time.Now(),
		"rejection_reason": nil, // Clear any previous rejection
	}

	if err := tx.Model(&model.LandlordProfile{}).
		Where("id = ?", landlordProfile.ID).
		Updates(updateData).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update landlord profile",
			"error":   err.Error(),
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to commit transaction",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Landlord verified successfully",
		"data": fiber.Map{
			"profile_id":     landlordProfile.ID,
			"uid":            uid,
			"account_status": "Verified",
			"user_type":      "Landlord",
			"verified_at":    time.Now().Format(time.RFC3339),
		},
	})
}

// RejectionRequest represents the payload for landlord rejection
type RejectionRequest struct {
	RejectionReason string `json:"rejection_reason" validate:"required,min=1"`
}

func RejectLandlordRequest(c *fiber.Ctx) error {
	// Get user UID from params
	uid := c.Params("uid")
	if uid == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing user UID parameter",
		})
	}

	// Parse rejection reason from request body
	var req RejectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Start database transaction
	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to start database transaction",
			"error":   tx.Error.Error(),
		})
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get latest landlord profile for the UID
	var landlordProfile model.LandlordProfile
	if err := tx.Where("uid = ?", uid).Order("created_at DESC").First(&landlordProfile).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "No landlord profile found for the provided user UID",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error retrieving landlord profile",
			"error":   err.Error(),
		})
	}

	// Check if already rejected to prevent duplicate rejections
	if !landlordProfile.RejectedAt.IsZero() {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "This landlord profile was already rejected",
		})
	}

	// Update user account status to Unverified
	if err := tx.Model(&model.User{}).
		Where("uid = ?", uid).
		Update("account_status", "Unverified").Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user status",
			"error":   err.Error(),
		})
	}

	// Update landlord profile with rejection details
	updateData := map[string]interface{}{
		"rejection_reason": req.RejectionReason,
		"verified_at":      nil,
		"rejected_at":      time.Now(),
	}

	if err := tx.Model(&model.LandlordProfile{}).
		Where("id = ?", landlordProfile.ID).
		Updates(updateData).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update landlord profile",
			"error":   err.Error(),
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to commit transaction",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Landlord registration rejected",
		"data": fiber.Map{
			"profile_id":       landlordProfile.ID,
			"uid":              uid,
			"account_status":   "Unverified",
			"rejection_reason": req.RejectionReason,
			"rejected_at":      time.Now().Format(time.RFC3339),
		},
	})
}

// APARTMENT REJECTION
// RejectionRequest represents the payload for landlord rejection
func RejectApartmentRequest(c *fiber.Ctx) error {
	// Get apartment ID from params
	apartmentID := c.Params("id")
	if apartmentID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing apartment ID parameter",
		})
	}

	// Parse rejection reason from request body
	var req RejectionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Start database transaction
	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to start database transaction",
			"error":   tx.Error.Error(),
		})
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get apartment by ID
	var apartment model.Apartment
	if err := tx.Where("id = ?", apartmentID).First(&apartment).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "Apartment not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error retrieving apartment",
			"error":   err.Error(),
		})
	}

	// Check if already rejected
	if apartment.Status == "Rejected" {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "This apartment was already rejected",
		})
	}

	// Update apartment with rejection details
	updateData := map[string]interface{}{
		"message": req.RejectionReason,
		"status":  "Rejected",
		// Add any other fields you want to update, like rejected_at if your table has that column
		// "rejected_at": time.Now(),
	}

	if err := tx.Model(&model.Apartment{}).
		Where("id = ?", apartmentID).
		Updates(updateData).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update apartment",
			"error":   err.Error(),
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to commit transaction",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Apartment request rejected",
		"data": fiber.Map{
			"apartment_id":     apartmentID,
			"rejection_reason": req.RejectionReason,
			"updated_at":       time.Now().Format(time.RFC3339),
		},
	})
}
