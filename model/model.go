package model

import (
	"time"
)

// Landlord confirmation to delete "rejected" apartment
type DeleteApartmentRequest struct {
	Confirm bool `json:"confirm"` // Landlord must confirm deletion
}

type User struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	Uid           string    `gorm:"uniqueIndex"`
	Email         string    `gorm:"unique"`
	PhoneNumber   string    `json:"phone_number"`
	Password      string    `json:"password,omitempty"` // Optional for email sign-up
	FirstName     string    `json:"first_name"`
	MiddleInitial string    `json:"middle_initial"`
	LastName      string    `json:"last_name"`
	Age           int       `json:"age"`
	Address       string    `json:"address"`
	ValidID       string    `json:"valid_id"`
	AccountStatus string    `gorm:"not null;default:'Pending'" json:"account_status"` // "Verified" / "Unverified"
	Provider      string    `gorm:"not null" json:"provider"`                         // "email", "google", "facebook"
	PhotoURL      string    `json:"photo_url"`
	UserType      string    `gorm:"not null" json:"user_type"` // "Landlord", "Tenant", "Admin"
	Birthday      string    `gorm:"not null" json:"birthday"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Admin model (Separate from User)
type Admins struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"unique;not null"`
	Password  string    `gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// Landlord Profile (Separate from User)
type LandlordProfile struct {
	ID             uint   `gorm:"primaryKey"`
	UserID         uint   `gorm:"not null;unique"`
	BusinessName   string `json:"business_name"`
	BusinessPermit string `json:"business_permit"`
}

// Apartment model
type Apartment struct {
	ID            uint    `gorm:"primaryKey"`
	UserID        uint    `gorm:"not null"`
	PropertyName  string  `gorm:"not null"`
	Address       string  `gorm:"not null"`
	PropertyType  string  `gorm:"not null"` // "Bed Space" or "Apartment"
	RentPrice     float64 `gorm:"not null"`
	LocationLink  string  `gorm:"not null"`
	Landmarks     string  `gorm:"not null"`
	ContactNumber string  `gorm:"not null"`
	Email         string  `gorm:"not null"`
	Facebook      string
	Status        string `gorm:"not null;default:'Pending'"` // "Pending", "Approved", "Rejected", "Open"
	CreatedAt     time.Time
}

// Apartment images
type ApartmentImage struct {
	ID          uint   `gorm:"primaryKey"`
	ApartmentID uint   `gorm:"not null"`
	ImageURL    string `gorm:"not null"`
}

// Inquiry model (With automatic expiration & notification)
type Inquiry struct {
	ID          uint      `gorm:"primaryKey"`
	TenantID    uint      `gorm:"not null"`
	ApartmentID uint      `gorm:"not null"`
	Message     string    `gorm:"not null"`
	Status      string    `gorm:"not null;default:'Pending'"` // "Pending", "Responded", "Expiring", "Expired"
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `gorm:"not null"`               // Automatically set to CreatedAt + 7 days
	Notified    bool      `gorm:"not null;default:false"` // Tracks if a notification was sent
}

// Amenity model
type Amenity struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null;unique"`
}

type ApartmentAmenity struct {
	ID          uint `gorm:"primaryKey"`
	ApartmentID uint `gorm:"not null"`
	AmenityID   uint `gorm:"not null"`
}

// House Rule model
type HouseRule struct {
	ID   uint   `gorm:"primaryKey"`
	Rule string `gorm:"not null;unique"`
}

type ApartmentHouseRule struct {
	ID          uint `gorm:"primaryKey"`
	ApartmentID uint `gorm:"not null"`
	HouseRuleID uint `gorm:"not null"`
}
