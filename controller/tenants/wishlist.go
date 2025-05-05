package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)


func AddToWishlist(c *fiber.Ctx) error {
	uid, err := GetUIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	type Request struct {
		ApartmentID uint `json:"apartment_id"`
	}
	var body Request
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check if already exists
	var existing model.Wishlist
	if err := middleware.DBConn.Where("uid = ? AND apartment_id = ?", uid, body.ApartmentID).First(&existing).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Apartment already in wishlist"})
	}

	// Add to wishlist
	wishlist := model.Wishlist{
		UID:         uid,
		ApartmentID: body.ApartmentID,
		CreatedAt:   time.Now(),
	}
	if err := middleware.DBConn.Create(&wishlist).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database error"})
	}

	return c.JSON(fiber.Map{"message": "Apartment added to wishlist"})
}

func RemoveFromWishlist(c *fiber.Ctx) error {
	log.Println("RemoveFromWishlist handler triggered") // Check if this is logged
	uid, err := GetUIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	apartmentID := c.Params("apartment_id")

	// Delete where UID and ApartmentID match
	if err := middleware.DBConn.Where("uid = ? AND apartment_id = ?", uid, apartmentID).Delete(&model.Wishlist{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove from wishlist"})
	}

	return c.JSON(fiber.Map{"message": "Apartment removed from wishlist"})
}

// Function to fetch approved apartments for tenants based on their wishlist
func FetchwishlistForTenant(c *fiber.Ctx) error {
	// Extract tenant UID from the JWT token
	uid, err := GetUIDFromToken(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing or invalid JWT",
		})
	}

	// Step 1: Retrieve all wishlist items for this tenant
	var wishlistItems []model.Wishlist
	if err := middleware.DBConn.
		Preload("Apartment").  // Preload the Apartment details
		Where("uid = ?", uid). // Filter wishlist by tenant UID
		Find(&wishlistItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to retrieve wishlist",
			"error":   err.Error(),
		})
	}

	// If no wishlist items found, return an empty array
	if len(wishlistItems) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No wishlist items found",
		})
	}

	// Step 2: Fetch approved apartments based on wishlist ApartmentIDs
	var apartmentIDs []uint
	for _, item := range wishlistItems {
		apartmentIDs = append(apartmentIDs, item.ApartmentID)
	}

	var apartments []model.Apartment
	if err := middleware.DBConn.
		Where("status = ? AND id IN (?)", "Approved", apartmentIDs). // Filter by approved status and matching apartment IDs
		Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch approved apartments",
			"error":   err.Error(),
		})
	}

	// If no approved apartments found
	if len(apartments) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No approved apartments available in your wishlist",
		})
	}

	// Step 3: Fetch the associated details like amenities, house rules, images, and inquiries
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
			"apartment_id":  apartment.ID,
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

	// 🎉 Success Response with wishlist apartments and their details
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"wishlist":   wishlistItems,    // The original wishlist data (apartment IDs, etc.)
		"apartments": apartmentDetails, // The approved apartment details with all info
	})
}
