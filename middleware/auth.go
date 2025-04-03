package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifies the JWT token and extracts user claims
func AuthMiddleware(c *fiber.Ctx) error {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		log.Println("[ERROR] Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Missing authentication token",
		})
	}

	// Extract token (format: "Bearer <token>")
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		log.Println("[ERROR] Invalid token format")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token format",
		})
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		log.Println("[ERROR] Invalid token:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid or expired token",
		})
	}

	// Extract user claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("[ERROR] Invalid token claims")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid token claims",
		})
	}

	log.Println("[INFO] Token validated successfully, user claims:", claims)

	// Store user details in Fiber Locals
	c.Locals("user", claims)

	// ✅ Token is valid, proceed to next handler
	return c.Next()
}
