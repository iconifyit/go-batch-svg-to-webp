package imageprocessor

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/google/uuid"

	fn "github.com/iconifyit/go-batch-svg-to-webp/src/common"
	dbsvc "github.com/iconifyit/go-batch-svg-to-webp/src/database"
	fileservice "github.com/iconifyit/go-batch-svg-to-webp/src/file-service"
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"

	"gorm.io/gorm"
)

// s3 source structure bucket/contributor/(icons|illustrations)/familyUUID/setUUID/file.(svg|png|webp)

type ImageProcessor struct {
	UUID          string                   `json:"uuid"`
	Contributor   string                   `json:"contributor"`
	Config        *Config                  `json:"config"`
	Session       *session.Session         `json:"session"`
	DownloadQueue chan imagefile.ImageFile `json:"-"`
	ProcessQueue  chan imagefile.ImageFile `json:"-"`
	FileService   fileservice.IFileService `json:"fileService"`
}

// Check if a path is a directory
func IsDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// Check if a file exists
func IsFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

// Creates new ImageProcessor instance
func NewImageProcessor(contributor, configFile string) *ImageProcessor {
	var sess *session.Session
	kUUID := uuid.New().String()

	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return nil
	}

	config.Contributor = contributor

	roleArn := config.RoleArn
	region := config.Region

	if roleArn == "" {
		log.Println("RoleArn is required in the config file")
		return nil
	}

	// if !config.IsLocal {
	sess, err = SessionWithRole(roleArn, region)
	if err != nil {
		log.Fatalf("Failed to create session with role: %v", err)
	}
	log.Println("Successfully assumed role and created session")
	// }

	imageProcessor := &ImageProcessor{
		Config:      config,
		UUID:        kUUID,
		Session:     sess,
		Contributor: contributor,
	}

	// If the contributor is not a valid contributor, we should not proceed
	if !imageProcessor.IsValidContributor(contributor) {
		log.Fatalf("Contributor %s is not a valid contributor", contributor)
	}

	config.LocalTarget = filepath.Join(config.WorkDir, contributor, kUUID)

	// Wire the file service with mode-appropriate roots: local paths for
	// local runs, bucket names for S3 runs. The assumed-role session is
	// passed through so the S3 implementation can build a real client.
	sourceRoot := config.LocalSource
	targetRoot := config.LocalTarget
	if !config.IsLocal {
		sourceRoot = config.SourceBucket
		targetRoot = config.TargetBucket
	}

	imageProcessor.FileService = fileservice.NewFileService(fileservice.ServiceInput{
		UUID:       kUUID,
		IsLocal:    config.IsLocal,
		Session:    sess,
		SourceRoot: sourceRoot,
		TargetRoot: targetRoot,
	})

	// Local mode requires the source directory to exist; S3 mode has no
	// local source to check.
	if config.IsLocal {
		exists, err := IsDir(config.LocalSource)
		if err != nil {
			log.Fatalf("Failed to check if local source directory exists: %v", err)
		}
		if !exists {
			log.Fatalf("Local source directory does not exist: %s", config.LocalSource)
		}
	}

	return imageProcessor
}

// SessionWithRole creates a new session with the specified role
func SessionWithRole(roleArn string, region string) (*session.Session, error) {
	// Start a base session (using your current user or environment credentials)
	baseSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	// Create an STS client
	stsClient := sts.New(baseSession)

	// Assume the role
	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String("svg-webp-app-session"),
	}

	assumeRoleOutput, err := stsClient.AssumeRole(assumeRoleInput)
	if err != nil {
		return nil, err
	}

	// Create a new session with the temporary credentials
	tempSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
		Credentials: credentials.NewStaticCredentials(
			*assumeRoleOutput.Credentials.AccessKeyId,
			*assumeRoleOutput.Credentials.SecretAccessKey,
			*assumeRoleOutput.Credentials.SessionToken,
		),
	})
	if err != nil {
		log.Fatalf("Failed to create session with assumed role credentials: %v", err)
	}

	return tempSession, nil
}

// LoadConfig reads the configuration from the provided path.
func LoadConfig(path string) (*Config, error) {
	config, err := NewConfig(path)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return nil, err
	}
	return config, nil
}

// IsLocalRun checks if the image processor is running locally
func (ip *ImageProcessor) IsLocalRun() bool {
	return ip.Config.IsLocal
}

// SetupLogging configures logging based on the provided setting
func (ip *ImageProcessor) SetupLogging() error {
	logSetting := ip.Config.LoggingOutput
	logFilePath := ip.Config.Logfile

	// Set log flags to include file and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Remove the existing log file
	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing log file: %v", err)
	}

	switch logSetting {
	case 0: // No output
		log.SetOutput(io.Discard)
	case 1: // Log to console
		log.SetOutput(os.Stdout)
	case 2: // Log to file
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		log.SetOutput(logFile)
	case 3: // Log to both console and file
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	default:
		return fmt.Errorf("invalid log setting: %d", logSetting)
	}
	return nil
}

