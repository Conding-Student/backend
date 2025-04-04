package model

import (
	"time"
)

// Landlord confirmation to delete "rejected" apartment
type DeleteApartmentRequest struct {
	Confirm bool `json:"confirm"` // Landlord must confirm deletion
}

type Admins struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	Uid           string    `gorm:"uniqueIndex"` // Unique user identifier
	Email         string    `gorm:"unique"`
	PhoneNumber   string    `json:"phone_number"`
	Password      string    `json:"password,omitempty"` // Optional for email sign-up
	Fullname      string    `json:"fullname"`
	Age           int       `json:"age"`
	Address       string    `json:"address"`
	ValidID       string    `json:"valid_id"`
	AccountStatus string    `gorm:"not null;default:'Pending'" json:"account_status"` // "Verified" / "Unverified"
	PhotoURL      string    `json:"photo_url"`
	UserType      string    `gorm:"not null" json:"user_type"` // "Landlord", "Tenant", "Admin"
	Birthday      string    `json:"birthday"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Admin model (Separate from User)
type Admin struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// Landlord Profile (Related to User via Uid)
type LandlordProfile struct {
	ID             uint   `gorm:"primaryKey"`
	Uid            string `gorm:"not null;uniqueIndex"`
	BusinessName   string `json:"business_name"`
	BusinessPermit string `json:"business_permit"`
}

// Apartment model
type Apartment struct {
	ID           uint      `gorm:"primaryKey"`
	Uid          string    `gorm:"not null"` // References User.Uid (Landlord)
	PropertyName string    `gorm:"not null"`
	Address      string    `gorm:"not null"`
	PropertyType string    `gorm:"not null"` // "Bed Space" or "Apartment"
	RentPrice    float64   `gorm:"not null"`
	LocationLink string    `gorm:"not null"`
	Landmarks    string    `gorm:"not null"`
	Status       string    `gorm:"not null;default:'Pending'"` // "Pending", "Approved", "Rejected", "Open"
	CreatedAt    time.Time `json:"created_at"`
}

// Apartment images
type ApartmentImage struct {
	ID          uint   `gorm:"primaryKey"`
	ApartmentID uint   `gorm:"not null;index"`
	ImageURL    string `gorm:"not null"`
}

// Inquiry model (With automatic expiration & notification)
type Inquiry struct {
	ID          uint      `gorm:"primaryKey"`
	UID         string    `gorm:"not null"` // This links to the `Uid` of User (Tenant)
	ApartmentID uint      `gorm:"not null;index"`
	Message     string    `gorm:"not null"`
	Status      string    `gorm:"not null;default:'Pending'"` // "Pending", "Responded", "Expiring", "Expired"
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `gorm:"not null"`               // Automatically set to CreatedAt + 7 days
	Notified    bool      `gorm:"not null;default:false"` // Tracks if a notification was sent

	// Relationship with User (Tenant)
	User User `gorm:"foreignKey:UID;references:Uid"`
}

// Amenity model
type Amenity struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null;unique"`
}

// Apartment Amenities (Many-to-Many Relationship)
type ApartmentAmenity struct {
	ID          uint `gorm:"primaryKey"`
	ApartmentID uint `gorm:"not null;index"`
	AmenityID   uint `gorm:"not null;index"`
}

// House Rule model
type HouseRule struct {
	ID   uint   `gorm:"primaryKey"`
	Rule string `gorm:"not null;unique"`
}

// Apartment House Rules (Many-to-Many Relationship)
type ApartmentHouseRule struct {
	ID          uint `gorm:"primaryKey"`
	ApartmentID uint `gorm:"not null;index"`
	HouseRuleID uint `gorm:"not null;index"`
}
