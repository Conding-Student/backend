package controller

import (
	"log"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"
	"github.com/Conding-Student/backend/model/response"

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

	// First try to find in users table
	var user model.User
	userResult := middleware.DBConn.Select("fullname").Where("uid = ?", uid).First(&user)

	if userResult.Error == nil {
		log.Printf("[DEBUG] Found user in users table for UID %s: %s", uid, user.Fullname)
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "Fullname retrieved successfully",
			Data: fiber.Map{
				"fullname": user.Fullname,
			},
		})
	}

	// If not found in users table, try admins table
	var admin model.Admins
	adminResult := middleware.DBConn.Select("fullname").Where("uid = ?", uid).First(&admin)

	if adminResult.Error == nil {
		log.Printf("[DEBUG] Found admin in admins table for UID %s: %s", uid, admin.Email)
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "Admin email retrieved successfully",
			Data: fiber.Map{
				// Using email since Admins struct doesn't have fullname
				"fullname": admin.Fullname, // Or you might want to return "Admin" as a role
			},
		})
	}

	log.Println("[ERROR] User/Admin not found for UID:", uid)
	return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
		RetCode: "404",
		Message: "User/Admin not found",
		Data:    nil,
	})
}

// GetUserRoleByUID retrieves the user's role based on their UID
func GetUserRoleByUID(c *fiber.Ctx) error {
	log.Println("[DEBUG] GetUserRoleByUID called")

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
	result := middleware.DBConn.Select("user_type").Where("uid = ?", uid).First(&user)
	if result.Error != nil {
		log.Println("[ERROR] User not found:", result.Error)
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
			Data:    nil,
		})
	}

	log.Printf("[DEBUG] User role for UID %s: %s", uid, user.UserType)

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User role retrieved successfully",
		Data: fiber.Map{
			"user_type": user.UserType,
		},
	})
}
func GetUserProfilePhotoByUID(c *fiber.Ctx) error {
	log.Println("[DEBUG] GetUserProfilePhotoByUID called")

	uid := c.Params("uid")
	if uid == "" {
		log.Println("[ERROR] UID param is missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "UID is required",
			Data:    nil,
		})
	}

	// First check in the User table
	var user model.User
	result := middleware.DBConn.Select("photo_url").Where("uid = ?", uid).First(&user)
	if result.Error == nil {
		log.Printf("[DEBUG] Profile photo URL for UID %s found in Users: %s", uid, user.PhotoURL)
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "Profile photo retrieved successfully",
			Data: fiber.Map{
				"photo_url": user.PhotoURL,
			},
		})
	}

	// If not found in User table, check Admins table
	var admin model.Admins
	result = middleware.DBConn.Select("photo_url").Where("uid = ?", uid).First(&admin)
	if result.Error == nil {
		log.Printf("[DEBUG] Profile photo URL for UID %s found in Admins: %s", uid, admin.PhotoURL)
		return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
			RetCode: "200",
			Message: "Profile photo retrieved successfully",
			Data: fiber.Map{
				"photo_url": admin.PhotoURL,
			},
		})
	}

	// If not found in either
	log.Println("[ERROR] UID not found in both Users and Admins")
	return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
		RetCode: "404",
		Message: "User not found",
		Data:    nil,
	})
}
