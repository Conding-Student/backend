package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

func InitializeFirebase() *firebase.App {
	opt := option.WithCredentialsFile("config/rentxpert-a987d-firebase-adminsdk-fbsvc-500f11a577.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("🔥 Error initializing Firebase App: %v", err)
	}

	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("🔥 Error getting Auth client: %v", err)
	}

	FirebaseAuth = authClient // ← important!
	log.Println("✅ Firebase Auth initialized successfully")
	return app
}
