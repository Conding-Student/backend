package controller

import (
	"intern_template_v1/middleware"
	"log"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ✅ Struct for admin registration
type AdminRegister struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ✅ Struct for the "admins" table
type Admin struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ✅ **Register Admin (Insert into DB)**
func RegisterAdmin(c *fiber.Ctx) error {
	var req AdminRegister

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
	newAdmin := Admin{
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
		"message":  "Admin registered successfully",
		"admin_id": newAdmin.ID,
	})
}
