package database

import (
	"github.com/iconifyit/go-batch-svg-to-webp/src/models"
)

// GetUserById fetches a single user with associated data.
func (svc *DatabaseService) GetUserById(id int) (*models.User, error) {
	var user models.User
	err := svc.DB.Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUser fetches a single user with associated data based on the query params.
func (svc *DatabaseService) GetUser(params *QueryParams) (*models.User, error) {
	var user models.User
	query := svc.DB.Model(&models.User{}).Preload("Roles").Preload("Images.ImageType")

	if params != nil {
		if params.Filters != nil {
			for _, filter := range params.Filters {
				query = filter(query)
			}
		}
	}

	if err := query.First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers fetches multiple users with associated data based on query params.
func (svc *DatabaseService) GetUsers(params *QueryParams) ([]models.User, error) {
	var users []models.User

	// Start building the query
	query := svc.DB.Model(&models.User{}).Preload("Roles").Preload("Images.ImageType")

	if params != nil {
		// Apply preload logic
		if params.Preload != nil {
			for _, preload := range params.Preload {
				query = query.Preload(preload)
			}
		}

		// Apply custom filters
		if params.Filters != nil {
			for _, filter := range params.Filters {
				query = filter(query)
			}
		}

		// Apply sorting
		if params.SortBy != "" {
			sortOrder := "asc" // Default to ascending
			if params.SortOrder == "desc" {
				sortOrder = "desc"
			}
			query = query.Order(params.SortBy + " " + sortOrder)
		}

		// Apply pagination
		if params.Offset > 0 {
			query = query.Offset(params.Offset)
		}
		if params.Limit > 0 {
			query = query.Limit(params.Limit)
		}
	}

	// Execute the query
	err := query.Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}
