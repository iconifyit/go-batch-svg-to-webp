package models

import (
	"time"
)

type ImageType struct {
	ID          int       `gorm:"primaryKey"`
	Label       string    `gorm:"type:varchar(255);not null"`
	Value       string    `gorm:"type:varchar(255);not null;unique"`
	Description string    `gorm:"type:varchar(255);not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
