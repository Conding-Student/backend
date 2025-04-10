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
    PropertyName string   `json:"property_name"`
    PropertyType string   `json:"property_type"`
    RentPrice    float64  `json:"rent_price"`
    LocationLink string   `json:"location_link"`
    Landmarks    string   `json:"landmarks"`
    Amenities    []string `json:"amenities"`
    HouseRules   []string `json:"house_rules"`
    ImageURLs    []string `json:"image_urls"`  // for images
    VideoURLs    []string `json:"video_urls"`  // for videos
    Latitude     float64  `json:"latitude"`    // New field for latitude
    Longitude    float64  `json:"longitude"`   // New field for longitude
}

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

    // Parse request body
    var req ApartmentRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{
            "message": "Invalid request format",
            "error":   err.Error(),
        })
    }

    // Validate required fields
    if req.PropertyName == "" || req.PropertyType == "" || req.RentPrice <= 0 || req.LocationLink == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{
            "message": "Missing required fields: property_name, property_type, rent_price, or location_link",
        })
    }

    // Validate latitude and longitude
    if req.Latitude == 0 || req.Longitude == 0 {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{
            "message": "Latitude and Longitude are required and must be valid coordinates",
        })
    }

    // Start transaction
    tx := middleware.DBConn.Begin()
    if tx.Error != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
            "message": "Database error: Unable to start transaction",
        })
    }

    // Create the apartment
    apartment := model.Apartment{
        Uid:          uid,
        PropertyName: req.PropertyName,
        PropertyType: req.PropertyType,
        RentPrice:    req.RentPrice,
        LocationLink: req.LocationLink,
        Landmarks:    req.Landmarks,
        Status:       "Pending", // Default status
        Latitude:     req.Latitude,
        Longitude:    req.Longitude,
    }

    // Insert apartment into the apartments table
    if err := tx.Create(&apartment).Error; err != nil {
        tx.Rollback()
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
            "message": "Database error: Unable to create apartment",
            "error":   err.Error(),
        })
    }

    // Upload images to Cloudinary
    var imageURLs []string
    for _, image := range req.ImageURLs {
        // Call Cloudinary upload logic
        uploadedURL, err := config.UploadImage(image)
        if err != nil {
            tx.Rollback()
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to upload image to Cloudinary",
                "error":   err.Error(),
            })
        }
        imageURLs = append(imageURLs, uploadedURL)
    }

    // Upload videos to Cloudinary
    var videoURLs []string
    for _, video := range req.VideoURLs {
        // Call Cloudinary upload logic
        uploadedURL, err := config.UploadVideo(video)
        if err != nil {
            tx.Rollback()
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to upload video to Cloudinary",
                "error":   err.Error(),
            })
        }
        videoURLs = append(videoURLs, uploadedURL)
    }

    // Insert image URLs into the database
    for _, imageURL := range imageURLs {
        apartmentImage := model.ApartmentImage{
            ApartmentID: apartment.ID,
            ImageURL:    imageURL,
        }
        if err := tx.Create(&apartmentImage).Error; err != nil {
            tx.Rollback()
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to save apartment image URL to database",
                "error":   err.Error(),
            })
        }
    }

    // Insert video URLs into the database (optional)
    for _, videoURL := range videoURLs {
        apartmentVideo := model.ApartmentVideo{
            ApartmentID: apartment.ID,
            VideoURL:    videoURL,
        }
        if err := tx.Create(&apartmentVideo).Error; err != nil {
            tx.Rollback()
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
                "message": "Failed to save apartment video URL to database",
                "error":   err.Error(),
            })
        }
    }

    // Commit the transaction
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
            "apartment_id": apartment.ID,
            "image_urls":   imageURLs,
            "video_urls":   videoURLs,
        },
    })
}
