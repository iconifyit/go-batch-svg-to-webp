package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                   int        `gorm:"primaryKey" json:"id"`
	Email                string     `gorm:"unique;not null" json:"email"`
	Username             string     `gorm:"not null" json:"username"`
	FirstName            string     `gorm:"column:first_name" json:"first_name"`
	LastName             string     `gorm:"column:last_name" json:"last_name"`
	Password             string     `json:"-"` // hidden in API output
	Provider             string     `gorm:"column:provider" json:"provider"`
	ResetPasswordToken   *string    `json:"-"`
	ResetPasswordExpires *time.Time `json:"-"`
	IsActive             bool       `gorm:"column:is_active" json:"is_active"`
	CreatedAt            time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at" json:"updated_at"`
	IsDeleted            bool       `gorm:"column:is_deleted" json:"is_deleted"`
	DisplayName          string     `gorm:"column:display_name" json:"display_name"`
	Image                *string    `gorm:"column:image" json:"image"`
	IsVerified           bool       `gorm:"column:is_verified" json:"is_verified"`
	VerifiedAt           *time.Time `gorm:"column:verified_at" json:"verified_at"`
	UUID                 string     `gorm:"type:uuid;default:gen_random_uuid()" json:"uuid"`
	TokenVersion         int        `gorm:"column:token_version" json:"token_version"`
	Roles                []Role     `gorm:"many2many:user_to_roles;joinForeignKey:UserID;joinReferences:UserRoleID" json:"roles"`
	Images               []Image    `gorm:"polymorphic:Entity;polymorphicValue:user" json:"images"`
}
