package fileservice

import (
	"fmt"
	fn "image-processor/src/common"
	imagefile "image-processor/src/image-file"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws/session"
)

type LocalFileService struct {
	UUID       string
	Session    *session.Session
	Config     FileServiceConfig
	SourceRoot string
	TargetRoot string
}

// Creates a new LocalFileService instance
// @Param config *ServiceInput
// @Return IFileService
func NewLocalFileService(config *ServiceInput) IFileService {
	return &LocalFileService{
		UUID:    config.UUID,
		Session: config.Session,
	}
}

// Transfer copies a file from the local source to the destination
// The name "Transfer" was selected because it is generic and can
// apply to both local and remote file copies, (e.g., S3)
// SourceFilePath: webpOutputFilePath,
// TargetFilePath: outputKey,
func (svc *LocalFileService) Transfer(input TransferInput) error {
	if err := fn.CopyFile(input.SourceFilePath, input.TargetFilePath); err != nil {
		return fmt.Errorf(
			"failed to copy local file %s to %s: %v",
			input.SourceFilePath,
			input.TargetFilePath,
			err,
		)
	}
	return nil
}

// ListFiles lists files in the source directory
func (svc *LocalFileService) ListFiles(input ListFilesInput, filter func(*string) bool) ([]string, error) {
	var files []string

	log.Printf("ListFiles : ListFilesInput %v", fn.ToJSON(input))

	// Output absolute paths for local files
	err := filepath.Walk(input.SourceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// log.Printf("ListFiles : path %v", path)
		if !info.IsDir() {
			if filter == nil {
				files = append(files, path)
			} else {
				if filter(&path) {
					imageFile, err := imagefile.NewImageFile(path)
					if err != nil || imageFile == nil {
						log.Printf("Path %s is not an image file. Skipping...", path)
					} else {
						log.Printf("ImageFile: %+v", imageFile)
						files = append(files, path) // Absolute paths
					}
				}
			}
			// files = append(files, path) // Absolute paths
		}
		return nil
	})
	return files, err
}

// ToImageFiles converts a list of file paths to ImageFile objects
// It filters out any paths that are not valid image files
// and returns a slice of valid ImageFile objects
// @Param files []string
// @Return []*imagefile.ImageFile
// @Return error
func (svc *LocalFileService) ToImageFiles(files []string) ([]*imagefile.ImageFile, error) {
	var imageFiles []*imagefile.ImageFile
	for _, file := range files {
		image, err := imagefile.NewImageFile(file)
		if err != nil || image == nil {
			log.Printf("Path %s is not an image file. Skipping...", file)
			continue
		}
		imageFiles = append(imageFiles, image)
	}
	return imageFiles, nil
}

// Download copies a file from the source to the destination
// The source is a local file path and the destination is a local file path
// @Param file *imagefile.ImageFile
// @Param dest string
// @Return string
// @Return error
func (svc *LocalFileService) Download(file *imagefile.ImageFile, dest string) (string, error) {
	// Construct the local path in the working directory
	log.Printf("Downloading file : %s to %s", file.ObjectKey, dest)

	// Ensure the directory structure exists for the local path
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory for local file: %v", err)
	}

	// Copy local files
	downloadFile := filepath.Join(svc.SourceRoot, file.ObjectKey)
	log.Printf("Copying from %s to %s", file.ObjectKey, dest)
	if err := fn.CopyFile(downloadFile, dest); err != nil {
		return "", fmt.Errorf("failed to copy local file: %v", err)
	}

	return dest, nil
}
