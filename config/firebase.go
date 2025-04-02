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
	opt := option.WithCredentialsFile("config/rentxpert-a987d-firebase-adminsdk-fbsvc-c025185942.json")

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v", err)
	}

	return app
}
