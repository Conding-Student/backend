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
	PropertyName string   `json:"property_name"`
	PropertyType string   `json:"property_type"`
	RentPrice    float64  `json:"rent_price"`
	LocationLink string   `json:"location_link"`
	Landmarks    string   `json:"landmarks"`
	Amenities    []string `json:"amenities"`
	HouseRules   []string `json:"house_rules"`
	ImageURLs    []string `json:"image_urls"`
}

// ✅ Function to create an apartment
func CreateApartment(c *fiber.Ctx) error {
	// 🔍 Extract user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	// 🆔 Extract Landlord Uid safely from JWT claims
	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// 🔍 Verify if the user is registered as a landlord
	var user model.User
	if err := middleware.DBConn.Where("uid = ? AND user_type = ?", uid, "Landlord").First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: User is not a registered landlord",
		})
	}

	// 📩 Parse request body
	var req ApartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// 📌 Validate required fields
	if req.PropertyName == "" || req.PropertyType == "" || req.RentPrice <= 0 || req.LocationLink == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required fields: property_name, property_type, rent_price, or location_link",
		})
	}

	// 🏡 Default Address Value
	address := "address"

	// Start transaction
	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to start transaction",
		})
	}

	// 🔍 Check if the apartment with the same PropertyName and LocationLink already exists for the same UID
	var existingApartment model.Apartment
	if err := tx.Where("property_name = ? AND location_link = ? AND uid = ?", req.PropertyName, req.LocationLink, uid).First(&existingApartment).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "Apartment with the same property name and location already exists for this landlord",
		})
	}

	// Create the apartment
	apartment := model.Apartment{
		Uid:          uid,
		PropertyName: req.PropertyName,
		Address:      address, // Default address
		PropertyType: req.PropertyType,
		RentPrice:    req.RentPrice,
		LocationLink: req.LocationLink,
		Landmarks:    req.Landmarks,
		Status:       "Pending", // Default status
		UserID:       uid,       // Assuming 'user_id' should be the same as 'uid'
	}

	// Insert apartment into the apartments table
	if err := tx.Create(&apartment).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to create apartment",
			"error":   err.Error(),
		})
	}

	// 🔹 Insert amenities (Avoid duplicates)
	for _, amenityName := range req.Amenities {
		var amenity model.Amenity
		if err := tx.Where("name = ?", amenityName).FirstOrCreate(&amenity, model.Amenity{Name: amenityName}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to add amenities",
				"error":   err.Error(),
			})
		}

		apartmentAmenity := model.ApartmentAmenity{
			ApartmentID: apartment.ID,
			AmenityID:   amenity.ID,
		}
		if err := tx.Create(&apartmentAmenity).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to link amenities",
				"error":   err.Error(),
			})
		}
	}

	// 🔹 Insert house rules (Avoid duplicates)
	for _, ruleName := range req.HouseRules {
		var houseRule model.HouseRule
		if err := tx.Where("rule = ?", ruleName).FirstOrCreate(&houseRule, model.HouseRule{Rule: ruleName}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to add house rules",
				"error":   err.Error(),
			})
		}

		apartmentHouseRule := model.ApartmentHouseRule{
			ApartmentID: apartment.ID,
			HouseRuleID: houseRule.ID,
		}
		if err := tx.Create(&apartmentHouseRule).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to link house rules",
				"error":   err.Error(),
			})
		}
	}

	// 📷 Insert images
	for _, imageURL := range req.ImageURLs {
		apartmentImage := model.ApartmentImage{
			ApartmentID: apartment.ID,
			ImageURL:    imageURL,
		}
		if err := tx.Create(&apartmentImage).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to insert apartment images",
				"error":   err.Error(),
			})
		}
	}

	// ✅ Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Transaction commit failed",
			"error":   err.Error(),
		})
	}

	// 🎉 Success Response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Apartment created successfully",
		"data": fiber.Map{
			"apartment_id":  apartment.ID,
			"uid":           apartment.Uid,
			"property_name": apartment.PropertyName,
			"address":       apartment.Address,
			"property_type": apartment.PropertyType,
			"rent_price":    apartment.RentPrice,
			"location_link": apartment.LocationLink,
			"landmarks":     apartment.Landmarks,
			"status":        apartment.Status,
		},
	})
}
