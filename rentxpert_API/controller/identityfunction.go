package controller

// import (
// 	"errors"
// 	"intern_template_v1/middleware"
// 	"intern_template_v1/model"
// 	"intern_template_v1/model/response"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/gorm"
// )

// type Book struct {
// 	Id       int    `json:"id"`
// 	Fullname string `json:"name"`
// 	Address  string `json:"author"`
// }

// func (Book) TableName() string {
// 	return "book"
// }

// func CreateBook(c *fiber.Ctx) error {
// 	// Parse the incoming JSON request into a Book struct
// 	bookCreateResponse := new(model.Book)
// 	if err := c.BodyParser(bookCreateResponse); err != nil {
// 		// Return a 400 Bad Request error if the request body cannot be parsed
// 		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
// 			"message": err.Error(),
// 		})
// 	}

// 	// Declare a variable to store the last inserted book
// 	var existingbook model.Book

// 	// Query the database to get the last inserted book, ordering by ID in descending order
// 	result := middleware.DBConn.Debug().Table("book").Order("id DESC").First(&existingbook)

// 	// Check if there was an error (excluding the case where no records were found)
// 	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
// 		// Return a 500 Internal Server Error response if database query fails
// 		return c.JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to create Identity",
// 			Data:    result.Error,
// 		})
// 	}

// 	// If a record exists, set the new book ID to the last ID + 1
// 	if result.RowsAffected > 0 {
// 		bookCreateResponse.Id = existingbook.Id + 1
// 	}

// 	// Insert the new book into the database
// 	if err := middleware.DBConn.Debug().Table("book").Create(bookCreateResponse).Error; err != nil {
// 		// Return a 500 Internal Server Error if the insert operation fails
// 		return c.JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to create an Identity",
// 			Data:    err,
// 		})
// 	}

// 	// Return a 200 response indicating the book was successfully created
// 	return c.JSON(response.ResponseModel{
// 		RetCode: "200",
// 		Message: "Identity successfully added",
// 		Data:    bookCreateResponse,
// 	})
// }

// func GetAllBooks(c *fiber.Ctx) error {
// 	var books []model.Book

// 	// Fetch all books from the "book" table
// 	result := middleware.DBConn.Debug().Table("book").Find(&books)
// 	if result.Error != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to retrieve books",
// 			Data:    result.Error.Error(),
// 		})
// 	}

// 	// If no books are found
// 	if len(books) == 0 {
// 		return c.Status(fiber.StatusNotFound).JSON(response.ResponseModel{
// 			RetCode: "404",
// 			Message: "No books found",
// 			Data:    nil,
// 		})
// 	}

// 	// Successfully retrieved books
// 	return c.Status(fiber.StatusOK).JSON(response.ResponseModel{
// 		RetCode: "200",
// 		Message: "Books retrieved successfully",
// 		Data:    books,
// 	})
// }

// // get single books
// func Getbook(c *fiber.Ctx) error {
// 	bookID := c.Params("id")

// 	var book model.Book
// 	if fetchErr := middleware.DBConn.Debug().Table("book").Where("id = ?", bookID).First(&book).Error; fetchErr != nil {
// 		return c.JSON(response.ResponseModel{
// 			RetCode: "404",
// 			Message: "books not found",
// 			Data:    fetchErr,
// 		})
// 	}
// 	return c.JSON(response.ResponseModel{
// 		RetCode: "200",
// 		Message: "Book found",
// 		Data:    book,
// 	})
// }

// // // UpdateBook updates an existing book
// // func UpdateBook(c *fiber.Ctx) error {
// // 	// Get ID from URL parameter
// // 	bookID := c.Params("id")

// // 	var existingbook model.Book
// // 	if fetchErr := middleware.DBConn.Debug().Table("book").Where("id = ?", &bookID).First(&existingbook).Error; fetchErr != nil {
// // 		return c.JSON(response.ResponseModel{
// // 			RetCode: "404",
// // 			Message: "books not found",
// // 			Data:    fetchErr,
// // 		})
// // 	}

// // 	var updatedBook model.Book
// // 	if err := c.BodyParser(updatedBook); err != nil {
// // 		return err
// // 	}

// // 	// update book in database
// // 	if updateErr := middleware.DBConn.Debug().Table("book").Where("id = ?", &bookID).Updates(&updatedBook).Error; updateErr != nil {
// // 		return c.JSON(response.ResponseModel{
// // 			RetCode: "404",
// // 			Message: "books id can't be found",
// // 			Data:    updateErr,
// // 		})
// // 	}
// // 	return c.JSON(response.ResponseModel{
// // 		RetCode: "200",
// // 		Message: "Book found",
// // 		Data:    updatedBook,
// // 	})

// // }

// // Update Book
// func UpdateBook(c *fiber.Ctx) error {
// 	bookID := c.Params("id") // Get the book ID from the request parameters

// 	var existingBook model.Book
// 	if fetchErr := middleware.DBConn.Debug().Table("book").Where("id=?", bookID).First(&existingBook).Error; fetchErr != nil {
// 		return c.JSON(response.ResponseModel{
// 			RetCode: "404",
// 			Message: "Book Not Found",
// 			Data:    fetchErr,
// 		})
// 	}
// 	UpdateBook := new(model.Book)
// 	if err := c.BodyParser(UpdateBook); err != nil {
// 		return err
// 	}

// 	if updateErr := middleware.DBConn.Debug().Table("book").Where("id=?", bookID).Updates(UpdateBook).Error; updateErr != nil {
// 		return c.JSON(response.ResponseModel{
// 			RetCode: "500",
// 			Message: "Failed to update book",
// 			Data:    updateErr,
// 		})
// 	}
// 	return c.JSON(response.ResponseModel{
// 		RetCode: "200",
// 		Message: "Book Successfully Updated",
// 		Data:    UpdateBook,
// 	})
// }
