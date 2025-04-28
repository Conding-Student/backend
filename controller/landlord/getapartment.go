package controller

import (
	"intern_template_v1/config"
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	//"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func FetchApartmentsByLandlord(c *fiber.Ctx) error {
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

	// Extract user claims from JWT
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	var apartments []model.Apartment
	if err := middleware.DBConn.Where("uid = ?", uid).Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to fetch apartments",
			"error":   err.Error(),
		})
	}

	if len(apartments) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No apartments found for this landlord",
		})
	}

	// Fetch landlord details once (all apartments share same landlord UID)
	var landlord model.User
	if err := middleware.DBConn.Where("uid = ?", uid).First(&landlord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Unable to fetch landlord details",
			"error":   err.Error(),
		})
	}

	var results []ApartmentDetails

	for _, apt := range apartments {
		var images []model.ApartmentImage
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
		var imageUrls []string
		for _, img := range images {
			imageUrls = append(imageUrls, img.ImageURL)
		}

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

// update apartment details

func UpdateApartment(c *fiber.Ctx) error {
    type UpdateInput struct {
        // Media fields
        ImageURLs []string `json:"image_urls"`
        VideoURLs []string `json:"video_urls"`
        
        // Property details
        PropertyName string    `json:"property_name"`
        Address      string    `json:"address"`
        PropertyType string    `json:"property_type"`
        RentPrice    float64   `json:"rent_price"`
        LocationLink string    `json:"location_link"`
        Landmarks    string    `json:"landmarks"`
        Latitude     *float64  `json:"latitude"`
        Longitude    *float64  `json:"longitude"`
        
        // Associated data
        Amenities  *[]string `json:"amenities"`
        HouseRules *[]string `json:"house_rules"`
    }

    // Authentication and validation
    userClaims, ok := c.Locals("user").(jwt.MapClaims)
    if !ok {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
    }

    uid, ok := userClaims["uid"].(string)
    if !ok || uid == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid UID"})
    }

    apartmentID := c.Params("id")
    var apartment model.Apartment
    if err := middleware.DBConn.First(&apartment, "id = ? AND uid = ?", apartmentID, uid).Error; err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Apartment not found or unauthorized"})
    }

    var input UpdateInput
    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid input", "error": err.Error()})
    }

    // Start transaction
    tx := middleware.DBConn.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 1. Update property details
    if input.PropertyName != "" {
        apartment.PropertyName = input.PropertyName
    }
    if input.Address != "" {
        apartment.Address = input.Address
    }
    if input.PropertyType != "" {
        apartment.PropertyType = input.PropertyType
    }
    if input.RentPrice != 0 {
        apartment.RentPrice = input.RentPrice
    }
    if input.LocationLink != "" {
        apartment.LocationLink = input.LocationLink
    }
    if input.Landmarks != "" {
        apartment.Landmarks = input.Landmarks
    }
    if input.Latitude != nil {
        apartment.Latitude = *input.Latitude
    }
    if input.Longitude != nil {
        apartment.Longitude = *input.Longitude
    }

    if err := tx.Save(&apartment).Error; err != nil {
        tx.Rollback()
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message": "Failed to update apartment details",
            "error": err.Error(),
        })
    }

    // 2. Handle media updates
    var imageURLs []string
    if len(input.ImageURLs) > 0 {
        // First delete existing images if we're replacing them
        if err := tx.Where("apartment_id = ?", apartment.ID).Delete(&model.ApartmentImage{}).Error; err != nil {
            tx.Rollback()
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to clear existing images",
                "error": err.Error(),
            })
        }

        // Upload and save new images
        for _, img := range input.ImageURLs {
            url, err := config.UploadImage(img)
            if err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to upload image",
                    "error": err.Error(),
                })
            }
            imageURLs = append(imageURLs, url)
            if err := tx.Create(&model.ApartmentImage{
                ApartmentID: apartment.ID,
                ImageURL:    url,
            }).Error; err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to save image URL",
                    "error": err.Error(),
                })
            }
        }
    }

    var videoURLs []string
    if len(input.VideoURLs) > 0 {
        // First delete existing videos if we're replacing them
        if err := tx.Where("apartment_id = ?", apartment.ID).Delete(&model.ApartmentVideo{}).Error; err != nil {
            tx.Rollback()
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to clear existing videos",
                "error": err.Error(),
            })
        }

        // Upload and save new videos
        for _, vid := range input.VideoURLs {
            url, err := config.UploadVideo(vid)
            if err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to upload video",
                    "error": err.Error(),
                })
            }
            videoURLs = append(videoURLs, url)
            if err := tx.Create(&model.ApartmentVideo{
                ApartmentID: apartment.ID,
                VideoURL:    url,
            }).Error; err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to save video URL",
                    "error": err.Error(),
                })
            }
        }
    }

    // 3. Update amenities if provided
    if input.Amenities != nil {
        // Clear existing amenities
        if err := tx.Where("apartment_id = ?", apartment.ID).Delete(&model.ApartmentAmenity{}).Error; err != nil {
            tx.Rollback()
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to clear existing amenities",
                "error": err.Error(),
            })
        }

        // Add new amenities
        for _, amenityName := range *input.Amenities {
            var amenity model.Amenity
            if err := tx.FirstOrCreate(&amenity, model.Amenity{Name: amenityName}).Error; err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to find or create amenity",
                    "error": err.Error(),
                })
            }

            if err := tx.Create(&model.ApartmentAmenity{
                ApartmentID: apartment.ID,
                AmenityID:   amenity.ID,
            }).Error; err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to link amenity to apartment",
                    "error": err.Error(),
                })
            }
        }
    }

    // 4. Update house rules if provided
    if input.HouseRules != nil {
        // Clear existing house rules
        if err := tx.Where("apartment_id = ?", apartment.ID).Delete(&model.ApartmentHouseRule{}).Error; err != nil {
            tx.Rollback()
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to clear existing house rules",
                "error": err.Error(),
            })
        }

        // Add new house rules
        for _, rule := range *input.HouseRules {
            var houseRule model.HouseRule
            if err := tx.FirstOrCreate(&houseRule, model.HouseRule{Rule: rule}).Error; err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to find or create house rule",
                    "error": err.Error(),
                })
            }

            if err := tx.Create(&model.ApartmentHouseRule{
                ApartmentID: apartment.ID,
                HouseRuleID: houseRule.ID,
            }).Error; err != nil {
                tx.Rollback()
                return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                    "message": "Failed to link house rule to apartment",
                    "error": err.Error(),
                })
            }
        }
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "message": "Transaction failed",
            "error": err.Error(),
        })
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": "Apartment updated successfully",
        "data": fiber.Map{
            "apartment":   apartment,
            "image_urls":  imageURLs,
            "video_urls":  videoURLs,
            "amenities":   input.Amenities,
            "house_rules": input.HouseRules,
        },
    })
}