package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"github.com/gofiber/fiber/v2"
)

// Get pending status
func GetPendingApartments(c *fiber.Ctx) error {
	var pendingApartments []model.Apartment

	// Fetch apartments where status is "Pending"
	result := middleware.DBConn.Where("status = ?", "Pending").Find(&pendingApartments)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch pending apartments",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "Pending apartments retrieved successfully",
		"apartments": pendingApartments,
	})
}

type VerifyApartmentRequest struct {
	Status string `json:"status"` // Expected values: "Approved" or "Rejected"
}

// Verify (Approve/Reject) an Apartment
func VerifyApartment(c *fiber.Ctx) error {
	apartmentID := c.Params("id") // Get apartment ID from the URL
	var req VerifyApartmentRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Check if the provided status is valid
	if req.Status != "Approved" && req.Status != "Rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Use 'Approved' or 'Rejected'.",
		})
	}

	var apartment model.Apartment
	result := middleware.DBConn.First(&apartment, apartmentID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Apartment not found",
		})
	}
	// If status is "Rejected", update status in DB
	if req.Status == "Rejected" {
		apartment.Status = "Rejected"
		middleware.DBConn.Save(&apartment)

		// Notify the landlord (assuming a notification system is implemented)
		return c.JSON(fiber.Map{
			"message":      "Apartment rejected. Waiting for landlord confirmation to delete.",
			"apartment_id": apartmentID,
			"status":       "Rejected",
		})
	}

	// Update the status
	apartment.Status = req.Status
	middleware.DBConn.Save(&apartment)

	return c.JSON(fiber.Map{
		"message":      "Apartment status updated successfully",
		"apartment_id": apartmentID,
		"status":       req.Status,
	})
}

type UpdateApartmentRequest struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Location string  `json:"location"`
	Price    float64 `json:"price"`
}

// UpdateApartmentInfo allows editing PropertyName, PropertyType, Address, and RentPrice
func UpdateApartmentInfo(c *fiber.Ctx) error {
	apartmentID := c.Params("id") // Apartment ID from URL
	var req struct {
		PropertyName string  `json:"property_name"`
		PropertyType string  `json:"property_type"`
		Address      string  `json:"address"`
		RentPrice    float64 `json:"rent_price"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Retrieve apartment
	var apartment model.Apartment
	result := middleware.DBConn.First(&apartment, apartmentID)
	if result.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Apartment not found",
		})
	}

	// Update the fields
	apartment.PropertyName = req.PropertyName
	apartment.PropertyType = req.PropertyType
	apartment.Address = req.Address
	apartment.RentPrice = req.RentPrice

	// Save updates
	if err := middleware.DBConn.Save(&apartment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update apartment info",
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Apartment information updated successfully",
		"apartment_id": apartmentID,
		"updated_info": fiber.Map{
			"property_name": apartment.PropertyName,
			"property_type": apartment.PropertyType,
			"address":       apartment.Address,
			"rent_price":    apartment.RentPrice,
		},
	})
}

// Delete Apartment when landlord confirms
// func ConfirmLandlord(c *fiber.Ctx) error {
// 	apartmentID := c.Params("uid")
// 	var req model.DeleteApartmentRequest

// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request format",
// 		})
// 	}

// 	if !req.Confirm {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Landlord must confirm deletion",
// 		})
// 	}

// 	var apartment model.Apartment
// 	result := middleware.DBConn.First(&apartment, apartmentID)
// 	if result.Error != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"error": "Apartment not found",
// 		})
// 	}

// 	// //Delete if only "rejected"
// 	if apartment.Status != "Rejected" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Only rejected apartments can be deleted",
// 		})
// 	}

// 	// Delete the apartment
// 	deleteResult := middleware.DBConn.Delete(&apartment)
// 	if deleteResult.Error != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to delete apartment",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "Apartment deleted successfully",
// 	})
// }
