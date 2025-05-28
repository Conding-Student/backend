package controller

import (
	"net/http"
	"strings"

	"github.com/Conding-Student/backend/config"
	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type LandlordRegistrationRequest struct {
	BusinessName    string   `json:"business_name" validate:"required"`
	BusinessAddress string   `json:"business_address" validate:"required"`
	BusinessContact string   `json:"business_contact" validate:"required"`
	IDImageURL      string   `json:"id_image_url" validate:"required,url"`
	PermitImageURLs []string `json:"permit_image_urls" validate:"required,min=1,dive,url"`
}

func RegisterLandlord(c *fiber.Ctx) error {
	// Get user from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid user UID",
		})
	}

	// Parse request body
	var req LandlordRegistrationRequest
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

	// Check if user exists
	var user model.User
	if err := tx.Where("uid = ?", uid).First(&user).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"message": "User not found",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error checking user",
			"error":   err.Error(),
		})
	}

	// New: Check if account status is already Pending
	if user.AccountStatus == "Pending" {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "You already have a pending registration request",
		})
	}

	// Check if already a landlord
	if user.UserType == "Landlord" {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "User is already registered as a landlord",
		})
	}

	// Upload ID image
	idImageURL, err := config.UploadImage(req.IDImageURL)
	if err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to upload ID image",
			"error":   err.Error(),
		})
	}

	// Upload business permits
	var permitURLs []string
	for _, permit := range req.PermitImageURLs {
		url, err := config.UploadImage(permit)
		if err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to upload business permit",
				"error":   err.Error(),
			})
		}
		permitURLs = append(permitURLs, url)
	}

	// Create landlord profile
	landlordProfile := model.LandlordProfile{
		Uid:            uid,
		VerificationID: idImageURL,
		BusinessName:   req.BusinessName,
		BusinessPermit: strings.Join(permitURLs, ","),
	}

	if err := tx.Create(&landlordProfile).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create landlord profile",
			"error":   err.Error(),
		})
	}

	// Update user status using raw SQL
	if err := tx.Exec("UPDATE users SET account_status = 'Pending' WHERE uid = ?", uid).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update user status",
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

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Landlord registration successful",
		"data": fiber.Map{
			"landlord_id":     landlordProfile.ID,
			"business_name":   landlordProfile.BusinessName,
			"verification_id": idImageURL,
			"permit_urls":     permitURLs,
		},
	})
}
