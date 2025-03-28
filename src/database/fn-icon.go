package database

import "image-processor/src/models"

// ==================================================================
// Icon functions
// ==================================================================

// GetIcon fetches a single icon with associated images.
func (svc *DatabaseService) GetIconById(id int) (*models.Icon, error) {
	var icon models.Icon
	err := svc.DB.Preload("Images.ImageType").First(&icon, id).Error
	if err != nil {
		return nil, err
	}
	return &icon, nil
}

// GetIllustrations fetches all illustrations with associated images.
func (svc *DatabaseService) GetIcon(params QueryParams) (*models.Icon, error) {
	var icon models.Icon

	// Start building the query with the model and preload Images.ImageType
	query := svc.DB.Model(&models.Icon{}).Preload("Images.ImageType")

	// Apply additional preloads from QueryParams
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

	// Fetch the first icon that matches the filters
	err := query.First(&icon).Error
	if err != nil {
		return nil, err
	}

	return &icon, nil
}

// GetIllustrations fetches all illustrations with associated images.
func (svc *DatabaseService) GetIcons(params QueryParams) ([]models.Icon, error) {
	var icons []models.Icon

	// Start building the query with the model and preload Images.ImageType
	query := svc.DB.Model(&models.Icon{}).Preload("Images.ImageType")

	// Apply additional preloads from QueryParams
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

	// Apply ordering, limit, and offset
	if params.Order != "" {
		query = query.Order(params.Order)
	}
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	// Execute the query to fetch the icons
	err := query.Find(&icons).Error
	if err != nil {
		return nil, err
	}

	return icons, nil
}
