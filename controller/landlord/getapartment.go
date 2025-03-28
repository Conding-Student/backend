package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// ✅ Fetch Apartments Owned by the Landlord (With Owner Details, Amenities, House Rules & Images)
func FetchApartmentsByLandlord(c *fiber.Ctx) error {
	// Get landlord ID from JWT claims
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: missing JWT claims",
		})
	}

	// Extract user ID from JWT claims
	idFloat, ok := userClaims["id"].(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: invalid user ID in token",
		})
	}
	landlordID := uint(idFloat)

	// ✅ Fetch Apartments with Landlord Info (Explicit Join with Users Table)
	var apartments []struct {
		model.Apartment
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	if err := middleware.DBConn.
		Table("apartments").
		Select("apartments.*, users.first_name, users.last_name, users.email").
		Joins("JOIN users ON users.id = apartments.user_id").
		Where("apartments.user_id = ?", landlordID).
		Find(&apartments).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch apartments",
			"error":   err.Error(),
		})
	}

	// ✅ Loop through Apartments to fetch Images, Amenities & House Rules
	var response []fiber.Map
	for _, apartment := range apartments {
		// Fetch amenities
		var amenities []string
		middleware.DBConn.
			Table("amenities").
			Select("amenities.name").
			Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
			Where("apartment_amenities.apartment_id = ?", apartment.ID).
			Pluck("name", &amenities)

		// Fetch house rules
		var houseRules []string
		middleware.DBConn.
			Table("house_rules").
			Select("house_rules.rule").
			Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
			Where("apartment_house_rules.apartment_id = ?", apartment.ID).
			Pluck("rule", &houseRules)

		// Fetch apartment images (aggregated)
		var imageUrls []string
		middleware.DBConn.
			Table("apartment_images").
			Select("image_url").
			Where("apartment_id = ?", apartment.ID).
			Pluck("image_url", &imageUrls)

		// ✅ Append data to response
		response = append(response, fiber.Map{
			"apartment_id":   apartment.ID,
			"property_name":  apartment.PropertyName,
			"address":        apartment.Address,
			"property_type":  apartment.PropertyType,
			"rent_price":     apartment.RentPrice,
			"location_link":  apartment.LocationLink,
			"contact_number": apartment.ContactNumber,
			"email":          apartment.Email,
			"facebook":       apartment.Facebook,
			"status":         apartment.Status,
			"owner_first":    apartment.FirstName,
			"owner_last":     apartment.LastName,
			"owner_email":    apartment.Email,
			"amenities":      amenities,
			"house_rules":    houseRules,
			"image_urls":     imageUrls, // ✅ Now returns images as an array
		})
	}

	// ✅ Return success response
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message":    "Apartments fetched successfully",
		"apartments": response,
	})
}
