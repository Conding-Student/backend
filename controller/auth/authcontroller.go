package controller

import (
	"context"
	"errors"
	"fmt"
	"intern_template_v1/middleware"
	"intern_template_v1/model"
	"log"
	"net/http"
	"time"

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

// Verify Firebase ID Token with account status check
func VerifyFirebaseToken(c *fiber.Ctx) error {
	var requestData struct {
		IDToken string `json:"id_token"`
	}

	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Firebase Auth Client check
	if firebaseAuthClient == nil {
		log.Println("[ERROR] Firebase Auth Client not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server configuration error",
		})
	}

	// Verify Firebase ID Token
	token, err := firebaseAuthClient.VerifyIDToken(context.Background(), requestData.IDToken)
	if err != nil {
		log.Printf("[ERROR] Firebase token verification failed: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token",
		})
	}

	// Extract UID and email
	uid := token.UID
	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		log.Println("[ERROR] Invalid email claim in token")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email in token",
		})
	}

	// 🛑 Check for deleted account first
	var existingUser model.User
	if err := middleware.DBConn.
		Unscoped(). // Include soft-deleted records
		Where("uid = ?", uid).
		First(&existingUser).Error; err == nil {

		if existingUser.AccountStatus == "Deleted" {
			log.Printf("[WARNING] Deleted account attempt: %s", uid)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Account deleted",
				"message": "This account has been permanently deleted",
			})
		}
	}

	// 💾 Create/update user account
	role, err := saveOrUpdateUser(uid, email)
	if err != nil {
		log.Printf("[ERROR] User save failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User account processing failed",
		})
	}

	// 🔑 Generate JWT
	newJWT, err := middleware.GenerateJWT(uid, email, role)
	if err != nil {
		log.Printf("[ERROR] JWT generation failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	log.Printf("✅ Successful authentication: %s (%s)", email, uid)
	return c.JSON(fiber.Map{
		"message":      "Authentication successful",
		"uid":          uid,
		"email":        email,
		"role":         role,
		"access_token": newJWT,
	})
}

// Enhanced saveOrUpdateUser function
func saveOrUpdateUser(uid, email string) (string, error) {
	var user model.User
	err := middleware.DBConn.Where("uid = ?", uid).First(&user).Error

	if err == nil {
		// Existing user - update email if changed
		if user.Email != email {
			user.Email = email
			if err := middleware.DBConn.Save(&user).Error; err != nil {
				return "", fmt.Errorf("email update failed: %v", err)
			}
		}
		return user.UserType, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new user
		newUser := model.User{
			Uid:           uid,
			Email:         email,
			AccountStatus: "Unverified",
			UserType:      "Tenant", // Default role
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := middleware.DBConn.Create(&newUser).Error; err != nil {
			return "", fmt.Errorf("user creation failed: %v", err)
		}
		return newUser.UserType, nil
	}

	return "", fmt.Errorf("database error: %v", err)
}

