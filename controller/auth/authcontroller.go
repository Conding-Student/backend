package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var firebaseAuthClient *auth.Client

// Initialize Firebase Auth Client
func InitFirebase(client *auth.Client) {
	firebaseAuthClient = client
}

// Struct for JWT Claims
type Claims struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// Verify Firebase ID Token
func VerifyFirebaseToken(c *fiber.Ctx) error {
	var requestData struct {
		IDToken string `json:"id_token"`
	}

	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Ensure Firebase Auth Client is initialized
	if firebaseAuthClient == nil {
		log.Println("[ERROR] Firebase Auth Client is not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server configuration error",
		})
	}

	// Verify Firebase ID Token
	token, err := firebaseAuthClient.VerifyIDToken(context.Background(), requestData.IDToken)
	if err != nil {
		log.Printf("[ERROR] Failed to verify Firebase ID token: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token",
		})
	}

	// Log the token claims
	log.Printf("✅ Firebase Token Verified! UID: %s, Claims: %+v\n", token.UID, token.Claims)

	// Extract email from claims safely
	emailClaim, emailExists := token.Claims["email"]
	if !emailExists {
		log.Println("[ERROR] Email claim not found in Firebase token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token - missing email",
		})
	}

	email, ok := emailClaim.(string)
	if !ok {
		log.Println("[ERROR] Email claim is not a valid string")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token - email format incorrect",
		})
	}

	// Get user role from DB or create a new user
	role := getUserRoleFromDB(token.UID, email)

	// Generate JWT
	newJWT, err := middleware.GenerateJWT(token.UID, email, role)
	if err != nil {
		log.Printf("[ERROR] Failed to generate JWT: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	// Successful response
	return c.JSON(fiber.Map{
		"message":      "Firebase token verified successfully",
		"uid":          token.UID,
		"access_token": newJWT,
	})
}

// Get user role from database or create a new user
func getUserRoleFromDB(uid, email string) string {
	var user model.User
	var role string

	result := middleware.DBConn.Where("uid = ?", uid).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Assign default role
			role = "Tenant"

			// Insert new user (without manually setting ID)
			newUser := model.User{
				Uid:      uid,
				Email:    email,
				UserType: role,
			}

			// Save new user in the database
			err := middleware.DBConn.Create(&newUser).Error
			if err != nil {
				log.Println("[ERROR] Failed to insert new user:", err)
				return "Tenant"
			}
			fmt.Println("🆕 New user added:", email, "Role:", role)
		} else {
			log.Println("[ERROR] Database error while fetching user role:", result.Error)
			return "Tenant"
		}
	} else {
		role = user.UserType
	}

	return role
}
