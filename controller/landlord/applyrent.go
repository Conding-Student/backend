package controller

import (
	"intern_template_v1/config"
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Struct for parsing apartment creation request
type ApartmentRequest struct {
	PropertyName  string   `json:"property_name"`
	PropertyType  string   `json:"property_type"`
	RentPrice     float64  `json:"rent_price"`
	LocationLink  string   `json:"location_link"`
	Landmarks     string   `json:"landmarks"`
	Amenities     []string `json:"amenities"`
	HouseRules    []string `json:"house_rules"`
	ImageURLs     []string `json:"image_urls"`
	VideoURLs     []string `json:"video_urls"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	AllowedGender string   `json:"allowed_gender"` // New field
}

func CreateApartment(c *fiber.Ctx) error {
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

	var req ApartmentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	if err := middleware.DBConn.Where("uid = ? AND user_type = ?", uid, "Landlord").First(&model.User{}).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: User is not a registered landlord",
		})
	}

	if req.PropertyName == "" || req.PropertyType == "" || req.RentPrice <= 0 || req.LocationLink == "" || req.AllowedGender == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing required fields: property_name, property_type, rent_price, location_link, or allowed_gender",
		})
	}

	if req.Latitude == 0 || req.Longitude == 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Latitude and Longitude are required and must be valid coordinates",
		})
	}

	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to start transaction",
		})
	}

	var existing model.Apartment
	if err := tx.Where("property_name = ? AND location_link = ? AND uid = ?", req.PropertyName, req.LocationLink, uid).First(&existing).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"message": "Apartment with the same property name and location already exists for this landlord",
		})
	}

	apartment := model.Apartment{
		Uid:            uid,
		PropertyName:   req.PropertyName,
		PropertyType:   req.PropertyType,
		RentPrice:      req.RentPrice,
		LocationLink:   req.LocationLink,
		Landmarks:      req.Landmarks,
		Status:         "Pending",
		Latitude:       req.Latitude,
		Longitude:      req.Longitude,
		UserID:         uid,
		Allowed_Gender: req.AllowedGender,
	}

	if err := tx.Create(&apartment).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Unable to create apartment",
			"error":   err.Error(),
		})
	}

	for _, name := range req.Amenities {
		var a model.Amenity
		if err := tx.Where("name = ?", name).FirstOrCreate(&a, model.Amenity{Name: name}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Database error: Unable to add amenities", "error": err.Error()})
		}
		tx.Create(&model.ApartmentAmenity{ApartmentID: apartment.ID, AmenityID: a.ID})
	}

	for _, rule := range req.HouseRules {
		var h model.HouseRule
		if err := tx.Where("rule = ?", rule).FirstOrCreate(&h, model.HouseRule{Rule: rule}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Database error: Unable to add house rules", "error": err.Error()})
		}
		tx.Create(&model.ApartmentHouseRule{ApartmentID: apartment.ID, HouseRuleID: h.ID})
	}

	var imageURLs []string
	for _, img := range req.ImageURLs {
		url, err := config.UploadImage(img)
		if err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to upload image to Cloudinary", "error": err.Error()})
		}
		imageURLs = append(imageURLs, url)
		tx.Create(&model.ApartmentImage{ApartmentID: apartment.ID, ImageURL: url})
	}

	var videoURLs []string
	for _, vid := range req.VideoURLs {
		url, err := config.UploadVideo(vid)
		if err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to upload video to Cloudinary", "error": err.Error()})
		}
		videoURLs = append(videoURLs, url)
		tx.Create(&model.ApartmentVideo{ApartmentID: apartment.ID, VideoURL: url})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error: Transaction commit failed",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Apartment created successfully",
		"data": fiber.Map{
			"apartment_id": apartment.ID,
			"image_urls":   imageURLs,
			"video_urls":   videoURLs,
		},
	})
}
