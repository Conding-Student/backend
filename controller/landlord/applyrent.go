package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for parsing apartment creation request
type ApartmentRequest struct {
	PropertyName  string   `json:"property_name"`
	Address       string   `json:"address"`
	PropertyType  string   `json:"property_type"`
	RentPrice     float64  `json:"rent_price"`
	LocationLink  string   `json:"location_link"`
	Landmarks     string   `json:"landmarks"`
	ContactNumber string   `json:"contact_number"`
	Email         string   `json:"email"`
	Facebook      string   `json:"facebook"`
	Amenities     []string `json:"amenities"`
	HouseRules    []string `json:"house_rules"`
	ImageURLs     []string `json:"image_urls"`
}

// ✅ Function for landlords to add their apartment
func CreateApartment(c *fiber.Ctx) error {
	// Get user claims from JWT stored in middleware
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: missing JWT claims",
		})
	}

	// Extract user ID safely from JWT claims
	idFloat, ok := userClaims["id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: invalid user ID in token",
		})
	}
	userIDExtracted := uint(idFloat)

	// 📌 Parse request body
	var req ApartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"ret_code": "400",
			"message":  "Invalid request format",
			"error":    err.Error(),
		})
	}

	// 📌 Validate required fields
	if req.PropertyName == "" || req.Address == "" || req.PropertyType == "" || req.RentPrice <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"ret_code": "400",
			"message":  "Missing or invalid required fields",
		})
	}

	// 📌 Create apartment entry
	apartment := model.Apartment{
		UserID:        userIDExtracted, // ✅ Using UserID since there's no LandlordID
		PropertyName:  req.PropertyName,
		Address:       req.Address,
		PropertyType:  req.PropertyType,
		RentPrice:     req.RentPrice,
		LocationLink:  req.LocationLink,
		Landmarks:     req.Landmarks,
		ContactNumber: req.ContactNumber,
		Email:         req.Email,
		Facebook:      req.Facebook,
		Status:        "Pending", // Default status
	}

	// 📌 Save apartment in DB
	if err := middleware.DBConn.Create(&apartment).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"ret_code": "500",
			"message":  "Database error: Unable to create apartment",
			"error":    err.Error(),
		})
	}

	// 🔄 Process Amenities
	for _, amenityName := range req.Amenities {
		var amenity model.Amenity
		if err := middleware.DBConn.Where("name = ?", amenityName).FirstOrCreate(&amenity, model.Amenity{Name: amenityName}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ret_code": "500",
				"message":  "Database error: Unable to save amenity",
				"error":    err.Error(),
			})
		}
		// ✅ Link amenity to the apartment
		if err := middleware.DBConn.Create(&model.ApartmentAmenity{ApartmentID: apartment.ID, AmenityID: amenity.ID}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ret_code": "500",
				"message":  "Database error: Unable to link amenity",
				"error":    err.Error(),
			})
		}
	}

	// 🔄 Process House Rules
	for _, ruleName := range req.HouseRules {
		var rule model.HouseRule
		if err := middleware.DBConn.Where("rule = ?", ruleName).FirstOrCreate(&rule, model.HouseRule{Rule: ruleName}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ret_code": "500",
				"message":  "Database error: Unable to save house rule",
				"error":    err.Error(),
			})
		}
		// ✅ Link house rule to the apartment
		if err := middleware.DBConn.Create(&model.ApartmentHouseRule{ApartmentID: apartment.ID, HouseRuleID: rule.ID}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ret_code": "500",
				"message":  "Database error: Unable to link house rule",
				"error":    err.Error(),
			})
		}
	}

	// 🔄 Process Images
	for _, imageURL := range req.ImageURLs {
		apartmentImage := model.ApartmentImage{
			ApartmentID: apartment.ID,
			ImageURL:    imageURL,
		}

		if err := middleware.DBConn.Create(&apartmentImage).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"ret_code": "500",
				"message":  "Database error: Unable to save apartment image",
				"error":    err.Error(),
			})
		}
	}

	// ✅ Return success response
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"ret_code": "201",
		"message":  "Apartment created successfully",
		"data": fiber.Map{
			"apartment_id":  apartment.ID,
			"property_name": apartment.PropertyName,
			"status":        apartment.Status,
		},
	})
}
