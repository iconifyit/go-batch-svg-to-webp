package models

import (
	"time"
)

type Role struct {
	ID        int       `gorm:"primaryKey;column:id" json:"id"`
	Label     string    `gorm:"column:label;not null" json:"label"`
	Value     string    `gorm:"column:value;not null" json:"value"`
	IsActive  bool      `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updated_at"`
}

// Set the table name for this model
func (Role) TableName() string {
	return "user_roles"
}
