package controller

import (
	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func UpdateAdminProfile(c *fiber.Ctx) error {
	type RequestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	updateData := map[string]interface{}{}

	if body.Email != "" {
		updateData["email"] = body.Email
	}

	if body.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "Error hashing password",
				"error":   err.Error(),
			})
		}
		updateData["password"] = string(hashedPassword)
	}

	if len(updateData) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "No fields provided to update",
		})
	}

	if err := middleware.DBConn.Model(&model.Admins{}).Where("id = ?", 1).Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update admin profile",
			"error":   err.Error(),
		})
	}

	// Get updated admin data for display
	var updatedAdmin model.Admins
	if err := middleware.DBConn.First(&updatedAdmin, 1).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve updated admin",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Admin profile updated successfully",
		"admin": fiber.Map{
			"email":    updatedAdmin.Email,
			"password": "********", // Masked
		},
	})
}
