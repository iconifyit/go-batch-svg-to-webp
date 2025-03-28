package models

import (
	"time"
)

type Icon struct {
	ID        int       `gorm:"primaryKey;column:id" json:"id"`
	Name      string    `gorm:"column:name;not null" json:"name"`
	Price     float64   `gorm:"column:price;type:numeric(8,2)" json:"price"`
	Width     float64   `gorm:"column:width;type:numeric(8,2)" json:"width"`
	Height    float64   `gorm:"column:height;type:numeric(8,2)" json:"height"`
	SetID     int       `gorm:"column:set_id" json:"set_id"`
	StyleID   int       `gorm:"column:style_id" json:"style_id"`
	TeamID    int       `gorm:"column:team_id" json:"team_id"`
	LicenseID int       `gorm:"column:license_id;default:21" json:"license_id"`
	IsActive  bool      `gorm:"column:is_active;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
	UserID    int       `gorm:"column:user_id" json:"user_id"`
	UniqueID  string    `gorm:"column:unique_id;size:12" json:"unique_id"`
	ColorData string    `gorm:"column:color_data" json:"color_data"`
	IsDeleted bool      `gorm:"column:is_deleted" json:"is_deleted"`
}
