package routes

import (
	controller "github.com/Conding-Student/backend/controller"
	"github.com/Conding-Student/backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	// SAMPLE ENDPOINT
	app.Get("/user/profile", middleware.AuthMiddleware, controller.GetUserProfile)
	app.Get("/users/fullname/:uid", middleware.AuthMiddleware, controller.GetFullnameByUID)
	app.Get("/user/:uid/role", middleware.AuthMiddleware, controller.GetUserRoleByUID)
	app.Get("/api/user/photo/:uid", controller.GetUserProfilePhotoByUID)

	// app.Get("/user/email", middleware.AuthMiddleware, controller.GetUserEmail)
	//app.Post("/property/add", middleware.AuthMiddleware, controller.AddProperty)
}
