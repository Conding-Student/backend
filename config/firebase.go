package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

// Firebase credentials struct for environment variables
type firebaseConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func InitializeFirebase() *firebase.App {
	// Get credentials from environment variables
	creds := firebaseConfig{
		Type:                    os.Getenv("FIREBASE_TYPE"),
		ProjectID:               os.Getenv("FIREBASE_PROJECT_ID"),
		PrivateKeyID:            os.Getenv("FIREBASE_PRIVATE_KEY_ID"),
		PrivateKey:              strings.ReplaceAll(os.Getenv("FIREBASE_PRIVATE_KEY"), `\n`, "\n"),
		ClientEmail:             os.Getenv("FIREBASE_CLIENT_EMAIL"),
		ClientID:                os.Getenv("FIREBASE_CLIENT_ID"),
		AuthURI:                 os.Getenv("FIREBASE_AUTH_URI"),
		TokenURI:                os.Getenv("FIREBASE_TOKEN_URI"),
		AuthProviderX509CertURL: os.Getenv("FIREBASE_AUTH_PROVIDER_X509_CERT_URL"),
		ClientX509CertURL:       os.Getenv("FIREBASE_CLIENT_X509_CERT_URL"),
		UniverseDomain:          os.Getenv("FIREBASE_UNIVERSE_DOMAIN"),
	}

	// Convert to JSON for Firebase initialization
	jsonData, err := json.Marshal(creds)
	if err != nil {
		log.Fatalf("🔥 Error marshaling Firebase credentials: %v", err)
	}

	// Initialize Firebase with JSON credentials
	opt := option.WithCredentialsJSON(jsonData)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("🔥 Error initializing Firebase App: %v", err)
	}

	// Initialize Auth client
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("🔥 Error getting Auth client: %v", err)
	}

	FirebaseAuth = authClient
	log.Println("✅ Firebase initialized successfully")
	return app
}

// Rest of your functions remain unchanged
func UnverifyUserEmail(uid string) error {
	_, err := FirebaseAuth.UpdateUser(context.Background(), uid, (&auth.UserToUpdate{}).EmailVerified(false))
	if err != nil {
		return fmt.Errorf("Error unverifying user: %v", err)
	}
	fmt.Println("🔄 User marked as unverified:", uid)
	return nil
}

func ResendVerificationEmail(uid string) (string, error) {
	user, err := FirebaseAuth.GetUser(context.Background(), uid)
	if err != nil {
		return "", fmt.Errorf("Error fetching user: %v", err)
	}

	email := user.Email
	if email == "" {
		return "", fmt.Errorf("User does not have an email set.")
	}

	link, err := FirebaseAuth.EmailVerificationLink(context.Background(), email)
	if err != nil {
		return "", fmt.Errorf("Error generating verification link: %v", err)
	}

	fmt.Println("📩 Verification link generated:", link)
	return link, nil
}

func UnverifyAndResendHandler(c *fiber.Ctx) error {
	uid := c.Params("uid")
	if uid == "" {
		return c.Status(400).SendString("Missing UID parameter")
	}

	// Step 1: Unverify the email
	err := UnverifyUserEmail(uid)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// Step 2: Resend verification email
	link, err := ResendVerificationEmail(uid)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.SendString(fmt.Sprintf("✅ Email unverified and verification link sent: %s", link))
}
