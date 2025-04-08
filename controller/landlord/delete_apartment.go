package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"
	"strconv"

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
