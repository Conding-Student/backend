package controller

// import (
// 	"intern_template_v1/model"   // Import the model for admin (if needed)
// 	"intern_template_v1/service" // Import your service
// 	"log"
// 	"net/http"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// // AdminLogin handles the admin login and returns a JWT token
// func AdminLogin(c *fiber.Ctx) error {
// 	var admin model.Admins

// 	// Bind the incoming JSON to the admin struct
// 	if err := c.BodyParser(&admin); err != nil {
// 		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid input",
// 		})
// 	}

// 	// Call the service to generate and save the admin token
// 	err := service.GenerateAndSaveAdminToken(c.Locals("db").(*gorm.DB), admin.ID, admin.Email)
// 	if err != nil {
// 		log.Println("Error generating or saving token:", err)
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to generate token",
// 		})
// 	}

// 	// Return success message
// 	return c.Status(http.StatusOK).JSON(fiber.Map{
// 		"message": "Admin logged in successfully",
// 	})
// }
