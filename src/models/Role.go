package models

import (
	"time"
)

type Role struct {
	ID        int       `gorm:"primaryKey;column:id"`
	Label     string    `gorm:"column:label;not null"`                  // Required
	Value     string    `gorm:"column:value;not null"`                  // Required
	IsActive  bool      `gorm:"column:is_active;not null;default:true"` // Defaults to true
	CreatedAt time.Time `gorm:"column:created_at;not null"`             // Required
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`             // Required
}

// Set the table name for this model
func (Role) TableName() string {
	return "user_roles"
}
