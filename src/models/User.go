package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                   int        `gorm:"primaryKey"`
	Email                string     `gorm:"unique;not null"`
	Username             string     `gorm:"not null"`
	FirstName            string     `gorm:"column:first_name"`
	LastName             string     `gorm:"column:last_name"`
	Password             string     `json:"-"`
	Provider             string     `gorm:"column:provider"`
	ResetPasswordToken   *string    `json:"-"`
	ResetPasswordExpires *time.Time `json:"-"`
	IsActive             bool       `gorm:"column:is_active"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
	IsDeleted            bool       `gorm:"column:is_deleted"`
	DisplayName          string     `gorm:"column:display_name"`
	Image                *string    `gorm:"column:image"`
	IsVerified           bool       `gorm:"column:is_verified"`
	VerifiedAt           *time.Time `gorm:"column:verified_at"`
	UUID                 string     `gorm:"type:uuid;default:gen_random_uuid()"`
	TokenVersion         int        `gorm:"column:token_version"`
	Roles                []Role     `gorm:"many2many:user_to_roles;joinForeignKey:UserID;joinReferences:UserRoleID"`
	Images               []Image    `gorm:"polymorphic:Entity;polymorphicValue:user"` // Polymorphic relationship
}
