package database

import "image-processor/src/models"

// GetSet fetches a single set with associated images.
func (svc *DatabaseService) GetSetById(id int) (*models.Set, error) {
	var set models.Set
	err := svc.DB.Preload("Images.ImageType").First(&set, id).Error
	if err != nil {
		return nil, err
	}
	return &set, nil
}

// GetSet fetches a single set with associated data based on the query params.
func (svc *DatabaseService) GetSet(params *QueryParams) (*models.Set, error) {
	var set models.Set

	// Start building the query
	query := svc.DB.Model(&models.Set{}).Preload("Images.ImageType")

	// Apply preload logic
	if params != nil {
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
	}

	// Apply ordering
	if params != nil && params.Order != "" {
		query = query.Order(params.Order)
	}

	// Execute the query
	err := query.First(&set).Error
	if err != nil {
		return nil, err
	}

	return &set, nil
}

// GetSets fetches multiple sets with associated data based on query params.
func (svc *DatabaseService) GetSets(params *QueryParams) ([]models.Set, error) {
	var sets []models.Set

	// Start building the query
	query := svc.DB.Model(&models.Set{}).Preload("Images.ImageType")

	// Apply preload logic
	if params != nil {
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

		// Apply ordering
		if params.Order != "" {
			query = query.Order(params.Order)
		}

		// Apply pagination
		if params.Limit > 0 {
			query = query.Limit(params.Limit)
		}
		if params.Offset > 0 {
			query = query.Offset(params.Offset)
		}
	}

	// Execute the query
	err := query.Find(&sets).Error
	if err != nil {
		return nil, err
	}

	return sets, nil
}
