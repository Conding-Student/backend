package controller

import (
	"context"

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
	if !ok || email == "" {
		log.Println("[ERROR] Email claim is not a valid string")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token - email format incorrect",
		})
	}

	// Store user in DB and get role
	role, ok := saveOrUpdateUser(token.UID, email)

	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store user profile",
		})
	}

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
		"email":        email,
		"role":         role,
		"access_token": newJWT,
	})
}

// Save or update user in database
func saveOrUpdateUser(uid, email string) (string, bool) {
	var user model.User
	defaultRole := "Tenant"

	// Check if user exists in DB
	result := middleware.DBConn.Where("uid = ?", uid).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// // Fetch additional user details from Firebase
			firebaseUser, err := firebaseAuthClient.GetUser(context.Background(), uid)
			if err != nil {
				log.Println("[ERROR] Failed to fetch user details from Firebase:", err)
				return defaultRole, false
			}

			// Extract Firebase user info
			provider := "firebase"
			if len(firebaseUser.ProviderUserInfo) > 0 {
				provider = firebaseUser.ProviderUserInfo[0].ProviderID
			}

			fullname := firebaseUser.DisplayName
			photoURL := firebaseUser.PhotoURL

			log.Printf("🔥 Firebase User Info → Fullname: %s | Provider: %s | PhotoURL: %s", fullname, provider, photoURL)

			// Create new user record
			newUser := model.User{
				Uid:         uid,
				Email:       email,
				UserType:    role,
				PhoneNumber: firebaseUser.PhoneNumber,
				Provider:    provider,
				PhotoURL:    firebaseUser.PhotoURL,
				Fullname:    firebaseUser.DisplayName,
				Birthday:    "", // Requires frontend to send the birthday separately
			}

			// Insert into DB
			if err := middleware.DBConn.Create(&newUser).Error; err != nil {
				log.Printf("[ERROR] Failed to insert new user (UID: %s): %v", uid, err)
				return defaultRole, false
			}

			log.Println("🆕 New user successfully added to the database.")
			return defaultRole, true

		} else {
			log.Printf("[ERROR] Failed to query user in DB: %v", result.Error)
			return defaultRole, false
		}
	}

	// User exists, return existing role
	return user.UserType, true
}
