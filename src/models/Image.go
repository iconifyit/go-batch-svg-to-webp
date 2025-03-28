package models

import (
	"time"
)

type Image struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	EntityID    int        `gorm:"not null" json:"entity_id"`
	EntityType  string     `gorm:"type:text;not null;check:entity_type IN ('family', 'icon', 'illustration', 'set', 'user', 'team', 'category')" json:"entity_type"`
	Visibility  string     `gorm:"type:text;not null;check:visibility IN ('public', 'private', 'hidden')" json:"visibility"`
	Access      string     `gorm:"type:text;not null;check:access IN ('admin', 'owner', 'customer', 'purchaser', 'subscriber', 'user', 'all')" json:"access"`
	Name        string     `gorm:"type:varchar(255);not null" json:"name"`
	FileType    string     `gorm:"type:text;not null;check:file_type IN ('png', 'svg', 'webp')" json:"file_type"`
	URL         string     `gorm:"type:text;not null" json:"url"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	ImageHash   string     `gorm:"type:varchar(255)" json:"image_hash"`
	UniqueID    string     `gorm:"type:varchar(12);unique;default:substring(gen_random_uuid()::text from 25 for 12)" json:"unique_id"`
	ImageTypeID *int       `gorm:"index" json:"image_type_id"`
	ColorData   string     `gorm:"type:varchar(255)" json:"color_data"`
	ObjectKey   string     `gorm:"type:varchar(255)" json:"object_key"`
	IsActive    bool       `gorm:"not null;default:true" json:"is_active"`
	IsDeleted   *bool      `gorm:"default:false" json:"is_deleted"`
	ImageType   *ImageType `gorm:"foreignKey:ImageTypeID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"image_type"`
}
