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
	Fullname      string    `json:"fullname"`
	Age           int       `json:"age"`
	Address       string    `json:"address"`
	ValidID       string    `json:"valid_id"`
	AccountStatus string    `gorm:"not null;default:'Pending'" json:"account_status"` // "Verified" / "Unverified"
	Provider      string    `gorm:"not null" json:"provider"`
	PhotoURL      string    `json:"photo_url"`
	UserType      string    `gorm:"not null" json:"user_type"` // "Landlord", "Tenant", "Admin"
	Birthday      time.Time `json:"birthday"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Apartment struct {
	ID           uint      `gorm:"primaryKey"`
	Uid          string    `gorm:"not null"` // Landlord's UID; no foreign key constraint here.
	PropertyName string    `gorm:"not null"`
	Address      string    `gorm:"not null"`
	PropertyType string    `gorm:"not null"`
	RentPrice    float64   `gorm:"not null"`
	LocationLink string    `gorm:"not null"`
	Landmarks    string    `gorm:"not null"`
	Status       string    `gorm:"not null;default:'Pending'"`
	Latitude     float64   `gorm:"null"`
	Longitude    float64   `gorm:"null"`
	UserID       string    `gorm:"not null"` // Add this field to reference the landlord's UID
	CreatedAt    time.Time `json:"created_at"`
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

// Apartment images
type ApartmentImage struct {
	ID          uint      `gorm:"primaryKey"`
	ApartmentID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	ImageURL    string    `gorm:"not null"`
	Apartment   Apartment `gorm:"foreignKey:ApartmentID;references:ID;constraint:OnDelete:CASCADE"`
}

// Apartment videos
type ApartmentVideo struct {
	ID          uint      `gorm:"primaryKey"`
	ApartmentID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	VideoURL    string    `gorm:"not null"`
	Apartment   Apartment `gorm:"foreignKey:ApartmentID;references:ID;constraint:OnDelete:CASCADE"`
}

// Inquiry model (With automatic expiration & notification)
type Inquiry struct {
	ID          uint      `gorm:"primaryKey"`
	UID         string    `gorm:"not null"`
	ApartmentID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	Message     string    `gorm:"not null"`
	Status      string    `gorm:"not null;default:'Pending'"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `gorm:"not null"`
	Notified    bool      `gorm:"not null;default:false"`
	User        User      `gorm:"foreignKey:UID;references:Uid"`
	Apartment   Apartment `gorm:"foreignKey:ApartmentID;references:ID;constraint:OnDelete:CASCADE"`
}

// Amenity model
type Amenity struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null;unique"`
}

// Apartment Amenities (Many-to-Many Relationship)
type ApartmentAmenity struct {
	ID          uint      `gorm:"primaryKey"`
	ApartmentID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	AmenityID   uint      `gorm:"not null;index"`
	Apartment   Apartment `gorm:"foreignKey:ApartmentID;references:ID;constraint:OnDelete:CASCADE"`
}

// House Rule model
type HouseRule struct {
	ID   uint   `gorm:"primaryKey"`
	Rule string `gorm:"not null;unique"`
}

// Apartment House Rules (Many-to-Many Relationship)
type ApartmentHouseRule struct {
	ID          uint      `gorm:"primaryKey"`
	ApartmentID uint      `gorm:"not null;index;constraint:OnDelete:CASCADE"`
	HouseRuleID uint      `gorm:"not null;index"`
	Apartment   Apartment `gorm:"foreignKey:ApartmentID;references:ID;constraint:OnDelete:CASCADE"`
}

type Wishlist struct {
	ID          uint      `gorm:"primaryKey"`
	UID         string    `gorm:"not null"`                             // Tenant's UID
	ApartmentID uint      `gorm:"not null;constraint:OnDelete:CASCADE"` // Foreign key referencing the Apartment model's ID
	CreatedAt   time.Time `json:"created_at"`
	Apartment   Apartment `gorm:"foreignKey:ApartmentID;references:ID;constraint:OnDelete:CASCADE"`
}
