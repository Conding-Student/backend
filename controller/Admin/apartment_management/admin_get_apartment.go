package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"intern_template_v1/model/response"
	"log"
	"math"
	"strconv"
	"strings"

	//"time"

	"github.com/gofiber/fiber/v2"
)

func GetFilteredApartments(c *fiber.Ctx) error {
	// 🔍 Filters
	propertyName := c.Query("property_name", "")
	propertyType := c.Query("property_type", "")
	rentPrice := c.Query("rent_price", "")
	status := c.Query("status", "")
	landlordName := c.Query("landlord_name", "")
	amenity := c.Query("amenities", "")
	houseRule := c.Query("house_rules", "")

	// 📄 Pagination
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// 🏘️ Base query
	query := middleware.DBConn.Table("apartments").
		Select("apartments.*, users.fullname as landlord_name").
		Joins("JOIN users ON users.uid = apartments.user_id")

	// ✅ Apply filters
	if propertyName != "" {
		query = query.Where("LOWER(apartments.property_name) LIKE ?", "%"+strings.ToLower(propertyName)+"%")
	}
	if propertyType != "" {
		query = query.Where("apartments.property_type = ?", propertyType)
	}
	if rentPrice != "" {
		query = query.Where("CAST(apartments.rent_price AS TEXT) LIKE ?", "%"+rentPrice+"%")
	}
	if status != "" {
		query = query.Where("apartments.status = ?", status)
	}
	if landlordName != "" {
		query = query.Where("LOWER(users.fullname) LIKE ?", "%"+strings.ToLower(landlordName)+"%")
	}

	// 🔗 Join with amenities if filtering
	if amenity != "" {
		query = query.Joins("JOIN apartment_amenities aa ON aa.apartment_id = apartments.id").
			Joins("JOIN amenities a ON a.id = aa.amenity_id").
			Where("LOWER(a.name) LIKE ?", "%"+strings.ToLower(amenity)+"%")
	}

	// 🔗 Join with house rules if filtering
	if houseRule != "" {
		query = query.Joins("JOIN apartment_house_rules ahr ON ahr.apartment_id = apartments.id").
			Joins("JOIN house_rules hr ON hr.id = ahr.house_rule_id").
			Where("LOWER(hr.rule) LIKE ?", "%"+strings.ToLower(houseRule)+"%")
	}

	// 🔢 Count total
	var total int64
	query.Count(&total)

	// ⛔ Reset pagination if out of bounds
	if offset >= int(total) {
		page = 1
		offset = 0
	}

	// 🧾 Final fetch
	type ApartmentInfo struct {
		model.Apartment
		LandlordName string `json:"landlord_name"`
	}

	var results []ApartmentInfo
	err := query.Offset(offset).Limit(limit).Scan(&results).Error
	if err != nil {
		log.Println("[ERROR] Failed to fetch apartments:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch apartment data",
			Data:    nil,
		})
	}

	// ✅ Response
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Apartments fetched successfully",
		Data: fiber.Map{
			"limit":       limit,
			"page":        page,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
			"apartments":  results,
		},
	})
}
