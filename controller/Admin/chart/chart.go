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
	var years []int

	// If no years are provided in the query, fetch all available years from the database
	if yearsParam == "" {
		var availableYears []int
		query := `SELECT DISTINCT EXTRACT(YEAR FROM created_at)::int AS year FROM users ORDER BY year`
		if err := middleware.DBConn.Raw(query).Scan(&availableYears).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch available years",
			})
		}

		// Use the fetched years if no years were provided in the query
		years = availableYears
	} else {
		// Otherwise, parse the years from the query
		yearStrings := strings.Split(yearsParam, ",")
		for _, yearStr := range yearStrings {
			year, err := strconv.Atoi(strings.TrimSpace(yearStr))
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid year format in the query",
				})
			}
			years = append(years, year)
		}
	}

	// Prepare the SQL query to get the user count per year
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

	// Check if we have any results; if not, return a message with count as 0 for each year
	var finalResults []YearlyUserCount
	for _, year := range years {
		// Check if a result for the year exists
		var found bool
		for _, result := range results {
			if result.Year == year {
				finalResults = append(finalResults, result)
				found = true
				break
			}
		}
		// If no data for that year, create an entry with a count of 0
		if !found {
			finalResults = append(finalResults, YearlyUserCount{
				Year:  year,
				Count: 0,
			})
		}
	}

	// Return the final result with user count per year
	return c.JSON(finalResults)
}
