package models

import (
	"time"
)

type Family struct {
	ID          int       `gorm:"primaryKey;column:id"`
	Name        string    `gorm:"column:name"`
	Price       float64   `gorm:"column:price;type:numeric(10,2)"`
	LicenseID   int       `gorm:"column:license_id;default:21"`
	TeamID      int       `gorm:"column:team_id"`
	Sort        int       `gorm:"column:sort;default:0"`
	IsActive    bool      `gorm:"column:is_active;default:true"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null"`
	Description string    `gorm:"column:description"`
	UserID      int       `gorm:"column:user_id"`
	UniqueID    string    `gorm:"column:unique_id;size:12"`
	IsDeleted   bool      `gorm:"column:is_deleted"`
	Images      []Image   `gorm:"polymorphic:Entity;polymorphicValue:family"`
}
