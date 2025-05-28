package config

import (
	"context"
	"fmt"
	"log"
	"os"

	//"time"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

// Route handler to resend verification email

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
	print(email)
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

func InitializeFirebase() *firebase.App {
	credJSON := os.Getenv("FIREBASE_CREDENTIALS")
	if credJSON == "" {
		log.Fatalf("FIREBASE_CREDENTIALS env var is not set!")
	}
	opt := option.WithCredentialsJSON([]byte(credJSON))

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("🔥 Error initializing Firebase App: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("🔥 Error getting Auth client: %v", err)
	}

	FirebaseAuth = authClient
	log.Println("✅ Firebase Auth initialized successfully")
	return app
}
