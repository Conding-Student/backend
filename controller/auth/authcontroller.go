package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/model"

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

	if firebaseAuthClient == nil {
		log.Println("[ERROR] Firebase Auth Client not initialized")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server configuration error",
		})
	}

	token, err := firebaseAuthClient.VerifyIDToken(context.Background(), requestData.IDToken)
	if err != nil {
		log.Printf("[ERROR] Firebase token verification failed: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token",
		})
	}
	uid := token.UID
	email, _ := token.Claims["email"].(string)
	fullName, _ := token.Claims["name"].(string)
	photoUrl, _ := token.Claims["picture"].(string)
	provider := ""
	if firebaseMap, ok := token.Claims["firebase"].(map[string]interface{}); ok {
		if signInProvider, ok := firebaseMap["sign_in_provider"].(string); ok {
			provider = signInProvider
		}
	}

	if provider == "facebook.com" {
		userRecord, err := firebaseAuthClient.GetUser(context.Background(), uid)
		if err == nil && userRecord.PhotoURL != "" {
			photoUrl = userRecord.PhotoURL
		}
	}

	// Check for deleted accounts
	var existingUser model.User
	if err := middleware.DBConn.
		Unscoped().
		Where("uid = ?", uid).
		First(&existingUser).Error; err == nil {
		if existingUser.AccountStatus == "Deleted" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Account deleted",
				"message": "This account has been permanently deleted",
			})
		}
	}

	// Save or update user with extracted data
	role, err := saveOrUpdateUserWithDetails(uid, email, fullName, photoUrl, provider)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User account processing failed",
		})
	}

	newJWT, err := middleware.GenerateJWT(uid, email, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Authentication successful",
		"uid":          uid,
		"email":        email,
		"fullname":     fullName,
		"photo_url":    photoUrl,
		"role":         role,
		"access_token": newJWT,
	})
}

// Verify Firebase ID Token with account status check
func VerifyFirebaseTokenAdmin(c *fiber.Ctx) error {
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
	var existingAdmin model.Admins
	if err := middleware.DBConn.
		Unscoped(). // Include soft-deleted records
		Where("uid = ?", uid).
		First(&existingAdmin).Error; err == nil {
		// Handle soft-deleted admin case if needed
	}

	// 💾 Create/update admin account
	role, err := saveOrUpdateAdmin(uid, email)
	if err != nil {
		log.Printf("[ERROR] Admin save failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Admin account processing failed",
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

	log.Printf("✅ Successful admin authentication: %s (%s)", email, uid)
	return c.JSON(fiber.Map{
		"message":      "Authentication successful",
		"uid":          uid,
		"email":        email,
		"role":         role,
		"access_token": newJWT,
	})
}

func saveOrUpdateAdmin(uid, email string) (string, error) {
	var admin model.Admins
	err := middleware.DBConn.Where("uid = ?", uid).First(&admin).Error

	if err == nil {
		// Existing admin - update email if changed
		if admin.Email != email {
			admin.Email = email
			if err := middleware.DBConn.Save(&admin).Error; err != nil {
				return "", fmt.Errorf("email update failed: %v", err)
			}
		}
		return "Admin", nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new admin
		newAdmin := model.Admins{
			Uid:       uid,
			Email:     email,
			Password:  "", // Optional: Set to default/blank or let frontend handle
			CreatedAt: time.Now(),
		}

		if err := middleware.DBConn.Create(&newAdmin).Error; err != nil {
			return "", fmt.Errorf("admin creation failed: %v", err)
		}
		return "Admin", nil
	}

	return "", fmt.Errorf("database error: %v", err)
}

func saveOrUpdateUserWithDetails(uid, email, fullname, photoUrl, provider string) (string, error) {
	var user model.User
	err := middleware.DBConn.Where("uid = ?", uid).First(&user).Error

	if err == nil {
		// Update fields if changed
		changed := false
		if user.Email != email {
			user.Email = email
			changed = true
		}
		if fullname != "" && user.Fullname != fullname {
			user.Fullname = fullname
			changed = true
		}
		if photoUrl != "" && user.PhotoURL != photoUrl {
			user.PhotoURL = photoUrl
			changed = true
		}
		if user.Provider != provider {
			user.Provider = provider
			changed = true
		}
		if changed {
			user.UpdatedAt = time.Now()
			if err := middleware.DBConn.Save(&user).Error; err != nil {
				return "", fmt.Errorf("user update failed: %v", err)
			}
		}
		return user.UserType, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newUser := model.User{
			Uid:           uid,
			Email:         email,
			Fullname:      fullname,
			PhotoURL:      photoUrl,
			AccountStatus: "Unverified",
			UserType:      "Tenant", // default role
			Provider:      provider,
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
