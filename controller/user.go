package controller

import (
	"log"

	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"intern_template_v1/model/response"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// GetUserProfile retrieves user profile based on JWT token and stores UserID in context
// GetUserProfile retrieves user profile based on JWT token
func GetUserProfile(c *fiber.Ctx) error {
	log.Println("[DEBUG] GetUserProfile called")

	// Get user claims from JWT stored in middleware
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		log.Println("[ERROR] JWT token is missing or invalid")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	log.Println("[DEBUG] JWT claims extracted:", userClaims)

	// Extract user data from JWT claims
	uid, uidOk := userClaims["uid"].(string)
	email, emailOk := userClaims["email"].(string)
	role, roleOk := userClaims["role"].(string)

	if !uidOk || !emailOk || !roleOk {
		log.Println("[ERROR] Missing required fields in JWT claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token data",
		})
	}

	log.Printf("[DEBUG] Extracted user info - UID: %s, Email: %s, Role: %s", uid, email, role)

	// ✅ Fetch user details from the database using UID (not "id")
	var user model.User
	result := middleware.DBConn.Where("uid = ?", uid).First(&user)

	if result.Error != nil {
		log.Println("[ERROR] Failed to fetch user profile:", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Could not retrieve user profile",
			Data:    nil,
		})
	}

	log.Printf("[DEBUG] User profile retrieved: %+v", user)

	// ✅ Return user profile
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User profile retrieved successfully",
		Data: fiber.Map{
			"uid":          user.Uid,
			"fullname":     user.Fullname,
			"email":        email,
			"phone_number": user.PhoneNumber,
			"photo_url":    user.PhotoURL,
			"address":      user.Address,
			"user_type":    user.UserType,
			"age":          user.Age,
			"birthday":     user.Birthday.Format("2006-01-02"), // format as string YYYY-MM-DD
		},
	})
}

func GetFullnameByUID(c *fiber.Ctx) error {
	log.Println("[DEBUG] GetFullnameByUID called")

	uid := c.Params("uid")
	if uid == "" {
		log.Println("[ERROR] UID param is missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "UID is required",
			Data:    nil,
		})
	}

	var user model.User
	result := middleware.DBConn.Select("fullname").Where("uid = ?", uid).First(&user)
	if result.Error != nil {
		log.Println("[ERROR] User not found:", result.Error)
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
			Data:    nil,
		})
	}

	log.Printf("[DEBUG] Fullname for UID %s: %s", uid, user.Fullname)

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Fullname retrieved successfully",
		Data: fiber.Map{
			"fullname": user.Fullname,
		},
	})
}
