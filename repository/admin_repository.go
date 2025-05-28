package repository

import (
	"github.com/Conding-Student/backend/model" // adjust the import path

	"gorm.io/gorm"
)

// SaveAdminToken inserts a new token for a specific admin.
func SaveAdminToken(db *gorm.DB, adminID uint, token string) error {
	newToken := model.AdminToken{
		AdminID: adminID,
		Token:   token,
	}
	return db.Create(&newToken).Error
}