// ShouldInclude checks if a file should be included based on the include and
// exclude lists. In local mode, prefixes are matched against the path
// relative to LocalSource; S3 object keys are already bucket-relative and
// are matched as-is, so the same prefixes work in both modes.
func (ip *ImageProcessor) ShouldInclude(filePath *string) bool {
	include := ip.Config.Include
	exclude := ip.Config.Exclude

	if filePath == nil {
		return false
	}

	fileName := filepath.Base(*filePath)

	// Exclude hidden files (e.g., .DS_Store or files starting with '.')
	if strings.HasPrefix(fileName, ".") {
		return false
	}

	// Strip the source root so prefixes like "iconify" match regardless of
	// where the source tree lives on disk. filepath.Rel is path-aware, so
	// unclean roots (./source, trailing slashes) still resolve; paths that
	// cannot be related to the root fall back to literal prefix trimming,
	// which keeps S3 object keys (already root-relative) untouched.
	sourceRoot := ip.Config.SourceBucket
	if ip.Config.IsLocal {
		sourceRoot = ip.Config.LocalSource
	}
	relPath := *filePath
	if sourceRoot != "" {
		// A path is outside the root only when Rel yields ".." itself or a
		// "../" prefix; a plain ".." string prefix would also wrongly match
		// names like "..icons".
		if rel, err := filepath.Rel(sourceRoot, *filePath); err == nil &&
			rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			relPath = filepath.ToSlash(rel)
		} else if strings.HasPrefix(relPath, sourceRoot+"/") {
			// Literal fallback (S3 keys, unrelatable paths): only strip the
			// root when the boundary is a separator, so a sibling like
			// "/data/source-old" is not mangled by a "/data/source" root.
			relPath = relPath[len(sourceRoot)+1:]
		} else if strings.HasSuffix(sourceRoot, "/") && strings.HasPrefix(relPath, sourceRoot) {
			relPath = relPath[len(sourceRoot):]
		}
	}

	// Check for exclusion
	for _, prefix := range exclude {
		if strings.HasPrefix(relPath, prefix) {
			return false // Exclude the file if it matches any prefix in `exclude`
		}
	}

	// If include list is empty, include all files (except excluded ones)
	if len(include) == 0 {
		return true
	}

	// Check for inclusion
	for _, prefix := range include {
		if strings.HasPrefix(relPath, prefix) {
			return true
		}
	}

	// If the file does not match any inclusion rule, exclude it
	return false
}

// ListFiles lists files either locally or from S3
func (ip *ImageProcessor) ListFiles() ([]string, error) {
	var sourceRoot string
	if ip.Config.IsLocal {
		sourceRoot = ip.Config.LocalSource
	} else {
		sourceRoot = ip.Config.SourceBucket
	}
	// S3 listing applies the filter per key, which is cheap. Local listing
	// runs per-file image parsing inside the service when given a filter,
	// so walk with a nil filter instead and apply the include/exclude
	// rules here - ImageFiles() performs the actual parse exactly once.
	if !ip.Config.IsLocal {
		return ip.FileService.ListFiles(
			fileservice.ListFilesInput{SourceRoot: sourceRoot},
			ip.ShouldInclude,
		)
	}

	files, err := ip.FileService.ListFiles(
		fileservice.ListFilesInput{SourceRoot: sourceRoot},
		nil,
	)
	if err != nil {
		return nil, err
	}
	var filtered []string
	for _, file := range files {
		path := file
		if ip.ShouldInclude(&path) {
			filtered = append(filtered, path)
		}
	}
	return filtered, nil
}

