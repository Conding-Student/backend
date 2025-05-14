package controller

import (
	"fmt"
	"intern_template_v1/config"
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	//"intern_template_v1/model/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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

// Request struct for updating availability
type UpdateAvailabilityRequest struct {
	Availability string `json:"availability"`
}

func UpdateApartmentAvailability(c *fiber.Ctx) error {
	fmt.Println("[DEBUG] Starting UpdateApartmentAvailability handler...")

	// Authentication & Authorization
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		fmt.Println("[SECURITY] Missing JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid authentication credentials",
		})
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		fmt.Println("[SECURITY] Invalid UID in token claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid user identification",
		})
	}
	fmt.Println("[DEBUG] UID extracted:", uid)

	// Input Validation
	apartmentID := c.Params("id")
	if apartmentID == "" || !isValidID(apartmentID) {
		fmt.Println("[VALIDATION] Invalid numeric ID format")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid apartment ID format - must be numeric",
		})
	}

	var req UpdateAvailabilityRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Println("[VALIDATION] Request parsing error:", err.Error())
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	if !isValidAvailability(req.Availability) {
		fmt.Println("[VALIDATION] Invalid availability value:", req.Availability)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Availability must be either 'Available' or 'Not Available'",
		})
	}

	// Business Logic Validation
	var apartment model.Apartment
	if err := middleware.DBConn.
		Where("id = ? AND uid = ?", apartmentID, uid).
		First(&apartment).Error; err != nil {

		fmt.Println("[AUTH] Apartment access violation - ID:", apartmentID, "UID:", uid)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Apartment not found or access denied",
		})
	}

	if apartment.Status != "Approved" {
		fmt.Printf("[BUSINESS] Attempt to update unapproved apartment - ID: %s Current Status: %s\n",
			apartmentID, apartment.Status)
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"message": "Cannot update availability for unapproved listings",
			"detail":  "The property must be approved by administrators first",
		})
	}

	// Prepare Update Data
	updates := map[string]interface{}{
		"availability": req.Availability,
		"updated_at":   time.Now(), // Add update timestamp
	}

	// Expiration Logic
	switch req.Availability {
	case "Available":
		if apartment.ExpiresAt != nil && apartment.ExpiresAt.After(time.Now()) {
			fmt.Printf("[BUSINESS] Keeping existing expiration: %v\n", apartment.ExpiresAt)
			updates["expires_at"] = apartment.ExpiresAt
		} else {
			newExpiration := time.Now().Add(14 * 24 * time.Hour)
			updates["expires_at"] = newExpiration
			fmt.Printf("[DEBUG] New expiration set: %v\n", newExpiration)
		}
	case "Not Available":
		updates["expires_at"] = gorm.Expr("NULL")
		fmt.Println("[DEBUG] Clearing expiration time")
	}

	// Database Operation
	if err := middleware.DBConn.
		Model(&model.Apartment{}).
		Where("id = ? AND status = 'Approved'", apartmentID). // Additional safety check
		Updates(updates).Error; err != nil {

		fmt.Println("[DATABASE] Update error:", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update property status",
			"error":   "Database operation failed",
		})
	}

	fmt.Println("[SUCCESS] Availability updated for apartment ID:", apartmentID)
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":    "Property availability updated",
		"expires_at": updates["expires_at"],
	})
}

// Helper functions
func isValidID(id string) bool {
	_, err := strconv.Atoi(id)
	return err == nil
}

func isValidAvailability(a string) bool {
	return a == "Available" || a == "Not Available"
}

func ManageApartmentExpirations() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute instead of hourly
	defer ticker.Stop()

	for {
		<-ticker.C
		currentTime := time.Now()

		// Directly update using a single query
		result := middleware.DBConn.Model(&model.Apartment{}).
			Where("expires_at < ? AND availability = ?", currentTime, "Available").
			Updates(map[string]interface{}{
				"availability": "Not Available",
				"expires_at":   gorm.Expr("NULL"),
			})

		if result.Error != nil {
			fmt.Printf("Error updating expired apartments: %v\n", result.Error)
		} else if result.RowsAffected > 0 {
			fmt.Printf("[%s] Expired %d apartments\n",
				currentTime.Format(time.RFC3339), result.RowsAffected)
		}
	}
}

// ✅ Background cleaner with enhanced logging
func ManageExpiredDeletions() {
	fmt.Println("[BACKGROUND CLEANER] Starting automatic deletion scheduler...")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		startTime := time.Now().UTC()
		fmt.Printf("\n[%s] Starting deletion cycle\n", startTime.Format(time.RFC3339))

		currentTime := time.Now().UTC()

		// Track deletions
		userCount := deleteExpiredRecords(&model.User{}, "account_status = ?", "Deleted", currentTime)
		apartmentCount := deleteExpiredRecords(&model.Apartment{}, "status = ?", "Deleted", currentTime)
		inquiryCount := deleteExpiredRecords(&model.Inquiry{}, "status = ?", "Rejected", currentTime)

		total := userCount + apartmentCount + inquiryCount

		if total > 0 {
			fmt.Printf("[%s] Deletion complete: %d total records purged (Users: %d, Apartments: %d, Inquiries: %d)\n",
				currentTime.Format(time.RFC3339),
				total,
				userCount,
				apartmentCount,
				inquiryCount)
		} else {
			fmt.Printf("[%s] No expired records found for deletion\n", currentTime.Format(time.RFC3339))
		}

		fmt.Printf("[%s] Cycle duration: %v\n\n",
			currentTime.Format(time.RFC3339),
			time.Since(startTime).Round(time.Millisecond))
	}
}

// ✅ Enhanced helper function with return count
func deleteExpiredRecords(model interface{}, statusQuery string, statusValue string, currentTime time.Time) int64 {
	result := middleware.DBConn.Unscoped().Where(
		"expires_at < ? AND "+statusQuery,
		currentTime,
		statusValue,
	).Delete(model)

	if result.Error != nil {
		fmt.Printf("Error deleting records: %v\n", result.Error)
		return 0
	}

	if result.RowsAffected > 0 {
		fmt.Printf("[%s] Permanently deleted %d %T records\n",
			currentTime.Format(time.RFC3339),
			result.RowsAffected,
			model)
	}

	return result.RowsAffected
}
