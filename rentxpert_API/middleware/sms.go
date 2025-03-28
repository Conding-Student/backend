package middleware

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var AuthClient *auth.Client

func InitFirebase() {
	opt := option.WithCredentialsFile("/Users/fdsap-intern7-eadaza/Desktop/backends/backend/rentxpert_API/kikko-926e7-firebase-adminsdk-fbsvc-e7a1016fbe.json") // Replace with actual path
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth: %v", err)
	}

	AuthClient = client
	fmt.Println("✅ Firebase Auth initialized successfully!")
}
