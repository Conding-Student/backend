package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"intern_template_v1/middleware"
	"intern_template_v1/model"

	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

var firebaseAuthClient *auth.Client

// Initialize Firebase Auth Client
func InitFirebase(client *auth.Client) {
	firebaseAuthClient = client
}

var jwtSecret = []byte("your-secret-key")

// Struct for JWT Claims
type Claims struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
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

	// Verify Firebase ID Token
	token, err := firebaseAuthClient.VerifyIDToken(context.Background(), requestData.IDToken)
	if err != nil {
		log.Printf("Failed to verify ID token: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid Firebase token",
		})
	}

	// Token verified, respond with custom JWT token
	fmt.Printf("✅ Firebase Token Verified! UID: %s\n", token.UID)

	uid := token.UID
	email := token.Claims["email"].(string)

	// Check user in PostgreSQL and get role
	role := getUserRoleFromDB(uid, email)

	newJWT, err := generateJWT(uid, email, role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate JWT"})
	}

	// Here, generate your own JWT token (if needed) and send it back
	return c.JSON(fiber.Map{
		"message":      "Firebase token verified successfully",
		"uid":          token.UID,
		"access_token": newJWT,
	})
}

// Generate JWT with Role
func generateJWT(uid, email, role string) (string, error) {
	claims := Claims{
		UID:   uid,
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(), // 1-day expiry
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

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
                log.Println("Error inserting new user:", err)
                return "Tenant"
            }
            fmt.Println("🆕 New user added:", email, "Role:", role)
        } else {
            log.Println("Error fetching user role:", result.Error)
            return "Tenant"
        }
    } else {
        role = user.UserType
    }

    return role
}
