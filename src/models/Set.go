package models

import (
	"time"
)

type Set struct {
	ID          int       `gorm:"primaryKey;column:id" json:"id"`
	Name        string    `gorm:"column:name;not null" json:"name"`
	Price       float64   `gorm:"column:price;type:numeric(10,2)" json:"price"`
	FamilyID    int       `gorm:"column:family_id" json:"family_id"`
	LicenseID   int       `gorm:"column:license_id;default:21" json:"license_id"`
	TypeID      int       `gorm:"column:type_id" json:"type_id"`
	StyleID     int       `gorm:"column:style_id" json:"style_id"`
	TeamID      int       `gorm:"column:team_id" json:"team_id"`
	Sort        int       `gorm:"column:sort;default:0" json:"sort"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	UserID      int       `gorm:"column:user_id" json:"user_id"`
	UniqueID    string    `gorm:"column:unique_id;size:12" json:"unique_id"`
	Description string    `gorm:"column:description" json:"description"`
	IsDeleted   bool      `gorm:"column:is_deleted" json:"is_deleted"`
}
