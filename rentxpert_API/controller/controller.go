package controller

// import (
// 	"intern_template_v1/middleware"
// 	errorModel "intern_template_v1/model/error"
// 	"intern_template_v1/model/response"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// // // SampleController1 returns a JSON response
// // func SampleController1(c *fiber.Ctx) error {
// // 	res := fiber.Map{
// // 		"name":   "John Doe",
// // 		"age":    25,
// // 		"status": "Active",
// // 	}
// // 	return c.JSON(res)
// // }

// // // SampleController2 returns a simple string message
// // func SampleController2(c *fiber.Ctx) error {
// // 	return c.SendString("Hello, Golang World!!!")
// // }

// // Define User Model
// type Login struct {
// 	gorm.Model
// 	ID      int    `json:"id"`
// 	Name    string `json:"name"`
// 	Section string `json:"section"`
// }

// // User storage
// var users []Login

// // UserRegistration handles user registration
// func UserRegistration(c *fiber.Ctx) error {
// 	user := new(Login)

// 	// Parse the request body
// 	if err := c.BodyParser(user); err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(response.ResponseModel{
// 			RetCode: "400",
// 			Message: "Invalid Request!",
// 			Data: errorModel.ErrorModel{
// 				Message:   "Failed to parse request",
// 				IsSuccess: false,
// 				Error:     err.Error(),
// 			},
// 		})
// 	}

// 	// Check if ID is already registered
// 	for _, existingUser := range users {
// 		if existingUser.ID == user.ID {
// 			return c.Status(fiber.StatusConflict).JSON(response.ResponseModel{
// 				RetCode: "409",
// 				Message: "ID already registered",
// 				Data: errorModel.ErrorModel{
// 					Message:   "This ID is already in use",
// 					IsSuccess: false,
// 					Error:     nil,
// 				},
// 			})
// 		}
// 	}

// 	// Add user to the list
// 	//users = append(users, *user)
// 	result := middleware.DBConn.Table("logins").Create(&user)
// 	// middleware.DBConn.AutoMigrate(&user)

// 	if result.Error != nil {
// 		return c.JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to create user",
// 			Data: errorModel.ErrorModel{
// 				Message:   result.Error.Error(),
// 				IsSuccess: false,
// 				Error:     result.Error,
// 			},
// 		})
// 	}
// 	// Successful response
// 	return c.Status(fiber.StatusCreated).JSON(response.ResponseModel{
// 		RetCode: "201",
// 		Message: "Your account has been successfully created. Please login.",
// 		Data:    user,
// 	})
// }

// // ReadUser retrieves a user by ID
// func ReadUser(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var user Login

// 	result := middleware.DBConn.Table("logins").Where("id = ?", id).First(&user)
// 	if result.Error != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
// 			RetCode: "404",
// 			Message: "User not found",
// 			Data: errorModel.ErrorModel{
// 				Message:   result.Error.Error(),
// 				IsSuccess: false,
// 				Error:     result.Error,
// 			},
// 		})
// 	}

// 	return c.JSON(response.ResponseModel{
// 		RetCode: "200",
// 		Message: "User retrieved successfully",
// 		Data:    user,
// 	})
// }
