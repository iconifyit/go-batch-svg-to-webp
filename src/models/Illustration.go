package models

import (
	"time"
)

type Illustration struct {
	ID        int       `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name;not null"`
	Price     float64   `gorm:"column:price;type:numeric(8,2)"`
	Width     float64   `gorm:"column:width;type:numeric(8,2)"`
	Height    float64   `gorm:"column:height;type:numeric(8,2)"`
	SetID     int       `gorm:"column:set_id"`
	StyleID   int       `gorm:"column:style_id"`
	TeamID    int       `gorm:"column:team_id"`
	LicenseID int       `gorm:"column:license_id;default:21"`
	IsActive  bool      `gorm:"column:is_active;default:true"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
	UserID    int       `gorm:"column:user_id"`
	UniqueID  string    `gorm:"column:unique_id;size:12"`
	ColorData string    `gorm:"column:color_data"`
	IsDeleted bool      `gorm:"column:is_deleted"`
}
