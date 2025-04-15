package controller

import (
	"intern_template_v1/middleware"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetUserStatsByYear(c *fiber.Ctx) error {
	type YearlyUserCount struct {
		Year  int `json:"year"`
		Count int `json:"count"`
	}

	// Parse years from query (e.g., "2023,2024,2025")
	yearsParam := c.Query("years") // expected format: "2023,2024,2025"
	if yearsParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Please provide years as comma-separated values",
		})
	}

	// Split years and build query
	yearStrings := strings.Split(yearsParam, ",")
	var years []int
	for _, yearStr := range yearStrings {
		year, err := strconv.Atoi(strings.TrimSpace(yearStr))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid year format in the query",
			})
		}
		years = append(years, year)
	}

	// Build SQL query with placeholders for years
	query := `
		SELECT EXTRACT(YEAR FROM created_at) AS year, COUNT(*) 
		FROM users 
		WHERE EXTRACT(YEAR FROM created_at) IN (?) 
		GROUP BY year 
		ORDER BY year`

	// Execute the query with the correct year values
	var results []YearlyUserCount
	if err := middleware.DBConn.Raw(query, years).Scan(&results).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching user stats",
			"msg":   err.Error(),
		})
	}

	// Return the result
	return c.JSON(results)
}

func GetAvailableUserYears(c *fiber.Ctx) error {
	var years []int
	query := `SELECT DISTINCT EXTRACT(YEAR FROM created_at)::int AS year FROM users ORDER BY year`
	if err := middleware.DBConn.Raw(query).Scan(&years).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch year list",
		})
	}

	// // Optional: Include future year like 2030
	// currentYear := time.Now().Year()
	// futureYear := 2030
	// if futureYear > currentYear {
	// 	years = append(years, futureYear)
	// }

	return c.JSON(years)
}
