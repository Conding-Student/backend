package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func FetchApprovedApartmentsForTenant(c *fiber.Ctx) error {
	type ApartmentDetails struct {
		model.Apartment
		LandlordName     string   `json:"landlord_name"`
		LandlordEmail    string   `json:"landlord_email"`
		LandlordPhone    string   `json:"landlord_phone"`
		LandlordAddress  string   `json:"landlord_address"`
		LandlordValidID  string   `json:"landlord_valid_id"`
		LandlordPhotoURL string   `json:"landlord_photo_url"`
		LandlordUserType string   `json:"landlord_user_type"`
		LandlordStatus   string   `json:"landlord_account_status"`
		Images           []string `json:"images"`
		Videos           []string `json:"videos"`
		Amenities        []string `json:"amenities"`
		HouseRules       []string `json:"house_rules"`
		InquiriesCount   int64    `json:"inquiries_count"`
	}

	// Get query parameters
	propertyTypes := c.Query("property_types")
	minPriceStr := c.Query("min_price")
	maxPriceStr := c.Query("max_price")
	amenitiesFilter := strings.Split(c.Query("amenities"), ",")
	houseRulesFilter := strings.Split(c.Query("house_rules"), ",")

	var apartments []model.Apartment
	db := middleware.DBConn.Where("status = ?", "Approved")

	// Filter by property types (if any)
	if propertyTypes != "" {
		propertyTypeList := strings.Split(propertyTypes, ",")
		for i := range propertyTypeList {
			propertyTypeList[i] = strings.TrimSpace(propertyTypeList[i])
		}
		db = db.Where("property_type IN ?", propertyTypeList)
	}

	// Filter by price range
	if minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			db = db.Where("rent_price >= ?", minPrice)
		}
	}
	if maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			db = db.Where("rent_price <= ?", maxPrice)
		}
	}

	if err := db.Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch apartments",
			"error":   err.Error(),
		})
	}

	var results []ApartmentDetails

	// Helper to check if apartment has all items in filter
	matchesAll := func(apartmentItems []string, filterItems []string) bool {
		for _, f := range filterItems {
			f = strings.TrimSpace(f)
			found := false
			for _, item := range apartmentItems {
				if strings.EqualFold(strings.TrimSpace(item), f) {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	}

	for _, apt := range apartments {
		var landlord model.User
		if err := middleware.DBConn.Where("uid = ?", apt.Uid).First(&landlord).Error; err != nil {
			continue
		}

		// Images
		var images []model.ApartmentImage
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
		var imageUrls []string
		for _, img := range images {
			imageUrls = append(imageUrls, img.ImageURL)
		}

		// Videos
		var videos []model.ApartmentVideo
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&videos)
		var videoUrls []string
		for _, vid := range videos {
			videoUrls = append(videoUrls, vid.VideoURL)
		}

		// Amenities
		var amenities []model.Amenity
		middleware.DBConn.
			Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
			Where("apartment_amenities.apartment_id = ?", apt.ID).
			Find(&amenities)
		var amenityNames []string
		for _, a := range amenities {
			amenityNames = append(amenityNames, a.Name)
		}

		// House Rules
		var houseRules []model.HouseRule
		middleware.DBConn.
			Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
			Where("apartment_house_rules.apartment_id = ?", apt.ID).
			Find(&houseRules)
		var ruleNames []string
		for _, r := range houseRules {
			ruleNames = append(ruleNames, r.Rule)
		}

		// Inquiries Count
		var inquiryCount int64
		middleware.DBConn.Model(&model.Inquiry{}).
			Where("apartment_id = ?", apt.ID).Count(&inquiryCount)

		// Apply amenity and rule filters
		if c.Query("amenities") != "" && !matchesAll(amenityNames, amenitiesFilter) {
			continue
		}
		if c.Query("house_rules") != "" && !matchesAll(ruleNames, houseRulesFilter) {
			continue
		}

		// Append result
		results = append(results, ApartmentDetails{
			Apartment:        apt,
			LandlordName:     landlord.Fullname,
			LandlordEmail:    landlord.Email,
			LandlordPhone:    landlord.PhoneNumber,
			LandlordAddress:  landlord.Address,
			LandlordValidID:  landlord.ValidID,
			LandlordPhotoURL: landlord.PhotoURL,
			LandlordUserType: landlord.UserType,
			LandlordStatus:   landlord.AccountStatus,
			Images:           imageUrls,
			Videos:           videoUrls,
			Amenities:        amenityNames,
			HouseRules:       ruleNames,
			InquiriesCount:   inquiryCount,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"apartments": results,
	})
}

//
