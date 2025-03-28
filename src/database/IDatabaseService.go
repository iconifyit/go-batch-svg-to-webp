package database

import "github.com/iconifyit/go-batch-svg-to-webp/src/models"

type IDatabaseService interface {
	GetFamilyById(id int) (*models.Family, error)
	GetFamily(params *QueryParams) (*models.Family, error)
	GetFamilies(params QueryParams) ([]models.Family, error)
	GetIconById(id int) (*models.Icon, error)
	GetIcon(params QueryParams) (*models.Icon, error)
	GetIcons(params QueryParams) ([]models.Icon, error)
	GetSetById(id int) (*models.Set, error)
	GetSet(params *QueryParams) (*models.Set, error)
	GetSets(params QueryParams) ([]models.Set, error)
	GetIllustrationById(id int) (*models.Illustration, error)
	GetIllustration(params QueryParams) (*models.Illustration, error)
	GetIllustrations(params QueryParams) ([]models.Illustration, error)
	GetUserById(id int) (*models.User, error)
	GetUser(params *QueryParams) (*models.User, error)
	GetUsers(params QueryParams) ([]models.User, error)
	Close() error
}
