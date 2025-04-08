package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"intern_template_v1/model/response"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetFilteredUserDetails(c *fiber.Ctx) error {
	// 🔍 Filters
	userType := c.Query("user_type", "")
	accountStatus := c.Query("account_status", "")
	name := c.Query("name", "")

	// 📄 Pagination
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 📌 Base query with selected fields only
	query := middleware.DBConn.Table("users").
		Select("uid, email, phone_number, fullname, address, valid_id, account_status, user_type").
		Where("account_status != ?", "deleted")

	// ✅ Apply filters
	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if accountStatus != "" {
		query = query.Where("account_status = ?", accountStatus)
	}
	if name != "" {
		query = query.Where("LOWER(fullname) LIKE ?", "%"+strings.ToLower(name)+"%")
	}

	// 🔢 Count total filtered rows
	var total int64
	query.Count(&total)

	// ✅ Prevent out-of-bound pages
	if offset >= int(total) {
		page = 1
		offset = 0
	}

	// 🧾 Fetch paginated results
	var users []struct {
		UID           string `json:"uid"`
		Email         string `json:"email"`
		PhoneNumber   string `json:"phone_number"`
		FullName      string `json:"fullname" gorm:"column:fullname"`
		Address       string `json:"address"`
		ValidID       string `json:"valid_id"`
		AccountStatus string `json:"account_status"`
		UserType      string `json:"user_type"`
	}

	err = query.Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		log.Println("[ERROR] Failed to fetch users:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to fetch user data",
			Data:    nil,
		})
	}

	// 📦 Paginated Response
	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "Paginated user list retrieved successfully",
		Data: fiber.Map{
			"limit":       limit,
			"page":        page,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
			"users":       users,
		},
	})
}

// ✅ Function to update a user's displayed fields
func UpdateUserDetails(c *fiber.Ctx) error {
	type UpdatePayload struct {
		UID           string `json:"uid"` // Required
		Email         string `json:"email"`
		PhoneNumber   string `json:"phone_number"`
		FullName      string `json:"fullname"`
		Address       string `json:"address"`
		ValidID       string `json:"valid_id"`
		AccountStatus string `json:"account_status"`
		UserType      string `json:"user_type"`
	}

	var payload UpdatePayload
	if err := c.BodyParser(&payload); err != nil {
		log.Println("[ERROR] Invalid update payload:", err)
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Invalid request body",
			Data:    nil,
		})
	}

	// Fetch the user to update
	var user model.User
	if err := middleware.DBConn.Where("uid = ?", payload.UID).First(&user).Error; err != nil {
		log.Println("[ERROR] User not found:", err)
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
			Data:    nil,
		})
	}

	// Update only the fields provided
	if payload.Email != "" {
		user.Email = payload.Email
	}
	if payload.PhoneNumber != "" {
		user.PhoneNumber = payload.PhoneNumber
	}
	if payload.FullName != "" {
		user.Fullname = payload.FullName
	}
	if payload.Address != "" {
		user.Address = payload.Address
	}
	if payload.ValidID != "" {
		user.ValidID = payload.ValidID
	}
	if payload.AccountStatus != "" {
		user.AccountStatus = payload.AccountStatus
	}
	if payload.UserType != "" {
		user.UserType = payload.UserType
	}

	// Save the updated user
	if err := middleware.DBConn.Save(&user).Error; err != nil {
		log.Println("[ERROR] Failed to update user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update user details",
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User details updated successfully",
		Data: fiber.Map{
			"uid":            user.Uid,
			"email":          user.Email,
			"phone_number":   user.PhoneNumber,
			"fullname":       user.Fullname,
			"address":        user.Address,
			"valid_id":       user.ValidID,
			"account_status": user.AccountStatus,
			"user_type":      user.UserType,
		},
	})
}

// ✅ Function to soft-delete a user by setting account_status to 'deleted'
func SoftDeleteUser(c *fiber.Ctx) error {
	uid := c.Params("uid") // Get the UID from the URL path parameter

	if uid == "" {
		log.Println("[ERROR] UID parameter missing")
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "UID is required for deletion",
			Data:    nil,
		})
	}

	// Fetch the user record by UID
	var user model.User
	if err := middleware.DBConn.Where("uid = ?", uid).First(&user).Error; err != nil {
		log.Println("[ERROR] User not found:", err)
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
			Data:    nil,
		})
	}

	// Soft delete by updating account_status
	user.AccountStatus = "deleted"

	if err := middleware.DBConn.Save(&user).Error; err != nil {
		log.Println("[ERROR] Failed to soft-delete user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete user",
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User soft-deleted successfully",
		Data: fiber.Map{
			"uid":            user.Uid,
			"account_status": user.AccountStatus,
		},
	})
}

//

// ApartmentDetail is a struct that holds all fields from the apartments table,
// as well as related aggregated data.
type ApartmentDetail struct {
	ID           uint      `json:"id"`
	Uid          string    `json:"uid"`           // Landlord UID from apartments table
	PropertyName string    `json:"property_name"` // Example column from apartments table
	Address      string    `json:"address"`       // Example column from apartments table
	PropertyType string    `json:"property_type"` // Example column from apartments table
	RentPrice    float64   `json:"rent_price"`    // Example column from apartments table
	LocationLink string    `json:"location_link"` // Example column from apartments table
	Landmarks    string    `json:"landmarks"`     // Example column from apartments table
	Status       string    `json:"status"`        // Example column from apartments table
	CreatedAt    time.Time `json:"created_at"`    // Example column from apartments table

	LandlordName  string `json:"landlord_name"`  // From users.fullname
	LandlordEmail string `json:"landlord_email"` // From users.email

	Images     string `json:"images"`      // Aggregated apartment images
	Amenities  string `json:"amenities"`   // Aggregated amenities names
	HouseRules string `json:"house_rules"` // Aggregated house rules
}

// GetApartmentDetails handles GET requests and returns all apartment details
// along with related landlord info, images, amenities, and house rules.
func GetApartmentDetails(c *fiber.Ctx) error {
	var apartments []ApartmentDetail

	err := middleware.DBConn.Table("apartments a").
		Select(`
			a.*, 
			u.fullname AS landlord_name, 
			u.email AS landlord_email, 
			STRING_AGG(DISTINCT ai.image_url, ', ') AS images, 
			STRING_AGG(DISTINCT am.name, ', ') AS amenities, 
			STRING_AGG(DISTINCT hr.rule, ', ') AS house_rules`).
		Joins("LEFT JOIN users u ON a.uid = u.uid").
		Joins("LEFT JOIN apartment_images ai ON ai.apartment_id = a.id").
		Joins("LEFT JOIN apartment_amenities aa ON aa.apartment_id = a.id").
		Joins("LEFT JOIN amenities am ON am.id = aa.amenity_id").
		Joins("LEFT JOIN apartment_house_rules ahr ON ahr.apartment_id = a.id").
		Joins("LEFT JOIN house_rules hr ON hr.id = ahr.house_rule_id").
		Group("a.id, u.uid, u.fullname, u.email").
		Order("a.created_at DESC").
		Find(&apartments).Error

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(apartments)
}
