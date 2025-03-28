package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ✅ **Admin Login Function**
func LoginHandler(c *fiber.Ctx) error {
	var loginData model.Admins
	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	var admin model.Admins
	result := middleware.DBConn.Table("admins").Where("email = ?", loginData.Email).First(&admin)
	if result.Error != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// 🔐 **Check the password**
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(loginData.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"admin": fiber.Map{
			"id":    admin.ID,
			"email": admin.Email,
		},
	})
}
