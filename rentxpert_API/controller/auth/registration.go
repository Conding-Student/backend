package controller

import (
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"log"
	"net/http"
	"regexp"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// ResponseModel struct for API responses
type ResponseModel struct {
	RetCode string      `json:"ret_code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Regex Validators
var (
	emailRegex         = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex         = regexp.MustCompile(`^[0-9]{10,15}$`)
	middleInitialRegex = regexp.MustCompile(`^[A-Z]?$`)
)

// Validation Functions
func isValidEmail(email string) bool          { return emailRegex.MatchString(email) }
func isValidPhoneNumber(phone string) bool    { return phoneRegex.MatchString(phone) }
func isValidMiddleInitial(middle string) bool { return middleInitialRegex.MatchString(middle) }
func isValidPassword(password string) bool    { return utf8.RuneCountInString(password) >= 8 }

// Validate user input
func validateUserInput(input *model.User, landlordProfile *model.LandlordProfile) *ResponseModel {
	if input.Email == "" || !isValidEmail(input.Email) {
		return &ResponseModel{"400", "Invalid or missing email", nil}
	}
	if input.Password == "" || !isValidPassword(input.Password) {
		return &ResponseModel{"400", "Password must be at least 8 characters", nil}
	}
	if input.FirstName == "" {
		return &ResponseModel{"400", "First name is required", nil}
	}
	if !isValidMiddleInitial(input.MiddleInitial) {
		return &ResponseModel{"400", "Middle initial must be a single uppercase letter or blank", nil}
	}
	if input.LastName == "" {
		return &ResponseModel{"400", "Last name is required", nil}
	}
	if input.Age < 18 {
		return &ResponseModel{"400", "User must be at least 18 years old", nil}
	}
	if input.Address == "" {
		return &ResponseModel{"400", "Address is required", nil}
	}
	if input.ValidID == "" {
		return &ResponseModel{"400", "Valid ID is required", nil}
	}
	if input.PhoneNumber == "" || !isValidPhoneNumber(input.PhoneNumber) {
		return &ResponseModel{"400", "Invalid phone number format (10-15 digits only)", nil}
	}
	if input.UserType != "Tenant" && input.UserType != "Landlord" {
		return &ResponseModel{"400", "Invalid user type (must be 'Tenant' or 'Landlord')", nil}
	}
	if input.UserType == "Landlord" {
		if landlordProfile == nil || landlordProfile.BusinessName == "" || landlordProfile.BusinessPermit == "" {
			return &ResponseModel{"400", "Business name and permit are required for landlords", nil}
		}
	}
	return nil
}

// Common function for user registration
func registerUser(c *fiber.Ctx, userType string) error {
	var input model.User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(ResponseModel{"400", "Invalid request format", err.Error()})
	}
	input.UserType = userType

	var landlordProfile *model.LandlordProfile
	if userType == "Landlord" {
		landlordProfile = &model.LandlordProfile{}
		if err := c.BodyParser(landlordProfile); err != nil {
			return c.Status(http.StatusBadRequest).JSON(ResponseModel{"400", "Invalid landlord profile format", err.Error()})
		}
	}

	if errResp := validateUserInput(&input, landlordProfile); errResp != nil {
		return c.Status(http.StatusBadRequest).JSON(errResp)
	}

	var existingUser model.User
	result := middleware.DBConn.Where("email = ? OR phone_number = ?", input.Email, input.PhoneNumber).First(&existingUser)
	if result.Error == nil {
		return c.Status(http.StatusConflict).JSON(ResponseModel{"409", "Email or phone number already in use", nil})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("[ERROR] Password Hashing:", err)
		return c.Status(http.StatusInternalServerError).JSON(ResponseModel{"500", "Something went wrong, please try again", nil})
	}
	input.Password = string(hashedPassword)

	if err := middleware.DBConn.Create(&input).Error; err != nil {
		log.Println("[ERROR] Database Insert:", err)
		return c.Status(http.StatusInternalServerError).JSON(ResponseModel{"500", "Failed to create user", nil})
	}

	if userType == "Landlord" {
		landlordProfile.UserID = input.ID
		if err := middleware.DBConn.Create(landlordProfile).Error; err != nil {
			log.Println("[ERROR] Landlord Profile Insert:", err)
			return c.Status(http.StatusInternalServerError).JSON(ResponseModel{"500", "Failed to create landlord profile", nil})
		}
	}

	responseData := fiber.Map{
		"id":             input.ID,
		"email":          input.Email,
		"first_name":     input.FirstName,
		"middle_initial": input.MiddleInitial,
		"last_name":      input.LastName,
		"age":            input.Age,
		"address":        input.Address,
		"valid_id":       input.ValidID,
		"user_type":      input.UserType,
		"phone_number":   input.PhoneNumber,
		"created_at":     input.CreatedAt,
	}

	if userType == "Landlord" {
		responseData["business_name"] = landlordProfile.BusinessName
		responseData["business_permit"] = landlordProfile.BusinessPermit
	}

	return c.Status(http.StatusCreated).JSON(ResponseModel{"201", "Registration successful", responseData})
}

// Register Landlord Endpoint
func RegisterLandlord(c *fiber.Ctx) error {
	return registerUser(c, "Landlord")
}

// Register Tenant Endpoint
func RegisterTenant(c *fiber.Ctx) error {
	return registerUser(c, "Tenant")
}
