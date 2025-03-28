package controller

import (
	"fmt"
	"intern_template_v1/middleware"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// SendOTP handles sending OTP to the user's phone number
func SendOTP(c *fiber.Ctx) error {
	type Request struct {
		PhoneNumber string `json:"phone_number"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	if req.PhoneNumber == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Phone number is required"})
	}

	authClient := middleware.AuthClient

	// Generate OTP and send via Firebase
	session, err := authClient.SessionCookie(c.Context(), req.PhoneNumber, 10*60) // 10 minutes validity
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP"})
	}

	fmt.Println("📩 OTP sent to:", req.PhoneNumber)
	return c.JSON(fiber.Map{"message": "OTP sent successfully", "session_info": session})
}
