package controller

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5" // Correct version for your case

	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Generate JWT Token using jwt/v5
func generateJWT(adminID uint, email string) (string, error) {
	// Get the secret key from the environment variable
	secretKey := os.Getenv("JWT_SECRET")

	// Create JWT claims
	claims := jwt.MapClaims{
		"adminID": adminID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	// Create the token using the claims and secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Admin Login Function with JWT
func LoginHandler(c *fiber.Ctx) error {
	var loginData model.Admins
	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Fetch the admin from the database using the provided email
	var admin model.Admins
	result := middleware.DBConn.Table("admins").Where("email = ?", loginData.Email).First(&admin)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Compare the provided password with the stored password hash
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(loginData.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Generate a JWT token after successful login
	token, err := generateJWT(admin.ID, admin.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate token",
		})
	}

	// Send the response with the JWT token
	return c.JSON(fiber.Map{
		"message": "Login successful",
		"admin": fiber.Map{
			"id":    admin.ID,
			"email": admin.Email,
		},
		"token": token, // Include the token in the response
	})
}
