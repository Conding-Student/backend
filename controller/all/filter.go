package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// FetchApprovedApartmentsForTenant returns filtered and sorted approved apartments
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
		RelevanceScore   int      `json:"-"`
	}

	// Get query parameters
	propertyTypes := c.Query("property_types")
	minPriceStr := c.Query("min_price")
	maxPriceStr := c.Query("max_price")
	amenitiesFilter := strings.Split(c.Query("amenities"), ",")
	houseRulesFilter := strings.Split(c.Query("house_rules"), ",")
	allowedGenders := c.Query("allowed_genders")

	var apartments []model.Apartment
	db := middleware.DBConn.Where("status = ?", "Approved")

	// Apply filters
	if propertyTypes != "" {
		propertyTypeList := strings.Split(propertyTypes, ",")
		for i := range propertyTypeList {
			propertyTypeList[i] = strings.TrimSpace(propertyTypeList[i])
		}
		db = db.Where("property_type IN ?", propertyTypeList)
	}

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

	if allowedGenders != "" {
		allowedGendersList := strings.Split(allowedGenders, ",")
		for i := range allowedGendersList {
			allowedGendersList[i] = strings.TrimSpace(allowedGendersList[i])
		}
		db = db.Where("allowed_gender IN ?", allowedGendersList)
	}

	if err := db.Find(&apartments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch apartments",
			"error":   err.Error(),
		})
	}

	var results []ApartmentDetails

	countMatches := func(apartmentItems, filterItems []string) int {
		count := 0
		for _, f := range filterItems {
			for _, item := range apartmentItems {
				if strings.EqualFold(strings.TrimSpace(item), strings.TrimSpace(f)) {
					count++
					break
				}
			}
		}
		return count
	}

	for _, apt := range apartments {
		var landlord model.User
		if err := middleware.DBConn.Where("uid = ?", apt.Uid).First(&landlord).Error; err != nil {
			continue
		}

		// Fetch images
		var images []model.ApartmentImage
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
		imageUrls := make([]string, len(images))
		for i, img := range images {
			imageUrls[i] = img.ImageURL
		}

		// Fetch videos
		var videos []model.ApartmentVideo
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&videos)
		videoUrls := make([]string, len(videos))
		for i, vid := range videos {
			videoUrls[i] = vid.VideoURL
		}

		// Fetch amenities
		var amenities []model.Amenity
		middleware.DBConn.
			Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
			Where("apartment_amenities.apartment_id = ?", apt.ID).
			Find(&amenities)
		amenityNames := make([]string, len(amenities))
		for i, a := range amenities {
			amenityNames[i] = a.Name
		}

		// Fetch house rules
		var houseRules []model.HouseRule
		middleware.DBConn.
			Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
			Where("apartment_house_rules.apartment_id = ?", apt.ID).
			Find(&houseRules)
		ruleNames := make([]string, len(houseRules))
		for i, r := range houseRules {
			ruleNames[i] = r.Rule
		}

		// Get inquiry count
		var inquiryCount int64
		middleware.DBConn.Model(&model.Inquiry{}).
			Where("apartment_id = ? AND (status = ? OR status = ?)", apt.ID, "Accepted", "Pending").
			Count(&inquiryCount)

		// Calculate relevance score
		score := 0
		score += countMatches(amenityNames, amenitiesFilter)
		score += countMatches(ruleNames, houseRulesFilter)

		if score > 0 || (c.Query("amenities") == "" && c.Query("house_rules") == "") {
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
				RelevanceScore:   score,
			})
		}
	}

	// Sort by relevance then by inquiries count
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RelevanceScore == results[j].RelevanceScore {
			return results[i].InquiriesCount > results[j].InquiriesCount
		}
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"apartments": results,
	})
}

