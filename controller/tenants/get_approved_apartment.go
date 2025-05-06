package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"sort"
	"strconv"
	"sync"

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

	// Get query parameters for filtering
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	gender := c.Query("gender")

	// Base query with indexes utilization
	query := middleware.DBConn.
		Where("status = ?", "Approved").
		Where("availability = ?", "Available").
		Order("created_at DESC")

	// Price filter
	if minPrice != "" && maxPrice != "" {
		query = query.Where("rent_price BETWEEN ? AND ?", minPrice, maxPrice)
	}

	// Gender filter
	if gender != "" {
		query = query.Where("allowed_gender = ?", gender)
	}

	var apartments []model.Apartment
	if err := query.Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch approved apartments",
			"error":   err.Error(),
		})
	}

	var results []ApartmentDetails

	for _, apt := range apartments {
		// Skip apartments that became unavailable after expiration check
		if apt.Availability != "Available" {
			continue
		}

		var landlord model.User
		if err := middleware.DBConn.
			Where("uid = ?", apt.Uid).
			First(&landlord).Error; err != nil {
			continue // Skip if landlord not found
		}

		// Concurrently fetch media and details
		var (
			images       []model.ApartmentImage
			videos       []model.ApartmentVideo
			amenities    []model.Amenity
			houseRules   []model.HouseRule
			inquiryCount int64
		)

		wg := &sync.WaitGroup{}
		wg.Add(5)

		go func() {
			middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
			wg.Done()
		}()

		go func() {
			middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&videos)
			wg.Done()
		}()

		go func() {
			middleware.DBConn.
				Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
				Where("apartment_amenities.apartment_id = ?", apt.ID).
				Find(&amenities)
			wg.Done()
		}()

		go func() {
			middleware.DBConn.
				Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
				Where("apartment_house_rules.apartment_id = ?", apt.ID).
				Find(&houseRules)
			wg.Done()
		}()

		go func() {
			middleware.DBConn.Model(&model.Inquiry{}).
				Where("property_id = ?", apt.ID).
				Count(&inquiryCount)
			wg.Done()
		}()

		wg.Wait()

		// Convert to simple arrays
		imageUrls := make([]string, len(images))
		for i, img := range images {
			imageUrls[i] = img.ImageURL
		}

		videoUrls := make([]string, len(videos))
		for i, vid := range videos {
			videoUrls[i] = vid.VideoURL
		}

		amenityNames := make([]string, len(amenities))
		for i, a := range amenities {
			amenityNames[i] = a.Name
		}

		ruleNames := make([]string, len(houseRules))
		for i, r := range houseRules {
			ruleNames[i] = r.Rule
		}

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

	// Sort by popularity (inquiries count) if requested
	if c.Query("sort") == "popularity" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].InquiriesCount > results[j].InquiriesCount
		})
	}

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
	start := (page - 1) * pageSize
	end := start + pageSize

	if end > len(results) {
		end = len(results)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"total":      len(results),
		"page":       page,
		"page_size":  pageSize,
		"apartments": results[start:end],
	})
}
