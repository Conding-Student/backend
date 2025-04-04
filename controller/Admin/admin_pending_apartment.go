package controller

// import (
// 	"intern_template_v1/middleware"
// 	"intern_template_v1/model"

// 	"github.com/gofiber/fiber/v2"
// )

// func GetPendingApartments(c *fiber.Ctx) error {
// 	var pendingApartments []model.Apartment

// 	// 🔍 Query only apartments with status = 'Pending'
// 	if err := middleware.DBConn.
// 		Where("status = ?", "Pending").
// 		Find(&pendingApartments).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"message": "Failed to retrieve pending apartments",
// 			"error":   err.Error(),
// 		})
// 	}

// 	// ✅ Success response
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"message":    "Pending apartments retrieved successfully",
// 		"apartments": pendingApartments,
// 	})
// }
