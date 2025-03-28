package fileservice

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3FileService struct {
	UUID         string
	Session      *session.Session
	Config       FileServiceConfig
	BucketName   string
	ObjectKey    string
	SourceBucket string
	TargetBucket string
}

func NewS3FileService(config *ServiceInput) IFileService {
	return &S3FileService{
		UUID:    config.UUID,
		Session: config.Session,
	}
}

// func NewS3FileService(config S3FileServiceConfig) IFileService {
// 	return &S3FileService{
// 		UUID:         config.UUID,
// 		Session:      config.Session,
// 		Config:       config.Config,
// 		BucketName:   config.Config.TargetBucket,
// 		ObjectKey:    config.ObjectKey,
// 		SourceBucket: config.SourceBucket,
// 		TargetBucket: config.TargetBucket,
// 	}
// }

// Uploads local file to an s3 bucket.
func (svc *S3FileService) Transfer(input TransferInput) error {
	// input.Bucket is the s3 bucket name
	// input.SourceFilePath is the local file path
	// input.TargetFilePath is the s3 object key
	// input.File - Not used in this implementation
	if input.Bucket == "" {
		input.Bucket = svc.BucketName
	}
	log.Printf("svc.Session: %v", svc.Session)
	client := s3.New(svc.Session)
	file, err := os.Open(input.SourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(input.Bucket),
		Key:         aws.String(input.TargetFilePath),
		Body:        file,
		ContentType: aws.String("image/" + filepath.Ext(input.SourceFilePath)),
	})
	return err
}

// ListFiles lists files in the source directory
func (svc *S3FileService) ListFiles(input ListFilesInput, filter func(*string) bool) ([]string, error) {
	var files []string
	client := s3.New(svc.Session)
	err := client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(svc.BucketName),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			log.Printf("\nObject Key: %s", *obj.Key)
			if filter(obj.Key) {
				files = append(files, *obj.Key)
			}
		}
		return true
	})
	return files, err
}

func (svc *S3FileService) ToImageFiles(files []string) ([]*imagefile.ImageFile, error) {
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

// Downloads a file from an s3 bucket.
func (svc *S3FileService) Download(file *imagefile.ImageFile, dest string) (string, error) {
	client := s3.New(svc.Session)
	s3Object, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(svc.SourceBucket),
		Key:    aws.String(file.ObjectKey),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download file from S3: %v", err)
	}
	defer s3Object.Body.Close()

	localPath := fmt.Sprintf("%s/%s", svc.Config.LocalSource, file.ObjectKey)

	outFile, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %v", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, s3Object.Body); err != nil {
		return "", fmt.Errorf("failed to copy S3 file to local: %v", err)
	}

	return localPath, nil
}

// Checks if an object exists in an s3 bucket.
func (svc *S3FileService) Exists(objectKey string) (bool, error) {
	client := s3.New(svc.Session)
	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(svc.BucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Upload object to S3 bucket
func (svc *S3FileService) Upload(localPath, objectKey string) error {
	client := s3.New(svc.Session)
	fileBuffer, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer fileBuffer.Close()

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(svc.BucketName),
		Key:         aws.String(objectKey),
		Body:        fileBuffer,
		ContentType: aws.String("image/" + filepath.Ext(localPath)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}
	return nil
}
