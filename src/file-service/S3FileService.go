package fileservice

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3FileService struct {
	UUID         string
	Session      *session.Session
	Config       FileServiceConfig
	BucketName   string
	ObjectKey    string
	SourceBucket string
	TargetBucket string

	// Client is the S3 API used by this service. Leave nil in production to
	// build a real client from Session; inject a fake in tests.
	Client s3iface.S3API
}

// client returns the injected S3 API when set, otherwise a real client
// built from the AWS session.
func (svc *S3FileService) client() s3iface.S3API {
	if svc.Client != nil {
		return svc.Client
	}
	return s3.New(svc.Session)
}

// targetBucket returns the explicitly set BucketName when present, falling
// back to the TargetBucket wired by NewFileService. Write and existence
// operations use this so factory-built services work without manual setup.
func (svc *S3FileService) targetBucket() string {
	if svc.BucketName != "" {
		return svc.BucketName
	}
	return svc.TargetBucket
}

// contentTypeForFile returns the MIME type for a file based on its
// extension, falling back to application/octet-stream for unknown types.
func contentTypeForFile(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	}
	return "application/octet-stream"
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
		input.Bucket = svc.targetBucket()
	}
	if input.Bucket == "" {
		return fmt.Errorf("no target bucket configured: set TransferInput.Bucket or the service TargetBucket")
	}
	log.Printf("svc.Session: %v", svc.Session)
	client := svc.client()
	file, err := os.Open(input.SourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer file.Close()

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(input.Bucket),
		Key:         aws.String(input.TargetFilePath),
		Body:        file,
		ContentType: aws.String(contentTypeForFile(input.SourceFilePath)),
	})
	return err
}

// ListFiles lists files in the requested source bucket (input.SourceRoot),
// falling back to the service's configured source bucket. A nil filter
// accepts every object, matching the LocalFileService behavior.
func (svc *S3FileService) ListFiles(input ListFilesInput, filter func(*string) bool) ([]string, error) {
	var files []string
	if filter == nil {
		filter = func(*string) bool { return true }
	}
	bucket := input.SourceRoot
	if bucket == "" {
		bucket = svc.SourceBucket
	}
	client := svc.client()
	err := client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
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
	client := svc.client()
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
	if svc.targetBucket() == "" {
		return false, fmt.Errorf("no target bucket configured: set BucketName or TargetBucket")
	}
	client := svc.client()
	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(svc.targetBucket()),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// Upload object to S3 bucket
func (svc *S3FileService) Upload(localPath, objectKey string) error {
	if svc.targetBucket() == "" {
		return fmt.Errorf("no target bucket configured: set BucketName or TargetBucket")
	}
	client := svc.client()
	fileBuffer, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer fileBuffer.Close()

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(svc.targetBucket()),
		Key:         aws.String(objectKey),
		Body:        fileBuffer,
		ContentType: aws.String(contentTypeForFile(localPath)),
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}
	return nil
}
