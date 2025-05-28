package handlers

import (
	"github.com/Conding-Student/backend/config" // adjust import path

	"github.com/gofiber/fiber/v2"
)

// TrackNotificationOpenHandler handles the POST request to mark a notification as opened.
func TrackNotificationOpenHandler(c *fiber.Ctx) error {
	logId := c.Params("logId")
	if logId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "logId is required",
		})
	}

	err := config.TrackNotificationOpen(logId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Failed to track notification open",
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Notification marked as opened",
	})
}
