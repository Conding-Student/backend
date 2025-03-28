package controller

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"intern_template_v1/model/response"
)

// LoginRequest struct allows login with either Email or Phone Number
type LoginRequest struct {
	Identifier string `json:"identifier"` // Can be Email or Phone Number
	Password   string `json:"password"`
}

// LoginUser authenticates a user and returns a JWT token
func LoginUser(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request format",
			Data:    nil,
		})
	}

	req.Identifier = strings.TrimSpace(req.Identifier)
	req.Password = strings.TrimSpace(req.Password)

	if req.Identifier == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Email/Phone and password are required",
			Data:    nil,
		})
	}

	var user model.User

	// Query user from DB using Email or PhoneNumber
	result := middleware.DBConn.Where("email = ? OR phone_number = ?", req.Identifier, req.Identifier).
		First(&user)

	if result.Error != nil {
		log.Printf("[ERROR] Database query error: %v", result.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Server error. Please try again later.",
			Data:    nil,
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "Account not found",
			Data:    nil,
		})
	}

	// Prevent login if account is "rejected"
	if user.AccountStatus == "Rejected" {
		log.Printf("[WARNING] Login attempt for rejected user: %s", user.Email)
		return c.Status(fiber.StatusForbidden).JSON(response.ResponseModel{
			RetCode: "403",
			Message: "Your account has been rejected. Please contact support.",
			Data:    nil,
		})
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("[WARNING] Incorrect password for user: %s", user.Email)
		return c.Status(fiber.StatusUnauthorized).JSON(response.ResponseModel{
			RetCode: "401",
			Message: "Incorrect password",
			Data:    nil,
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Email, user.UserType)
	if err != nil {
		log.Printf("[ERROR] Failed to generate JWT token: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Could not generate authentication token",
			Data:    nil,
		})
	}

	// ✅ Return user details, including AccountStatus
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Login successful",
		Data: fiber.Map{
			"id":             user.ID,
			"first_name":     user.FirstName,
			"middle_initial": user.MiddleInitial,
			"last_name":      user.LastName,
			"email":          user.Email,
			"phone_number":   user.PhoneNumber,
			"address":        user.Address,
			"user_type":      user.UserType,
			"account_status": user.AccountStatus, // ✅ Included to control UI access
			"token":          token,
		},
	})
}
