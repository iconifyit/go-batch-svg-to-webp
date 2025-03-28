package models

import (
	"time"
)

// UserToRole represents the join table between users and roles
type UserToRole struct {
	ID         int       `gorm:"primaryKey;column:id"`
	UserID     int       `gorm:"column:user_id"`
	UserRoleID int       `gorm:"column:user_role_id"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (UserToRole) TableName() string {
	return "user_to_roles"
}
