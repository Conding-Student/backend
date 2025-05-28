package controller

import (
	"log"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ✅ **Register Admin (Insert into DB)**
func RegisterAdmin(c *fiber.Ctx) error {
	var req model.Admins

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// 🔐 **Hash the password before saving**
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server error",
		})
	}

	// ✅ Insert new admin into "admins" table
	newAdmin := model.Admins{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	result := middleware.DBConn.Table("admins").Create(&newAdmin)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register admin",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "Admin registered successfully",
		"admin_id":   newAdmin.ID,
		"created_at": newAdmin.CreatedAt,
	})
}
