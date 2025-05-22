package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
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

func Signup(c *fiber.Ctx) error {
	var user model.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}
	
	// Database operation
	result := middleware.DBConn.Create(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save user",
		})
	}
	
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User saved successfully",
		"user":    user,
	})
}