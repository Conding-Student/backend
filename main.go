package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"intern_template_v1/config"
	"intern_template_v1/controller"
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

	// Initialize PayMongo service
	service := &controller.PayMongoService{
		DB:        middleware.DBConn, // Use the global DB connection
		PublicKey: "pk_test_r4grVzaAZPPhvjakJZX7Fjpv",
		SecretKey: "sk_test_87FgkvAzg7CcPaRxemAL8Qhd",
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
	app.Get("/success", func(c *fiber.Ctx) error {
		log.Printf("Received PayMongo success redirect")
		return c.JSON(fiber.Map{"status": "success", "message": "Payment authorized"})
	})
	app.Get("/failed", func(c *fiber.Ctx) error {
		log.Printf("Received PayMongo failed redirect")
		return c.JSON(fiber.Map{"status": "failed", "message": "Payment failed"})
	})


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
	log.Fatal(app.Listen("0.0.0.0:8080")) // Allows external devices to connect
}
