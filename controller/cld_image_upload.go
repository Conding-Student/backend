package controller

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"intern_template_v1/config"
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
)

func UploadApartmentImageHandler(c *fiber.Ctx) error {
    // Retrieve apartment ID from the request
	  apartmentIDStr := c.Params("apartmentID")
	  apartmentID64, err := strconv.ParseUint(apartmentIDStr, 10, 32) // Convert string to uint64
	  if err != nil {
		  return c.Status(http.StatusBadRequest).SendString("Invalid apartment ID")
	  }
	  apartmentID := uint(apartmentID64) // Convert uint64 to uint
    
    // Parse the file from the form
    fileHeader, err := c.FormFile("file")
    if err != nil {
        return c.Status(http.StatusBadRequest).SendString("File not found")
    }

    // Save file locally to temp
    tempDir := "./uploads"
    os.MkdirAll(tempDir, os.ModePerm)
    tempFilePath := filepath.Join(tempDir, fileHeader.Filename)

    file, err := fileHeader.Open()
    if err != nil {
        return c.Status(http.StatusInternalServerError).SendString("Failed to open uploaded file")
    }
    defer file.Close()

    outFile, err := os.Create(tempFilePath)
    if err != nil {
        return c.Status(http.StatusInternalServerError).SendString("Failed to create temp file")
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, file)
    if err != nil {
        return c.Status(http.StatusInternalServerError).SendString("Failed to save file")
    }

    // Upload to Cloudinary
    uploadedURL, err := config.UploadImage(tempFilePath)
    if err != nil {
        return c.Status(http.StatusInternalServerError).SendString(fmt.Sprintf("Cloudinary upload failed: %v", err))
    }

    // Store the image URL in the database
	db := middleware.DBConn.Begin()

    // Create a new ApartmentImage record with the Cloudinary URL
    apartmentImage := model.ApartmentImage{
        ApartmentID: apartmentID, // You should convert apartmentID to uint
        ImageURL:    uploadedURL,
    }

    if err := db.Create(&apartmentImage).Error; err != nil {
        return c.Status(http.StatusInternalServerError).SendString("Failed to save apartment image to database")
    }

    // Clean up the temp file
    os.Remove(tempFilePath)

    return c.Status(http.StatusOK).JSON(fiber.Map{
        "image_url": uploadedURL,
    })
}

func UploadVideoHandler(c *fiber.Ctx) error {
	// Get file from request
	file, err := c.FormFile("video")
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Failed to read video file")
	}

	// Save file temporarily
	filePath := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := c.SaveFile(file, filePath); err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to save video file")
	}

	// Upload video to Cloudinary
	url, err := config.UploadVideo(filePath)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to upload video to Cloudinary")
	}

	// Return the video URL from Cloudinary
	return c.JSON(fiber.Map{
		"message": "Video uploaded successfully",
		"url":     url,
	})
}
