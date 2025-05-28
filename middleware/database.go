// package middleware

// import (
// 	"fmt"
// 	"log"

// 	"github.com/Conding-Student/backend/model" // Corrected models import

// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// var DBConn *gorm.DB

// // ConnectDB initializes the database connection and runs migrations
// func ConnectDB() bool {
// 	// Database Config (using environment variables)
// 	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=%s",
// 		GetEnv("DB_HOST"), GetEnv("DB_PORT"), GetEnv("DB_NAME"),
// 		GetEnv("DB_UNME"), GetEnv("DB_PWRD"), GetEnv("DB_SSLM"),
// 		GetEnv("DB_TMEZ"))

// 	var err error
// 	DBConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("❌ Database connection failed:", err)
// 		return true
// 	}

// 	// ✅ Run AutoMigrate for all models
// 	err = DBConn.AutoMigrate(
// 		&model.User{},
// 		&model.Admins{},
// 		&model.Apartment{},
// 		&model.LandlordProfile{},
// 		&model.ApartmentImage{},
// 		&model.ApartmentVideo{},
// 		&model.Inquiry{},
// 		&model.Amenity{},
// 		&model.ApartmentAmenity{},
// 		&model.HouseRule{},
// 		&model.ApartmentHouseRule{},
// 		&model.Wishlist{},
// 		&model.RecentlyViewed{},
// 		&model.RentalAgreement{},
// 		&model.Rating{},
// 		&model.Transaction{},
// 		// &model.AdminToken{},
// 	)

// 	// ✅ Create unique index (outside AutoMigrate)
// 	if err := DBConn.Exec(`
// 		CREATE UNIQUE INDEX IF NOT EXISTS idx_apartment_tenant
// 		ON rental_agreements (apartment_id, tenant_id)
// 	`).Error; err != nil {
// 		log.Fatal("❌ Failed to create unique index:", err)
// 		return true
// 	}

// 	if err != nil {
// 		log.Fatal("❌ Migration failed:", err)
// 		return true
// 	}

// 	log.Println("✅ Database connected and migrations successful.")
// 	return false
// }

package middleware

import (
	"fmt"
	"log"
	"os"
	"time"

	//"github.com/Conding-Student/backend/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DBConn *gorm.DB

func ConnectDB() bool {
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s TimeZone=%s",
		GetEnv("DB_HOST"), GetEnv("DB_PORT"), GetEnv("DB_NAME"),
		GetEnv("DB_UNME"), GetEnv("DB_PWRD"), GetEnv("DB_SSLM"),
		GetEnv("DB_TMEZ"))

	// Configure GORM to skip schema introspection and disable foreign keys
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		Logger:                                   logger.Default.LogMode(logger.Error),
	}

	var err error
	DBConn, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
		return true
	}

	// Configure connection pool
	sqlDB, _ := DBConn.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Conditionally run migrations based on environment variable
	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := runMigrations(); err != nil {
			log.Fatal("❌ Migration failed:", err)
			return true
		}
	}

	log.Println("✅ Database connected successfully.")
	return false
}

func runMigrations() error {
	log.Println("🏃 Running database migrations...")

	// Keep models list for reference (not used in production)
	// models := []interface{}{
	// 	&model.User{},
	// 	&model.Admins{},
	// 	&model.Apartment{},
	// 	&model.LandlordProfile{},
	// 	&model.ApartmentImage{},
	// 	&model.ApartmentVideo{},
	// 	&model.Inquiry{},
	// 	&model.Amenity{},
	// 	&model.ApartmentAmenity{},
	// 	&model.HouseRule{},
	// 	&model.ApartmentHouseRule{},
	// 	&model.Wishlist{},
	// 	&model.RecentlyViewed{},
	// 	&model.RentalAgreement{},
	// 	&model.Rating{},
	// 	&model.Transaction{},
	// 	// ... other models
	// }

	// Execute raw SQL for production migrations
	migrations := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_apartment_tenant 
		 ON rental_agreements (apartment_id, tenant_id)`,
	}

	for _, migration := range migrations {
		if err := DBConn.Exec(migration).Error; err != nil {
			return err
		}
	}

	log.Println("✅ Database migrations completed.")
	return nil
}
