package imagefile

type IImageFile interface {
	Exists() (bool, error)
	ToJSON() map[string]interface{}
	ToSnakeCase() map[string]interface{}
}
