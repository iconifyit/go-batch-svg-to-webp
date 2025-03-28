package fileservice

import (
	imagefile "image-processor/src/image-file"
)

type IFileService interface {
	Transfer(input TransferInput) error
	ListFiles(input ListFilesInput, callback func(*string) bool) ([]string, error)
	ToImageFiles(files []string) ([]*imagefile.ImageFile, error)
	Download(file *imagefile.ImageFile, dest string) (string, error)
}

func NewFileService(input ServiceInput) IFileService {
	if input.IsLocal {
		return &LocalFileService{
			SourceRoot: input.SourceRoot,
			TargetRoot: input.TargetRoot,
		}
	}
	return &S3FileService{
		SourceBucket: input.SourceRoot,
		TargetBucket: input.TargetRoot,
	}
}
