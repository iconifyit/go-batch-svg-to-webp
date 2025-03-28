package fileservice

import imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"

type IS3FileService interface {
	Transfer(input TransferInput) error
	ListFiles(input ListFilesInput, filter func(*string) bool) ([]string, error)
	ToImageFiles(files []string) ([]*imagefile.ImageFile, error)
	Download(file *imagefile.ImageFile, dest string) (string, error)
	Exists(objectKey string) (bool, error)
	Upload(localPath, objectKey string) error
}