// FetchSingleApartmentDetails returns complete details for a specific apartment
func FetchSingleApartmentDetails(c *fiber.Ctx) error {
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

	apartmentID := c.Params("id")

	var apt model.Apartment
	if err := middleware.DBConn.Where("id = ? AND (status = ? OR status = ?)", apartmentID, "Approved", "Pending").First(&apt).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Apartment not found or not approved/pending",
			"error":   err.Error(),
		})
	}

	var landlord model.User
	if err := middleware.DBConn.Where("uid = ?", apt.Uid).First(&landlord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Landlord not found",
			"error":   err.Error(),
		})
	}

	// Fetch images
	var images []model.ApartmentImage
	middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
	imageUrls := make([]string, len(images))
	for i, img := range images {
		imageUrls[i] = img.ImageURL
	}

	// Fetch videos
	var videos []model.ApartmentVideo
	middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&videos)
	videoUrls := make([]string, len(videos))
	for i, vid := range videos {
		videoUrls[i] = vid.VideoURL
	}

	// Fetch amenities
	var amenities []model.Amenity
	middleware.DBConn.
		Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
		Where("apartment_amenities.apartment_id = ?", apt.ID).
		Find(&amenities)
	amenityNames := make([]string, len(amenities))
	for i, a := range amenities {
		amenityNames[i] = a.Name
	}

	// Fetch house rules
	var houseRules []model.HouseRule
	middleware.DBConn.
		Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
		Where("apartment_house_rules.apartment_id = ?", apt.ID).
		Find(&houseRules)
	ruleNames := make([]string, len(houseRules))
	for i, r := range houseRules {
		ruleNames[i] = r.Rule
	}

	// Get inquiry count
	var inquiryCount int64
	middleware.DBConn.Model(&model.Inquiry{}).
		Where("apartment_id = ?", apt.ID).
		Count(&inquiryCount)

	result := ApartmentDetails{
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
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// SearchApartments returns apartments matching search criteria with full details
func SearchApartments(c *fiber.Ctx) error {
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

	var req struct {
		SearchTerm string `query:"search_term" validate:"required"`
	}

	if err := c.QueryParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid search parameters",
			"error":   err.Error(),
		})
	}

	if req.SearchTerm == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Search term is required",
		})
	}

	searchPattern := "%" + req.SearchTerm + "%"
	var apartments []model.Apartment
	if err := middleware.DBConn.Where("address ILIKE ?", searchPattern).Find(&apartments).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error",
			"error":   err.Error(),
		})
	}

	results := make([]ApartmentDetails, 0, len(apartments))

	for _, apt := range apartments {
		var landlord model.User
		if err := middleware.DBConn.Where("uid = ?", apt.Uid).First(&landlord).Error; err != nil {
			continue
		}

		// Fetch images
		var images []model.ApartmentImage
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&images)
		imageUrls := make([]string, len(images))
		for i, img := range images {
			imageUrls[i] = img.ImageURL
		}

		// Fetch videos
		var videos []model.ApartmentVideo
		middleware.DBConn.Where("apartment_id = ?", apt.ID).Find(&videos)
		videoUrls := make([]string, len(videos))
		for i, vid := range videos {
			videoUrls[i] = vid.VideoURL
		}

		// Fetch amenities
		var amenities []model.Amenity
		middleware.DBConn.
			Joins("JOIN apartment_amenities ON amenities.id = apartment_amenities.amenity_id").
			Where("apartment_amenities.apartment_id = ?", apt.ID).
			Find(&amenities)
		amenityNames := make([]string, len(amenities))
		for i, a := range amenities {
			amenityNames[i] = a.Name
		}

		// Fetch house rules
		var houseRules []model.HouseRule
		middleware.DBConn.
			Joins("JOIN apartment_house_rules ON house_rules.id = apartment_house_rules.house_rule_id").
			Where("apartment_house_rules.apartment_id = ?", apt.ID).
			Find(&houseRules)
		ruleNames := make([]string, len(houseRules))
		for i, r := range houseRules {
			ruleNames[i] = r.Rule
		}

		// Get inquiry count
		var inquiryCount int64
		middleware.DBConn.Model(&model.Inquiry{}).
			Where("apartment_id = ?", apt.ID).
			Count(&inquiryCount)

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

	// Sort by most popular first
	sort.Slice(results, func(i, j int) bool {
		return results[i].InquiriesCount > results[j].InquiriesCount
	})

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Apartments retrieved successfully",
		"data":    results,
	})
}
