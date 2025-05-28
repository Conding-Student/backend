package config

import (
	"context"
	"encoding/base64"

	//"fmt"
	"io/ioutil"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

func InitializeFirebase() *firebase.App {
	// Get base64-encoded JSON from environment variable
	encodedCreds := os.Getenv("FIREBASE_CREDENTIALS_B64")
	if encodedCreds == "" {
		log.Fatalf("🔥 FIREBASE_CREDENTIALS_B64 environment variable not set")
	}

	// Decode the base64 string
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		log.Fatalf("🔥 Failed to decode Firebase credentials: %v", err)
	}

	// Write to a temporary file (Firebase SDK needs a file path)
	tmpFile := "firebase_credentials.json"
	err = ioutil.WriteFile(tmpFile, decodedCreds, 0600)
	if err != nil {
		log.Fatalf("🔥 Failed to write Firebase credentials: %v", err)
	}

	// Initialize Firebase App
	opt := option.WithCredentialsFile(tmpFile)
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
