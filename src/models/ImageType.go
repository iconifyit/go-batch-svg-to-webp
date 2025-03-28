package models

import (
	"time"
)

type ImageType struct {
	ID          int       `gorm:"primaryKey" json:"id"`
	Label       string    `gorm:"type:varchar(255);not null" json:"label"`
	Value       string    `gorm:"type:varchar(255);not null;unique" json:"value"`
	Description string    `gorm:"type:varchar(255);not null" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
