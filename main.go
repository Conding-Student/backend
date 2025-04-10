package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"intern_template_v1/config"
	authController "intern_template_v1/controller/auth" // alias local auth package as authController
	"intern_template_v1/middleware"
	"intern_template_v1/routes"

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


	// Step 4: Create Fiber App
	app := fiber.New(fiber.Config{
		AppName: middleware.GetEnv("PROJ_NAME"),
	})


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

	// CORS CONFIG
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// LOGGER
	app.Use(logger.New())

	// Step 6: Start Server
	log.Fatal(app.Listen("0.0.0.0:8080")) // Allows external devices to connect
}
