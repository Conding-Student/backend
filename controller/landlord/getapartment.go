package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	//"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Function to fetch apartment details along with its amenities and house rules based on ApartmentID
// Function to get apartment details based on UID
// Function to get apartment details based on UID
func FetchApartmentsByLandlord(c *fiber.Ctx) error {
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

	// Fetch apartments associated with the UID
	var apartments []model.Apartment
	if err := middleware.DBConn.Where("uid = ?", uid).Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch apartments",
			"error":   err.Error(),
		})
	}

	// If no apartments found
	if len(apartments) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No apartments found for this landlord",
		})
	}

	// Prepare the result with apartment details, amenities, house rules, images, and inquiries
	var apartmentDetails []fiber.Map
	for _, apartment := range apartments {
		// Fetch amenities associated with the apartment
		var amenities []model.Amenity
		if err := middleware.DBConn.
			Joins("JOIN apartment_amenities aa ON aa.amenity_id = amenities.id").
			Where("aa.apartment_id = ?", apartment.ID).
			Find(&amenities).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to fetch amenities",
				"error":   err.Error(),
			})
		}

		// Fetch house rules associated with the apartment
		var houseRules []model.HouseRule
		if err := middleware.DBConn.
			Joins("JOIN apartment_house_rules ahr ON ahr.house_rule_id = house_rules.id").
			Where("ahr.apartment_id = ?", apartment.ID).
			Find(&houseRules).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to fetch house rules",
				"error":   err.Error(),
			})
		}

		// Fetch images for the apartment
		var images []model.ApartmentImage
		if err := middleware.DBConn.Where("apartment_id = ?", apartment.ID).Find(&images).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to fetch images",
				"error":   err.Error(),
			})
		}

		// Fetch the number of inquiries for the apartment
		var inquiryCount int64
		if err := middleware.DBConn.Model(&model.Inquiry{}).Where("apartment_id = ?", apartment.ID).Count(&inquiryCount).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Database error: Unable to count inquiries",
				"error":   err.Error(),
			})
		}

		// Prepare the apartment details with amenities, house rules, images, and inquiries
		apartmentDetails = append(apartmentDetails, fiber.Map{
			"property_name": apartment.PropertyName,
			"property_type": apartment.PropertyType,
			"rent_price":    apartment.RentPrice,
			"location_link": apartment.LocationLink,
			"landmarks":     apartment.Landmarks,
			"images": func() []fiber.Map {
				var imageDetails []fiber.Map
				for _, image := range images {
					imageDetails = append(imageDetails, fiber.Map{
						"image_url": image.ImageURL, // Exclude apartment_id
					})
				}
				return imageDetails
			}(),
			"amenities": func() []string {
				var amenityNames []string
				for _, amenity := range amenities {
					amenityNames = append(amenityNames, amenity.Name)
				}
				return amenityNames
			}(),
			"house_rules": func() []string {
				var ruleNames []string
				for _, rule := range houseRules {
					ruleNames = append(ruleNames, rule.Rule)
				}
				return ruleNames
			}(),
			"inquiries_count": inquiryCount,
		})
	}

	// 🎉 Success Response with apartment details
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"apartments": apartmentDetails,
	})
}
