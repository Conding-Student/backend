package middleware

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Secret key for signing the JWT token
var jwtSecret = []byte("8860a0e3f642c8a8ef7ba5d3408f19ee683d404e3fb9bee56bf5d7699de3ea142d49cda5e7799055f1a9878c4e2e94937b0359cddee2fbe11c77609c1799dd6c4d6f9c3f4d3d796b36f4c92a36d2ba4272ddd88974b5c5cce169cd85b716b7b73d6d7ca61d29dafc027d092ec2633643e89f203d7e3e1b3a3012b649512ff953190eee2346a5eb7d4a756bfb5ec9614f74c79d85468f933fd1934fd4d0da883f653dc9f6aaeb6056b31782844b8e547253507d05c2f5649e77608d9ee95bbcb7cc99a6b167c5b205be1efd6989af9965666f63fa5f09c01f5d9bda0acd27b5e9bcaa080e0c64536b643430d4d6907d3bca35c1e6cbc2e602214659b97d841c13") // 🔴 Change this to a strong secret key

// GenerateJWT creates a JWT token for a user
func GenerateJWT(userID uint, email, userType string) (string, error) {
	// Define token expiration time (e.g., 24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Define token claims
	claims := jwt.MapClaims{
		"id":        userID,
		"email":     email,
		"user_type": userType,
		"exp":       expirationTime.Unix(), // Expiration time (Unix format)
	}

	// Create the token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token using the secret key
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
