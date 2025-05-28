package controller

import (
	"strings"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
)

func CountUsersByType(c *fiber.Ctx) error {
	userType := c.Params("user_type") // e.g. /admin/count/All or /admin/count/Landlord

	var count int64
	var err error

	if userType == "" || strings.ToLower(userType) == "all" {
		// Count all users with specific statuses
		err = middleware.DBConn.Model(&model.User{}).
			Where("account_status IN ?", []string{"Unverified", "Pending", "Verified"}).
			Count(&count).Error
	} else {
		// Count users by specific type with specific statuses
		err = middleware.DBConn.Model(&model.User{}).
			Where("user_type = ?", userType).
			Where("account_status IN ?", []string{"Unverified", "Pending", "Verified"}).
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

	// Only count apartments that are "Approved"
	if propertyType == "" || strings.ToLower(propertyType) == "all" {
		// Count all "Approved" apartments
		err = middleware.DBConn.Model(&model.Apartment{}).
			Where("status = ?", "Approved").
			Count(&count).Error
	} else {
		// Count "Approved" apartments with specific property type
		err = middleware.DBConn.Model(&model.Apartment{}).
			Where("property_type = ? AND status = ?", propertyType, "Approved").
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

func CountApartmentsByStatusAndType(c *fiber.Ctx) error {
	status := c.Params("status")              // e.g. Approved, Pending
	propertyType := c.Params("property_type") // e.g. Apartment, Condo, All

	if status == "" || (strings.ToLower(status) != "pending" && strings.ToLower(status) != "approved") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid or missing status. Must be 'Pending' or 'Approved'.",
		})
	}

	var count int64
	var err error

	query := middleware.DBConn.Model(&model.Apartment{}).Where("status = ?", status)

	if propertyType != "" && strings.ToLower(propertyType) != "all" {
		query = query.Where("property_type = ?", propertyType)
	}

	err = query.Count(&count).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error counting apartments",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":        status,
		"property_type": propertyType,
		"count":         count,
	})
}

// counting users based on account status
func CountUsersByStatusAndType(c *fiber.Ctx) error {
	accountStatus := c.Params("account_status") // e.g. Pending, Verified
	userType := c.Params("user_type")           // e.g. Landlord, Tenant

	if accountStatus == "" || (strings.ToLower(accountStatus) != "pending" && strings.ToLower(accountStatus) != "verified") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid or missing account_status. Must be 'Pending' or 'Verified'.",
		})
	}

	if userType == "" || (strings.ToLower(userType) != "landlord" && strings.ToLower(userType) != "tenant") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid or missing user_type. Must be 'Landlord' or 'Tenant'.",
		})
	}

	var count int64
	err := middleware.DBConn.Model(&model.User{}).
		Where("account_status = ? AND user_type = ?", accountStatus, userType).
		Count(&count).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Error counting users",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"account_status": accountStatus,
		"user_type":      userType,
		"count":          count,
	})
}
