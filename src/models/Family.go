package models

import (
	"time"
)

type Family struct {
	ID          int       `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name" json:"name"`
	Price       float64   `gorm:"column:price;type:numeric(10,2)" json:"price"`
	LicenseID   int       `gorm:"column:license_id;default:21" json:"license_id"`
	TeamID      int       `gorm:"column:team_id" json:"team_id"`
	Sort        int       `gorm:"column:sort;default:0" json:"sort"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	Description string    `gorm:"column:description" json:"description"`
	UserID      int       `gorm:"column:user_id" json:"user_id"`
	UniqueID    string    `gorm:"column:unique_id;size:12" json:"unique_id"`
	IsDeleted   bool      `gorm:"column:is_deleted" json:"is_deleted"`
	Images      []Image   `gorm:"polymorphic:Entity;polymorphicValue:family" json:"images"`
}
