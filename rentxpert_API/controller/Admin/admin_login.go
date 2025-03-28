package controller

import (
	"intern_template_v1/middleware"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ✅ Struct for admin login
type AdminLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ✅ **Admin Login Function**
func LoginHandler(c *fiber.Ctx) error {
	var loginData AdminLogin
	if err := c.BodyParser(&loginData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	var admin Admin
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