// ListDirs lists directories in a given folder, 1-level deep
func (ip *ImageProcessor) ListDirs(rootdir string) ([]string, error) {
	var dirs []string
	if _, err := IsDir(rootdir); err != nil {
		return nil, fmt.Errorf("failed to check if source directory is a directory: %v", err)
	}

	entries, err := os.ReadDir(rootdir)
	if err != nil {
		return nil, fmt.Errorf("failed to read local source directory: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	return dirs, nil
}

// IsValidContributor checks if a contributor is valid based on the database
func (ip *ImageProcessor) IsValidContributor(username string) bool {
	db, err := dbsvc.NewDatabaseService()
	if err != nil {
		log.Fatalf("Failed to create database service: %v", err)
	}

	user, err := db.GetUser(&dbsvc.QueryParams{
		Filters: []func(tx *gorm.DB) *gorm.DB{
			dbsvc.Where("username", username),
		},
	})
	if err != nil {
		return false
	}
	if user == nil {
		return false
	}
	return true
}

// ListFilesForContributor lists files for a specific contributor
func (ip *ImageProcessor) ListFilesForContributor(contributor string) ([]string, error) {
	var files []string
	files, err := ip.FileService.ListFiles(fileservice.ListFilesInput{
		SourceRoot: ip.Config.LocalSource + "/" + contributor,
	}, nil)
	return files, err
}

// Worker function to process files
func (ip *ImageProcessor) worker(id int, jobs <-chan imagefile.ImageFile, wg *sync.WaitGroup, errorChan chan<- error) {
	defer wg.Done()
	for file := range jobs {
		log.Printf("Worker %d: Processing file %v", id, file)
		if err := ip.ProcessFile(file); err != nil {
			errorChan <- fmt.Errorf("worker %d: error processing file %v: %v", id, file, err)
			return
		}
		log.Printf("Worker %d: Successfully processed file: %v", id, file)
	}
}

// Convert a list of filepaths to a list of ImageFile objects
func (ip *ImageProcessor) ImageFiles(files []string) []imagefile.ImageFile {
	var imageFiles []imagefile.ImageFile
	for _, file := range files {
		image, err := imagefile.NewImageFile(file)
		if err != nil || image == nil {
			log.Printf("Path %s is not an image file. Skipping...", file)
			continue
		}
		log.Printf("ImageFile: %+v", fn.ToJSON(image))
		imageFiles = append(imageFiles, *image)
	}
	return imageFiles
}

// ProcessFiles processes all files
func (ip *ImageProcessor) ProcessFiles() error {
	files, err := ip.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	log.Printf("Files to process: %v", files)

	// Create a channel to queue up files for processing
	jobs := make(chan imagefile.ImageFile, len(files))
	errorChan := make(chan error, 1) // Buffer size 1 for fail-fast error handling

	// Create a wait group to synchronize the workers
	var wg sync.WaitGroup

	// Start the workers
	for i := 1; i <= ip.Config.WorkerPoolSize; i++ {
		wg.Add(1)
		go ip.worker(i, jobs, &wg, errorChan) // Call worker with all required arguments
	}

	imageFiles := ip.ImageFiles(files)

	// Queue up the files for processing in a separate goroutine
	go func() {
		for _, file := range imageFiles {
			select {
			case jobs <- file:
			case <-errorChan:
				close(jobs)
				return
			}
		}
		close(jobs)
	}()

	// Wait for workers in a separate goroutine
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Monitor for errors or completion
	select {
	case err := <-errorChan:
		log.Printf("Critical error encountered: %v", err)
		close(jobs)
		return err
	case <-done:
		log.Println("All workers completed successfully.")
		return nil
	}
}

// Run starts the image processing
func (ip *ImageProcessor) Run() error {
	if err := ip.SetupLogging(); err != nil {
		return fmt.Errorf("failed to set up logging: %v", err)
	}

	if fn.StringInSlice(ip.Contributor, ip.Config.Exclude) {
		log.Fatalf("Contributor %s is in the exclude list", ip.Contributor)
	}

	log.Printf("Config: %v", fn.ToJSON(ip.Config))

	files, err := ip.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	imageFiles := ip.ImageFiles(files)

	log.Printf("Files for contributor %s: %d", ip.Contributor, len(files))
	log.Printf("ImageFiles for contributor %s : %v", ip.Contributor, len(imageFiles))

	// Initialize queues
	ip.DownloadQueue = make(chan imagefile.ImageFile, len(imageFiles))
	ip.ProcessQueue = make(chan imagefile.ImageFile, len(imageFiles))
	errorChan := make(chan error, 1)

	// WaitGroups for synchronization
	var downloadWG sync.WaitGroup
	var processWG sync.WaitGroup

	// Start download workers
	for i := 0; i < ip.Config.DownloadWorkerPoolSize; i++ {
		downloadWG.Add(1)
		go func(workerID int) {
			defer downloadWG.Done()
			ip.downloadWorker(workerID, errorChan)
		}(i)
	}

	// Start processing workers
	for i := 0; i < ip.Config.ProcessWorkerPoolSize; i++ {
		processWG.Add(1)
		go func(workerID int) {
			defer processWG.Done()
			ip.processWorker(workerID, errorChan)
		}(i)
	}

	// Queue up files for downloading
	go func() {
		for _, file := range imageFiles {
			ip.DownloadQueue <- file
		}
		close(ip.DownloadQueue)
	}()

	// Wait for downloads in a separate goroutine
	go func() {
		downloadWG.Wait()
		close(ip.ProcessQueue)
	}()

	// Wait for processing workers to complete in a separate goroutine
	go func() {
		processWG.Wait()
		close(errorChan)
	}()

	// Monitor for errors or completion
	for err := range errorChan {
		if err != nil {
			log.Printf("Critical error encountered: %v", err)
			return err
		}
	}

	// Cleanup temporary files and directories
	ip.Cleanup()

	log.Println("Image processor completed successfully.")
	return nil
}
