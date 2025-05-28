// controller/user_search.go
package controller

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"
	"github.com/Conding-Student/backend/model/response"

	"github.com/gofiber/fiber/v2"
)

// Search request structure
type UserSearchRequest struct {
	SearchField string `query:"field" validate:"required"`
	SearchTerm  string `query:"search_term" validate:"required"`
}

// 👤 SearchUsers endpoint
func Apartmentfilteradmin(c *fiber.Ctx) error {
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
		"property_name":  true,
		"address":        true,
		"property_type":  true,
		"landmarks":      true,
		"status":         true,
		"allowed_gender": true,
		"availability":   true,
	}

	if !allowedFields[req.SearchField] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid search field",
			"allowed_fields": []string{
				"uid", "property_name", "property_type",
				"landmarks", "address",
				"status", "allowed_gender", "availability",
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
	var users []model.Apartment
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

func GetFilteredUserDetailspart2(c *fiber.Ctx) error {
	// 🔍 Existing Filters
	userType := c.Query("user_type", "")
	accountStatus := c.Query("account_status", "")
	name := c.Query("name", "")

	// 🔍 New Search Filters
	searchField := c.Query("field", "")
	searchTerm := c.Query("search_term", "")

	// 📄 Pagination
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	// Validate search parameters if present
	if searchField != "" || searchTerm != "" {
		// Ensure both search parameters are present
		if searchField == "" || searchTerm == "" {
			return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
				RetCode: "400",
				Message: "Both field and search_term are required for search",
				Data:    nil,
			})
		}

		// Validate search term is not empty
		if searchTerm == "" {
			return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
				RetCode: "400",
				Message: "Search term cannot be empty",
				Data:    nil,
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

		if !allowedFields[searchField] {
			allowedFieldsList := make([]string, 0, len(allowedFields))
			for k := range allowedFields {
				allowedFieldsList = append(allowedFieldsList, k)
			}
			return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
				RetCode: "400",
				Message: "Invalid search field",
				Data: fiber.Map{
					"allowed_fields": allowedFieldsList,
				},
			})
		}
	}

	// Convert pagination parameters
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Base query with account status filter
	query := middleware.DBConn.Table("users").
		Select("id, uid, email, phone_number, fullname, address, valid_id, account_status, user_type").
		Where("account_status IN ?", []string{"Unverified", "Pending", "Verified"})

	// Apply filters
	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if accountStatus != "" {
		query = query.Where("account_status = ?", accountStatus)
	}
	if name != "" {
		query = query.Where("LOWER(fullname) LIKE ?", "%"+strings.ToLower(name)+"%")
	}
	if searchField != "" && searchTerm != "" {
		query = query.Where(searchField+" ILIKE ?", "%"+searchTerm+"%")
	}

	// Check if any filters were applied (excluding base account_status filter)
	hasFilters := userType != "" || accountStatus != "" || name != "" || (searchField != "" && searchTerm != "")

	// Apply default tenant filter if no parameters
	if !hasFilters {
		query = query.Where("user_type = ?", "Tenant")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Println("[ERROR] Failed to count users:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to count user data",
			Data:    nil,
		})
	}

	// Prevent out-of-bound pages
	if offset >= int(total) {
		page = 1
		offset = 0
	}

	// Fetch paginated results
	var users []struct {
		ID            int    `json:"id" gorm:"column:id"` // Explicitly map to the "id" column
		UID           string `json:"uid"`
		Email         string `json:"email"`
		PhoneNumber   string `json:"phone_number"`
		FullName      string `json:"fullname" gorm:"column:fullname"`
		Address       string `json:"address"`
		ValidID       string `json:"valid_id"`
		AccountStatus string `json:"account_status"`
		UserType      string `json:"user_type"`
	}

	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		log.Println("[ERROR] Failed to fetch users:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user data",
			Data:    nil,
		})
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Return paginated response
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Paginated user list retrieved successfully",
		Data: fiber.Map{
			"limit":       limit,
			"page":        page,
			"total":       total,
			"total_pages": totalPages,
			"users":       users,
		},
	})
}
