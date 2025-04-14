package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	//"github.com/golang-jwt/jwt/v5"
)

// Function to fetch approved apartments for tenants
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
		Videos           []string `json:"videos"` // <-- Added field for videos
		Amenities        []string `json:"amenities"`
		HouseRules       []string `json:"house_rules"`
		InquiriesCount   int64    `json:"inquiries_count"`
	}

	var apartments []model.Apartment
	if err := middleware.DBConn.Where("status = ?", "Approved").Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch approved apartments",
			"error":   err.Error(),
		})
	}

	var results []ApartmentDetails

	for _, apt := range apartments {
		var landlord model.User
		if err := middleware.DBConn.Where("uid = ?", apt.Uid).First(&landlord).Error; err != nil {
			// Skip if no landlord is found
			continue
		}

		var images []model.ApartmentImage
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
		var imageUrls []string
		for _, img := range images {
			imageUrls = append(imageUrls, img.ImageURL)
		}

			// Fetch Videos
			var videos []model.ApartmentVideo
			middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&videos)
			var videoUrls []string
			for _, vid := range videos {
				videoUrls = append(videoUrls, vid.VideoURL)
			}

		var amenities []model.Amenity
		middleware.DBConn.
			Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
			Where("apartment_amenities.apartment_id = ?", apt.ID).Find(&amenities)
		var amenityNames []string
		for _, a := range amenities {
			amenityNames = append(amenityNames, a.Name)
		}

		var houseRules []model.HouseRule
		middleware.DBConn.
			Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
			Where("apartment_house_rules.apartment_id = ?", apt.ID).Find(&houseRules)
		var ruleNames []string
		for _, r := range houseRules {
			ruleNames = append(ruleNames, r.Rule)
		}

		var inquiryCount int64
		middleware.DBConn.Model(&model.Inquiry{}).Where("apartment_id = ?", apt.ID).Count(&inquiryCount)

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
			Videos:           videoUrls, // <-- Added to response
			Amenities:        amenityNames,
			HouseRules:       ruleNames,
			InquiriesCount:   inquiryCount,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"apartments": results,
	})
}
