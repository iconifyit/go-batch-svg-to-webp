package database

import "gorm.io/gorm"

type QueryParams struct {
	Preload   []string
	Offset    int
	Limit     int
	Filters   []func(tx *gorm.DB) *gorm.DB
	Order     string
	SortBy    string
	SortOrder string
}
