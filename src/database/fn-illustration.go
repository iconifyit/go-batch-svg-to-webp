package database

import "github.com/iconifyit/go-batch-svg-to-webp/src/models"

// ==================================================================
// Illustration functions
// ==================================================================

// GetIllustration fetches a single illustration with associated images.
func (svc *DatabaseService) GetIllustrationById(id int) (*models.Illustration, error) {
	var illustration models.Illustration
	err := svc.DB.Preload("Images.ImageType").First(&illustration, id).Error
	if err != nil {
		return nil, err
	}
	return &illustration, nil
}

// GetIllustrations fetches all illustrations with associated images.
func (svc *DatabaseService) GetIllustration(params QueryParams) (*models.Illustration, error) {
	var illustration models.Illustration

	// Start building the query with the model and preload Images.ImageType
	query := svc.DB.Model(&models.Illustration{}).Preload("Images.ImageType")

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

	// Fetch the first illustration that matches the filters
	err := query.First(&illustration).Error
	if err != nil {
		return nil, err
	}

	return &illustration, nil
}

// GetIllustrations fetches all illustrations with associated images.
func (svc *DatabaseService) GetIllustrations(params QueryParams) ([]models.Illustration, error) {
	var illustrations []models.Illustration

	// Start building the query with the model and preload Images.ImageType
	query := svc.DB.Model(&models.Illustration{}).Preload("Images.ImageType")

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

	// Execute the query to fetch the illustrations
	err := query.Find(&illustrations).Error
	if err != nil {
		return nil, err
	}

	return illustrations, nil
}
