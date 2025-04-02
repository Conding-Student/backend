package middleware

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtSecret = []byte("c4077a98dc8cfc12579593500d723270ef97f8dc28005cf8cad6d45e52ef346ffbf6c457d9fb1efff69c9102d59c1a37c518daa4a9fa8ebc6ff720b1747c31b9cdaa15c470a7a0218b42f9a2fecc283da64f372be74ca519ef37ba3f2461491f330f2e093ebacd448033d08a2ccbe616d5113a3e2a839aae7004bb786a68030ff6c0bba8302ccac771b2ba7ca890a73f90260d4b6f39e5356a6da59549a0674bee875711165d6c8e9efe08d96b9851ab1741e9c182eb7f5ea5a2dab6070c7e0c3b552ecff372d0953f34c3b9913f5284c47d8c445d3aad72ea1816cb80b5ebf927235e17d6f2e63188c0af9006f2df2eff8b0483bc8e17ecbd4d4eb96c91bc86")

// Struct for JWT Claims
type Claims struct {
	UID   string `json:"uid"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

// Generate JWT with Role
func GenerateJWT(uid, email, role string) (string, error) {
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
	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// Print the generated token
	fmt.Println("Generated JWT Token:", signedToken)

	return signedToken, nil
}
