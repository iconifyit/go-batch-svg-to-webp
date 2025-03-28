package database

import "github.com/iconifyit/go-batch-svg-to-webp/src/models"

// ==================================================================
// Family functions
// ==================================================================

// GetFamily fetches a single family with associated images.
func (svc *DatabaseService) GetFamilyById(id int) (*models.Family, error) {
	var family models.Family
	err := svc.DB.Preload("Images.ImageType").First(&family, id).Error
	if err != nil {
		return nil, err
	}
	return &family, nil
}

// GetFamily fetches a single family with associated images based on the query params.
func (svc *DatabaseService) GetFamily(params *QueryParams) (*models.Family, error) {
	var family models.Family

	// Start building the query
	query := svc.DB.Model(&models.Family{}).Preload("Images.ImageType")

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
	err := query.First(&family).Error
	if err != nil {
		return nil, err
	}

	return &family, nil
}

// GetFamilies fetches all families with associated images.
// GetFamilies fetches all families with associated images.
func (svc *DatabaseService) GetFamilies(params QueryParams) ([]models.Family, error) {
	var families []models.Family

	// Start building the query with default preload for Images and their ImageType
	query := svc.DB.Model(&models.Family{}).Preload("Images.ImageType")

	// Apply additional preload logic from QueryParams
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

	// Execute the query
	err := query.Find(&families).Error
	if err != nil {
		return nil, err
	}

	return families, nil
}
