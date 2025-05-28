package controller

import (
	"net/http"
	"strconv"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// DeleteApartmentRequest is used to confirm deletion
type DeleteApartmentRequest struct {
	Confirm bool `json:"Confirm"`
}

// DeleteApartment handles deletion of a rejected apartment along with cascading related data.
// It verifies the landlord's UID via JWT, checks that the apartment status is "Rejected",
// and executes the deletion if confirmed.
func DeleteApartment(c *fiber.Ctx) error {
	// 🔐 Extract JWT claims to get the landlord UID.
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// 🆔 Parse the apartment ID from URL parameter
	idParam := c.Params("id")
	apartmentID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid apartment id parameter",
			"error":   err.Error(),
		})
	}

	// 📥 Parse request body to confirm deletion
	var delReq DeleteApartmentRequest
	if err := c.BodyParser(&delReq); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	if !delReq.Confirm {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Deletion not confirmed",
		})
	}

	// 🔍 Retrieve the apartment by id and uid to verify status
	var apartment model.Apartment
	if err := middleware.DBConn.Where("id = ? AND uid = ?", apartmentID, uid).First(&apartment).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Apartment not found",
			"error":   err.Error(),
		})
	}

	// ❗ Only allow deletion if the apartment's status is "Rejected"
	if apartment.Status != "Rejected" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Only rejected apartments can be deleted",
		})
	}

	// 🗑 Delete the apartment via raw SQL (cascading deletions will occur based on your DB constraints)
	result := middleware.DBConn.Exec("DELETE FROM apartments WHERE id = ? AND uid = ?", apartment.ID, uid)
	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete apartment",
			"error":   result.Error.Error(),
		})
	}

	// Check if any rows were affected.
	if result.RowsAffected == 0 {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "No apartment deleted. It may not exist or you might not have permission.",
		})
	}

	// 🎉 Successfully deleted
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Apartment and all related data deleted successfully",
	})
}

// DeleteApartment allows landlords to delete any apartment they own, regardless of status,
// as long as they confirm the action. It also cascades deletions via DB constraints.
func DeleteApartmentAny(c *fiber.Ctx) error {
	// 🔐 Extract JWT claims to get the landlord UID
	userClaims, ok := c.Locals("user").(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Missing JWT claims",
		})
	}

	uid, ok := userClaims["uid"].(string)
	if !ok || uid == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: Invalid landlord UID",
		})
	}

	// 🆔 Parse apartment ID from the route
	idParam := c.Params("id")
	apartmentID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid apartment ID parameter",
			"error":   err.Error(),
		})
	}

	// 🔍 Check if apartment exists and belongs to the landlord
	var apartment model.Apartment
	if err := middleware.DBConn.Where("id = ? AND uid = ?", apartmentID, uid).First(&apartment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Apartment not found or you do not have permission",
			"error":   err.Error(),
		})
	}

	// 🗑 Delete apartment (cascading deletes via foreign key constraints)
	result := middleware.DBConn.Exec("DELETE FROM apartments WHERE id = ? AND uid = ?", apartment.ID, uid)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete apartment",
			"error":   result.Error.Error(),
		})
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No apartment deleted. It may not exist or you might not have permission.",
		})
	}

	// ✅ Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Apartment and all related data deleted successfully",
	})
}
