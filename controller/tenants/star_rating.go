package controller

import (
	"errors"

	"strconv"
	"time"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// ConfirmRental allows a tenant or landlord to confirm a rental agreement
func ConfirmRental(c *fiber.Ctx) error {
	type request struct {
		ApartmentID uint   `json:"apartment_id"`
		IsRenting   bool   `json:"is_renting"`
		TenantID    string `json:"tenant_id,omitempty"` // For landlord confirmations
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	// Extract user claims
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	uid, uidOk := userClaims["uid"].(string)
	userType, userTypeOk := userClaims["role"].(string)
	if !uidOk || !userTypeOk {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	// Find or create rental agreement
	var agreement model.RentalAgreement
	var apartment model.Apartment

	// Get the apartment to verify it exists
	if err := middleware.DBConn.First(&apartment, req.ApartmentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Apartment not found",
		})
	}

	// Tenant confirmation flow
	if userType == "Tenant" {
		// Each tenant has their own agreement record
		err := middleware.DBConn.
			Where("apartment_id = ? AND tenant_id = ?", req.ApartmentID, uid).
			First(&agreement).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new agreement for this tenant
			agreement = model.RentalAgreement{
				ApartmentID:       req.ApartmentID,
				TenantID:          uid,
				LandlordID:        apartment.UserID,
				TenantConfirmed:   req.IsRenting,
				LandlordConfirmed: false, // Landlord needs to confirm separately
				StartDate:         time.Now(),
				IsActive:          true,
			}
			if err := middleware.DBConn.Create(&agreement).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create rental agreement",
				})
			}
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check rental agreement",
			})
		} else {
			// Update existing tenant agreement
			agreement.TenantConfirmed = req.IsRenting
			if err := middleware.DBConn.Save(&agreement).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update rental agreement",
				})
			}
		}
	} else if userType == "Landlord" {
		// Landlord confirms a specific tenant's agreement
		if req.TenantID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant ID required for landlord confirmation",
			})
		}

		// Verify the landlord owns this apartment
		if apartment.UserID != uid {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not the landlord of this apartment",
			})
		}

		err := middleware.DBConn.
			Where("apartment_id = ? AND tenant_id = ?", req.ApartmentID, req.TenantID).
			First(&agreement).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new agreement with landlord confirmation
			agreement = model.RentalAgreement{
				ApartmentID:       req.ApartmentID,
				TenantID:          req.TenantID,
				LandlordID:        uid,
				TenantConfirmed:   false, // Tenant needs to confirm separately
				LandlordConfirmed: req.IsRenting,
				StartDate:         time.Now(),
				IsActive:          true,
			}
			if err := middleware.DBConn.Create(&agreement).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create rental agreement",
				})
			}
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check rental agreement",
			})
		} else {
			// Update existing agreement with landlord confirmation
			agreement.LandlordConfirmed = req.IsRenting
			if err := middleware.DBConn.Save(&agreement).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update rental agreement",
				})
			}
		}
	} else {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only tenants or landlords can confirm rentals",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Rental confirmation updated successfully",
		"data": fiber.Map{
			"tenant_confirmed":   agreement.TenantConfirmed,
			"landlord_confirmed": agreement.LandlordConfirmed,
		},
	})
}

// SubmitRating lets a tenant submit or update their rating for an apartment
func SubmitRating(c *fiber.Ctx) error {
	type request struct {
		ApartmentID uint   `json:"apartment_id"`
		Rating      int    `json:"rating"`
		Comment     string `json:"comment"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	uid := userClaims["uid"].(string)
	userType := userClaims["role"].(string)

	if userType != "Tenant" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Only tenants can rate apartments"})
	}

	var agreement model.RentalAgreement
	if err := middleware.DBConn.
		Where("apartment_id = ? AND tenant_id = ? AND tenant_confirmed = ? AND landlord_confirmed = ?",
			req.ApartmentID, uid, true, true).
		First(&agreement).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "You must be a confirmed tenant to rate this apartment"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to verify rental status"})
	}

	rating := model.Rating{
		ApartmentID: req.ApartmentID,
		TenantID:    uid,
		Rating:      req.Rating,
		Comment:     req.Comment,
	}

	result := middleware.DBConn.Where("apartment_id = ? AND tenant_id = ?", req.ApartmentID, uid).
		Assign(rating).
		FirstOrCreate(&rating)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to submit rating"})
	}

	return c.JSON(fiber.Map{
		"message": "Rating submitted successfully",
		"data":    rating,
	})
}

type RatingResponse struct {
	ID        uint      `json:"id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	Tenant    struct {
		Fullname string `json:"fullname"`
		PhotoURL string `json:"photo_url"`
	} `json:"tenant"`
}

func GetApartmentRatings(c *fiber.Ctx) error {
	apartmentID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid apartment ID"})
	}

	var ratings []model.Rating
	if err := middleware.DBConn.
		Preload("Tenant").
		Where("apartment_id = ?", apartmentID).
		Find(&ratings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch ratings"})
	}

	var total int
	var response []RatingResponse

	for _, r := range ratings {
		total += r.Rating

		response = append(response, RatingResponse{
			ID:        r.ID,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt,
			Tenant: struct {
				Fullname string `json:"fullname"`
				PhotoURL string `json:"photo_url"`
			}{
				Fullname: r.Tenant.Fullname,
				PhotoURL: r.Tenant.PhotoURL,
			},
		})
	}

	average := 0.0
	if len(ratings) > 0 {
		average = float64(total) / float64(len(ratings))
	}

	return c.JSON(fiber.Map{
		"average_rating": average,
		"ratings":        response,
	})
}

// GetTenantIDsByApartment returns all tenant UIDs for a given apartment's rental agreements
// GetTenantIDByRentalAgreementID returns the tenant UID for a given rental agreement ID
func GetTenantIDByRentalAgreementID(c *fiber.Ctx) error {
	agreementID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid rental agreement ID"})
	}

	var agreement model.RentalAgreement
	if err := middleware.DBConn.First(&agreement, agreementID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Rental agreement not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve agreement"})
	}

	return c.JSON(fiber.Map{
		"rental_agreement_id": agreement.ID,
		"tenant_id":           agreement.TenantID,
	})
}

func CheckRatingEligibility(c *fiber.Ctx) error {
	apartmentIDStr := c.Query("apartment_id")
	tenantID := c.Query("tenant_id")

	if apartmentIDStr == "" || tenantID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing apartment_id or tenant_id",
		})
	}

	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid apartment ID",
		})
	}

	var agreement model.RentalAgreement
	err = middleware.DBConn.
		Where("apartment_id = ? AND tenant_id = ? AND tenant_confirmed = ? AND landlord_confirmed = ?",
			apartmentID, tenantID, true, true).
		First(&agreement).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(fiber.Map{"can_rate": false})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error while checking eligibility",
		})
	}

	return c.JSON(fiber.Map{"can_rate": true})
}
