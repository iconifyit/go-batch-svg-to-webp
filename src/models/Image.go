package models

import (
	"time"
)

type Image struct {
	ID          int        `gorm:"primaryKey"`
	EntityID    int        `gorm:"not null"`
	EntityType  string     `gorm:"type:text;not null;check:entity_type IN ('family', 'icon', 'illustration', 'set', 'user', 'team', 'category')"`
	Visibility  string     `gorm:"type:text;not null;check:visibility IN ('public', 'private', 'hidden')"`
	Access      string     `gorm:"type:text;not null;check:access IN ('admin', 'owner', 'customer', 'purchaser', 'subscriber', 'user', 'all')"`
	Name        string     `gorm:"type:varchar(255);not null"`
	FileType    string     `gorm:"type:text;not null;check:file_type IN ('png', 'svg', 'webp')"`
	URL         string     `gorm:"type:text;not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	ImageHash   string     `gorm:"type:varchar(255)"`
	UniqueID    string     `gorm:"type:varchar(12);unique;default:substring(gen_random_uuid()::text from 25 for 12)"`
	ImageTypeID *int       `gorm:"index"`
	ColorData   string     `gorm:"type:varchar(255)"`
	ObjectKey   string     `gorm:"type:varchar(255)"`
	IsActive    bool       `gorm:"not null;default:true"`
	IsDeleted   *bool      `gorm:"default:false"`
	ImageType   *ImageType `gorm:"foreignKey:ImageTypeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
