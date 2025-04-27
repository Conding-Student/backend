package controller

import (
	"intern_template_v1/config"
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type UpdateMediaRequest struct {
	ImageURLs []string `json:"image_urls"`
	VideoURLs []string `json:"video_urls"`
}

func UpdateApartmentMedia(c *fiber.Ctx) error {
	// Get user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// Get landlord UID
	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// Get apartment ID from URL parameters
	apartmentID := c.Params("id")
	if apartmentID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Apartment ID is required",
		})
	}

	// Parse request body
	var req UpdateMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Verify apartment exists and belongs to the current landlord
	var apartment model.Apartment
	if err := middleware.DBConn.Where("id = ? AND uid = ?", apartmentID, uid).First(&apartment).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Apartment not found or unauthorized",
		})
	}

	// Start transaction
	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to start transaction",
		})
	}

	// Upload new images
	var imageURLs []string
	for _, img := range req.ImageURLs {
		url, err := config.UploadImage(img)
		if err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to upload image to Cloudinary",
				"error":   err.Error(),
			})
		}
		imageURLs = append(imageURLs, url)
		if err := tx.Create(&model.ApartmentImage{ApartmentID: apartment.ID, ImageURL: url}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to save image URL",
				"error":   err.Error(),
			})
		}
	}

	// Upload new videos
	var videoURLs []string
	for _, vid := range req.VideoURLs {
		url, err := config.UploadVideo(vid)
		if err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to upload video to Cloudinary",
				"error":   err.Error(),
			})
		}
		videoURLs = append(videoURLs, url)
		if err := tx.Create(&model.ApartmentVideo{ApartmentID: apartment.ID, VideoURL: url}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Failed to save video URL",
				"error":   err.Error(),
			})
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Transaction commit failed",
			"error":   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Apartment media updated successfully",
		"data": fiber.Map{
			"image_urls": imageURLs,
			"video_urls": videoURLs,
		},
	})
}
