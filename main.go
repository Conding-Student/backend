package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Conding-Student/backend/config"
	"github.com/Conding-Student/backend/controller"
	authController "github.com/Conding-Student/backend/controller/auth" // alias local auth package as authController
	"github.com/Conding-Student/backend/middleware"
	"github.com/Conding-Student/backend/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {

	if middleware.ConnectDB() {
		log.Fatal("🔥 Failed to connect to the database")
	}

	config.InitCloudinary()
	// Step 1: Initialize Firebase App
	firebaseApp := config.InitializeFirebase()
	fmt.Println("✅ Firebase Initialized:", firebaseApp)

	// Step 2: Initialize Firebase Auth with context
	firebaseAuthClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("🔥 Error initializing Firebase Auth: %v", err)
	}

	// Step 3: Set Firebase Auth in the auth controller
	authController.InitFirebase(firebaseAuthClient)

	// Load .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize PayMongo service
	service := &controller.PayMongoService{
		DB:        middleware.DBConn, // Use the global DB connection
		PublicKey: os.Getenv("PAYMONGO_PUBLIC_KEY"),
		SecretKey: os.Getenv("PAYMONGO_SECRET_KEY"),
	}

	// Step 4: Create Fiber App
	app := fiber.New(fiber.Config{
		AppName: middleware.GetEnv("PROJ_NAME"),
	})

	app.Post("/api/create-source", service.CreateSource)
	app.Post("/api/webhook", service.HandleWebhook)
	app.Get("/api/transaction/:source_id", service.GetTransaction)
	app.Get("/api/transactions", service.GetTransactions)

	// PayMongo redirect routes
app.Get("/success", service.HandleSuccessRedirect) // Add this
    app.Get("/failed", service.HandleFailedRedirect)   // Optional for failed redirects

	// CORS CONFIG
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	// Example usage of the JWT secret key
	jwtSecret := os.Getenv("JWT_SECRET")
	log.Println("Loaded JWT Secret:", jwtSecret) // 🔍 Debugging: REMOVE in production

	// Do not remove this endpoint
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(204) // No Content
	})

	// Step 5: Register Routes
	routes.AppRoutes(app)
	routes.UserRoutes(app)

	// LOGGER
	app.Use(logger.New())

	// Step 6: Start Server
	app.Listen("0.0.0.0:8080") // ✅ use this instead of "127.0.0.1:8080"

}
