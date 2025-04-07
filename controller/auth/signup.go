package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// CalculateAge calculates age based on the given birthdate
func CalculateAge(birthday string) (int, error) {
	// Parse the birthday string
	birthDate, err := time.Parse("2006-01-02", birthday)
	if err != nil {
		return 0, err
	}

	// Get the current date
	now := time.Now()

	// Calculate age
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age-- // Adjust if birthday hasn't occurred yet this year
	}

	return age, nil
}

func SaveUser(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		log.Println("Error parsing request body:", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}


	// Check if user already exists
	var existingUser model.User
	result := middleware.DBConn.Where("email = ?", user.Email).First(&existingUser)
	if result.RowsAffected > 0 {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "User already exists",
		})
	}

	// Save user to database with age
	if err := middleware.DBConn.Create(&user).Error; err != nil {
		log.Println("Error saving user:", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save user",
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "User saved successfully",
		"user":    user,
	})
}
