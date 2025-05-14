package controller

import (
	//"fmt"
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

	// Fetch users with status "Unverified" OR "Pending" (across all tenants)
	query := middleware.DBConn.Table("users").
		Select("uid, email, phone_number, fullname, address, valid_id, account_status, user_type").
		Where("account_status IN ?", []string{"Unverified", "Pending", "Verified"}) // Explicit status filter

	// ✅ Apply filters
	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if accountStatus != "" {
		query = query.Where("LOWER(account_status) LIKE ?", "%"+strings.ToLower(accountStatus)+"%")
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

func GetFilteredUserDetailsadvance(c *fiber.Ctx) error {
	// 🔍 Basic Filters
	userType := c.Query("user_type", "")
	accountStatus := c.Query("account_status", "")
	name := c.Query("name", "")

	// 🔍 Advanced Search Filters
	searchField := c.Query("field", "")
	searchTerm := c.Query("search_term", "")

	// 📄 Pagination
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "10")

	// Validate Advanced Search Parameters
	if (searchField != "" && searchTerm == "") || (searchTerm != "" && searchField == "") {
		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
			RetCode: "400",
			Message: "Both 'field' and 'search_term' must be provided together",
			Data:    nil,
		})
	}

	// ✅ Set Tenant as default ONLY when no filters are active
	if userType == "" && searchField == "" && searchTerm == "" && name == "" && accountStatus == "" {
		userType = "Tenant" // Default to Tenant users
	}

	// Validate search field if provided
	if searchField != "" {
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
			return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
				RetCode: "400",
				Message: "Invalid search field",
				Data: fiber.Map{
					"allowed_fields": []string{
						"uid", "email", "phone_number",
						"fullname", "address",
						"account_status", "user_type",
					},
				},
			})
		}
	}

	// Pagination Processing
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Base Query
	query := middleware.DBConn.Table("users").
		Select("uid, email, phone_number, fullname, address, valid_id, account_status, user_type").
		Where("account_status IN ?", []string{"Unverified", "Pending", "Verified"})

	// ✅ Apply Basic Filters
	if userType != "" {
		query = query.Where("user_type = ?", userType)
	}
	if accountStatus != "" {
		query = query.Where("account_status = ?", accountStatus)
	}
	if name != "" {
		query = query.Where("fullname ILIKE ?", "%"+name+"%")
	}

	// ✅ Apply Advanced Search
	if searchField != "" && searchTerm != "" {
		query = query.Where(searchField+" ILIKE ?", "%"+searchTerm+"%")
	}

	// 🔢 Count Total Filtered Rows
	var total int64
	query.Count(&total)

	// ✅ Prevent Out-of-Bound Pages
	if offset >= int(total) {
		page = 1
		offset = 0
	}

	// 🧾 Fetch Paginated Results
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

// ✅ Function to soft-delete a user and related apartments and inquiries, setting expiration time
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

	// Start a transaction
	tx := middleware.DBConn.Begin()
	if tx.Error != nil {
		log.Println("[ERROR] Failed to start transaction:", tx.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to start transaction",
			Data:    nil,
		})
	}

	// Fetch the user record by UID within the transaction
	var user model.User
	if err := tx.Where("uid = ?", uid).First(&user).Error; err != nil {
		tx.Rollback()
		log.Println("[ERROR] User not found:", err)
		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
			RetCode: "404",
			Message: "User not found",
			Data:    nil,
		})
	}

	// Calculate expiration time (e.g., 90 days from now)
	expirationTime := time.Now().UTC().Add(90 * 24 * time.Hour) // Adjust duration as needed
	//expirationTime := time.Now().UTC().Add(1 * time.Minute) // Test with 2 minutes
	// Soft delete the user by updating account_status and expires_at
	user.AccountStatus = "Deleted"
	user.ExpiresAt = expirationTime
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		log.Println("[ERROR] Failed to soft-delete user:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to delete user",
			Data:    nil,
		})
	}

	// Prepare expiration time pointer for Apartment (which uses *time.Time)
	expirationTimePtr := &expirationTime

	// Soft delete all related apartments by setting status to "Deleted" and expires_at
	if err := tx.Model(&model.Apartment{}).
		Where("user_id = ?", uid).
		Updates(map[string]interface{}{
			"status":     "Deleted",
			"expires_at": expirationTimePtr,
		}).Error; err != nil {
		tx.Rollback()
		log.Println("[ERROR] Failed to update related apartments:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update related apartments",
			Data:    nil,
		})
	}

	// Update all related inquiries by setting status to "Rejected" and expires_at
	if err := tx.Model(&model.Inquiry{}).
		Where("tenant_uid = ? OR landlord_uid = ?", uid, uid). // Use proper column names
		Updates(map[string]interface{}{
			"status":     "Rejected",
			"expires_at": expirationTime,
		}).Error; err != nil {
		tx.Rollback()
		log.Println("[ERROR] Failed to update related inquiries:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to update related inquiries",
			Data:    nil,
		})
	}
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Println("[ERROR] Failed to commit transaction:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
			RetCode: "500",
			Message: "Failed to complete deletion",
			Data:    nil,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
		RetCode: "200",
		Message: "User and related data soft-deleted successfully. All records will expire on " + expirationTime.Format(time.RFC3339),
		Data: fiber.Map{
			"uid":            user.Uid,
			"account_status": user.AccountStatus,
			"expires_at":     user.ExpiresAt.Format(time.RFC3339),
		},
	})
}
