package config

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Initialize Cloudinary client
var cld *cloudinary.Cloudinary

func InitCloudinary() {
	var err error
	cld, err = cloudinary.NewFromParams("rentxpert", "922588228586574", "NK4iVkSsSDlYiEE2WmxWH4L9fKc")
	if err != nil {
		log.Fatal("Failed to initialize Cloudinary: ", err)
	}
}

// Upload image to Cloudinary
func UploadImage(filePath string) (string, error) {
	// Upload image to Cloudinary
	resp, err := cld.Upload.Upload(context.Background(), filePath, uploader.UploadParams{
		ResourceType: "image", // Specify image type
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %v", err)
	}

	// Return the secure URL of the uploaded image
	return resp.SecureURL, nil
}


// Upload video to Cloudinary
func UploadVideo(filePath string) (string, error) {
	// Upload video to Cloudinary
	resp, err := cld.Upload.Upload(context.Background(), filePath, uploader.UploadParams{
		ResourceType: "video", // Specify video type
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload video to Cloudinary: %v", err)
	}

	// Return the secure URL of the uploaded video
	return resp.SecureURL, nil
}

