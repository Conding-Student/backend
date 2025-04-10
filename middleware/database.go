package middleware

import (
	"fmt"
	"intern_template_v1/model" // Updated models import
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBConn *gorm.DB

// ConnectDB initializes the database connection and runs migrations
func ConnectDB() bool {
	// Database Config (using environment variables)
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=%s",
		GetEnv("DB_HOST"), GetEnv("DB_PORT"), GetEnv("DB_NAME"),
		GetEnv("DB_UNME"), GetEnv("DB_PWRD"), GetEnv("DB_SSLM"),
		GetEnv("DB_TMEZ"))

	var err error
	DBConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
		return true
	}

	// ✅ Run AutoMigrate for all models
	err = DBConn.AutoMigrate(
		&model.User{},
		&model.Admins{},
		&model.Apartment{},
		&model.LandlordProfile{},
		&model.ApartmentImage{},
		&model.ApartmentVideo{},
		&model.Inquiry{},
		&model.Amenity{},
		&model.ApartmentAmenity{},
		&model.HouseRule{},
		&model.ApartmentHouseRule{},
		&model.Wishlist{},
	)

	if err != nil {
		log.Fatal("❌ Migration failed:", err)
		return true
	}

	log.Println("✅ Database connected and migrations successful.")
	return false
}
