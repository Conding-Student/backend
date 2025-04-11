package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func CountUsersByType(c *fiber.Ctx) error {
	userType := c.Params("user_type") // e.g. /admin/count/All or /admin/count/Landlord

	var count int64
	var err error

	if userType == "" || strings.ToLower(userType) == "all" {
		// Count all users
		err = middleware.DBConn.Model(&model.User{}).Count(&count).Error
	} else {
		// Count users by specific type
		err = middleware.DBConn.Model(&model.User{}).
			Where("user_type = ?", userType).
			Count(&count).Error
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error counting users",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_type": userType,
		"count":     count,
	})
}

func CountApartmentsByStatus(c *fiber.Ctx) error {
	status := c.Params("status")

	if status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing status parameter",
		})
	}

	var count int64
	if err := middleware.DBConn.Model(&model.Apartment{}).
		Where("status = ?", status).
		Count(&count).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error counting apartments",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": status,
		"count":  count,
	})
}

func CountApartmentsByPropertyType(c *fiber.Ctx) error {
	propertyType := c.Params("property_type") // e.g. Apartment, Condo, All

	var count int64
	var err error

	if propertyType == "" || strings.ToLower(propertyType) == "all" {
		// Count all apartments regardless of property type
		err = middleware.DBConn.Model(&model.Apartment{}).Count(&count).Error
	} else {
		// Count apartments with specific property type
		err = middleware.DBConn.Model(&model.Apartment{}).
			Where("property_type = ?", propertyType).
			Count(&count).Error
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error counting apartments",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"property_type": propertyType,
		"count":         count,
	})
}
