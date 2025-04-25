// controller/user_search.go
package controller

import (
	"fmt"
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Search request structure
type UserSearchRequest struct {
	SearchField string `query:"field" validate:"required"`
	SearchTerm  string `query:"search_term" validate:"required"`
}

// 👤 SearchUsers endpoint
func SearchUsers(c *fiber.Ctx) error {
	// Parse search parameters
	var req UserSearchRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid search parameters",
			"error":   err.Error(),
		})
	}

	// Validate allowed search fields
	allowedFields := map[string]bool{
		"uid":            true,
		"email":          true,
		"phone_number":   true,
		"fullname":       true,
		"address":        true,
		"account_status": true,
		"user_type":      true,
	}

	if !allowedFields[req.SearchField] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid search field",
			"allowed_fields": []string{
				"uid", "email", "phone_number",
				"fullname", "address",
				"account_status", "user_type",
			},
		})
	}

	// Validate search term
	if req.SearchTerm == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Search term cannot be empty",
		})
	}

	// Build search pattern
	searchPattern := "%" + req.SearchTerm + "%"

	// Execute database query
	var users []model.User
	result := middleware.DBConn.
		Where(req.SearchField+" ILIKE ?", searchPattern).
		Find(&users)

	if result.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Database error",
			"error":   result.Error.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Users found by %s", req.SearchField),
		"data":    users,
	})
}
