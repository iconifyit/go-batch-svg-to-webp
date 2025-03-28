package models

import (
	"time"
)

type Set struct {
	ID          int       `gorm:"primaryKey;column:id"`
	Name        string    `gorm:"column:name;not null"`
	Price       float64   `gorm:"column:price;type:numeric(10,2)"`
	FamilyID    int       `gorm:"column:family_id"`
	LicenseID   int       `gorm:"column:license_id;default:21"`
	TypeID      int       `gorm:"column:type_id"`
	StyleID     int       `gorm:"column:style_id"`
	TeamID      int       `gorm:"column:team_id"`
	Sort        int       `gorm:"column:sort;default:0"`
	IsActive    bool      `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
	UserID      int       `gorm:"column:user_id"`
	UniqueID    string    `gorm:"column:unique_id;size:12"`
	Description string    `gorm:"column:description"`
	IsDeleted   bool      `gorm:"column:is_deleted"`
}
