package routes

import (
	controller "intern_template_v1/controller"
	"intern_template_v1/middleware"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	// SAMPLE ENDPOINT
	app.Get("/user/profile", middleware.AuthMiddleware, controller.GetUserProfile)

	// app.Get("/user/email", middleware.AuthMiddleware, controller.GetUserEmail)
	//app.Post("/property/add", middleware.AuthMiddleware, controller.AddProperty)
}
