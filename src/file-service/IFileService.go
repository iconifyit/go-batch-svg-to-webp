package fileservice

import (
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

type IFileService interface {
	Transfer(input TransferInput) error
	ListFiles(input ListFilesInput, callback func(*string) bool) ([]string, error)
	ToImageFiles(files []string) ([]*imagefile.ImageFile, error)
	Download(file *imagefile.ImageFile, dest string) (string, error)
}

// NewFileService selects the storage implementation for the run. It
// delegates to the exported constructors so factory-built and directly
// constructed services always share the same wiring.
func NewFileService(input ServiceInput) IFileService {
	if input.IsLocal {
		return NewLocalFileService(&input)
	}
	return NewS3FileService(&input)
}
