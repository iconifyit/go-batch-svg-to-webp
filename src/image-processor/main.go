package imageprocessor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/google/uuid"
)

type ImageProcessor struct {
	UUID          string
	Config        *Config
	Session       *session.Session
	DownloadQueue chan string
	ProcessQueue  chan string
}

// SetupLogging configures logging based on the provided setting
func (ip *ImageProcessor) SetupLogging() error {
	logSetting := ip.Config.LoggingOutput
	logFilePath := ip.Config.Logfile

	// Remove the existing log file
	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing log file: %v", err)
	}

	switch logSetting {
	// No output
	case 0:
		log.SetOutput(io.Discard)
		// Log to console
	case 1:
		log.SetOutput(os.Stdout)
		// Log to file
	case 2:
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		log.SetOutput(logFile)
		// Log to both console and file
	case 3:
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

// ShouldInclude checks if a file should be included based on the include and exclude lists
func (ip *ImageProcessor) ShouldInclude(filePath *string) bool {
	include := ip.Config.Include
	exclude := ip.Config.Exclude

	if filePath == nil {
		return false
	}

	log.Printf("ShouldInclude file: %s", *filePath)
	log.Printf("Include: %v", include)
	log.Printf("Exclude: %v\n", exclude)

	fileName := filepath.Base(*filePath)

	// Exclude hidden files (e.g., .DS_Store or files starting with '.')
	if strings.HasPrefix(fileName, ".") {
		return false
	}

	// Check for exclusion
	for _, prefix := range exclude {
		log.Printf("Checking prefix: %s - %s", prefix, *filePath)
		if strings.HasPrefix(*filePath, prefix) {
			return false // Exclude the file if it matches any prefix in `exclude`
		}
	}

	// If include list is empty, include all files (except excluded ones)
	if len(include) == 0 {
		return true
	}

	// Check for inclusion
	for _, prefix := range include {
		if strings.HasPrefix(*filePath, prefix) {
			return true
		}
	}

	// If the file does not match any inclusion rule, exclude it
	return false
}

// ListFiles lists files either locally or from S3
func (ip *ImageProcessor) ListFiles() ([]string, error) {
	var files []string

	if ip.Config.IsLocal {
		err := filepath.Walk(ip.Config.LocalSource, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, _ := filepath.Rel(ip.Config.LocalSource, path)
				if ip.ShouldInclude(&relPath) {
					files = append(files, relPath)
				}
			}
			return nil
		})
		return files, err
	}

	svc := s3.New(ip.Session)
	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(ip.Config.SourceBucket),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			log.Printf("\nObject Key: %s", *obj.Key)
			if ip.ShouldInclude(obj.Key) {
				files = append(files, *obj.Key)
			}
		}
		return true
	})

	return files, err
}

// CopyFile copies a file from src to dst
func (ip *ImageProcessor) CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// ConvertSVGToPNG converts an SVG file to a PNG file
func (ip *ImageProcessor) ConvertSVGToPNG(svgPath, pngPath string, size int) error {
	cmd := exec.Command(
		"rsvg-convert",
		"--background-color",
		"white",
		"-w",
		fmt.Sprint(size),
		"-h",
		fmt.Sprint(size),
		svgPath,
		"-o",
		pngPath,
	)
	return cmd.Run()
}

// RunFFmpeg runs an FFmpeg command to convert a video file to WebP
func (ip *ImageProcessor) RunFFmpeg(input, output string) error {
	var args []string

	if ip.Config.UseHardwareAcceleration {
		log.Printf("\nUsing hardware acceleration: ffmpeg -hwaccel videotoolbox -i %s -vf format=yuv420p -q:v 75 %s\n", input, output)
		args = []string{
			"-hwaccel", "videotoolbox", // Use VideoToolbox for hardware acceleration
			"-i", input,
			"-vf", "format=yuv420p",
			"-q:v", "75",
			output,
		}
	} else {
		log.Printf("\nUsing software encoding: ffmpeg -i %s -vf format=yuv420p -q:v 75 %s\n", input, output)
		args = []string{
			"-i", input,
			"-vf", "format=yuv420p",
			"-q:v", "75",
			output,
		}
	}

	cmd := exec.Command(ip.Config.FFmpegPath, args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v, output: %s", err, string(outputBytes))
	}

	return nil
}

// Watermark adds a watermark to an image
func (ip *ImageProcessor) Watermark(inputFilePath, outputFilePath string, size int) error {
	ffmpegPath := ip.Config.FFmpegPath
	watermarkSVGPath := ip.Config.WatermarkPath

	// Create a unique temporary file for the watermark PNG
	tempDir := filepath.Dir(ip.Config.WorkDir)
	tempWatermarkPNG := filepath.Join(tempDir, fmt.Sprintf("watermark-%s.png", uuid.New().String()))

	log.Printf("Converting watermark SVG to PNG: %s -> %s\n", watermarkSVGPath, tempWatermarkPNG)

	// Convert watermark SVG to PNG
	cmd := exec.Command("rsvg-convert", "-w", fmt.Sprint(size), "-h", fmt.Sprint(size), "-o", tempWatermarkPNG, watermarkSVGPath)
	if outputBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to convert watermark SVG to PNG: %v, output: %s", err, string(outputBytes))
	}

	// Build the ffmpeg command
	var args []string
	if ip.Config.UseHardwareAcceleration {
		log.Printf("Using hardware acceleration for watermark: ffmpeg -hwaccel videotoolbox -i %s -i %s -filter_complex overlay=W-w-10:H-h-10 -q:v 75 %s\n", inputFilePath, tempWatermarkPNG, outputFilePath)
		args = []string{
			"-hwaccel", "videotoolbox", // Apply hardware acceleration for decoding
			"-i", inputFilePath,
			"-i", tempWatermarkPNG,
			"-filter_complex", "overlay=W-w-10:H-h-10", // Apply watermark overlay
			"-q:v", "75",
			outputFilePath,
		}
	} else {
		log.Printf("Using software encoding for watermark: ffmpeg -i %s -i %s -filter_complex overlay=W-w-10:H-h-10 -q:v 75 %s\n", inputFilePath, tempWatermarkPNG, outputFilePath)
		args = []string{
			"-i", inputFilePath,
			"-i", tempWatermarkPNG,
			"-filter_complex", "overlay=W-w-10:H-h-10", // Apply watermark overlay
			"-q:v", "75",
			outputFilePath,
		}
	}

	// Execute the ffmpeg command
	cmd = exec.Command(ffmpegPath, args...)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add watermark: %v, output: %s", err, string(outputBytes))
	}

	// Clean up the temporary watermark PNG file
	if err := os.Remove(tempWatermarkPNG); err != nil {
		log.Printf("Failed to clean up watermark PNG: %v", err)
	}
	return nil
}

// LoadWatermark loads the watermark SVG content
func (ip *ImageProcessor) LoadWatermark() (string, error) {
	path := ip.Config.WatermarkPath
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read watermark file %s: %v", path, err)
	}
	return string(data), nil
}

// Worker function to process files
func (ip *ImageProcessor) worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range jobs {
		log.Printf("Worker %d: Processing file %s", id, file)
		if err := ip.ProcessFile(file); err != nil {
			log.Printf("Worker %d: Failed to process file %s: %v", id, file, err)
		} else {
			log.Printf("Worker %d: Successfully processed file: %s", id, file)
		}
	}
}

// ProcessFile processes a single file
func (ip *ImageProcessor) ProcessFile(file string) error {
	sourceDir := filepath.Join(ip.Config.WorkDir, "source", ip.UUID)
	intermediateDir := filepath.Join(ip.Config.WorkDir, "intermediate", ip.UUID)
	targetDir := filepath.Join(ip.Config.OutputDir, ip.UUID)

	// Ensure directories exist
	if err := os.MkdirAll(sourceDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create source directory: %v", err)
	}

	if err := os.MkdirAll(intermediateDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create intermediate directory: %v", err)
	}

	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	originalFileName := filepath.Base(file)
	sanitizedFileName := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
	sourceFilePath := filepath.Join(sourceDir, originalFileName)

	// Handle local vs. remote file retrieval
	if ip.Config.IsLocal {
		inputFile := filepath.Join(ip.Config.LocalSource, file)
		if err := ip.CopyFile(inputFile, sourceFilePath); err != nil {
			return fmt.Errorf("failed to copy local file %s to %s: %v", inputFile, sourceFilePath, err)
		}
	} else {
		svc := s3.New(ip.Session)
		s3Object, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(ip.Config.SourceBucket),
			Key:    aws.String(file),
		})
		if err != nil {
			return fmt.Errorf("failed to download file from S3: %v", err)
		}
		defer s3Object.Body.Close()

		outFile, err := os.Create(sourceFilePath)
		if err != nil {
			return fmt.Errorf("failed to create local file %s: %v", sourceFilePath, err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, s3Object.Body); err != nil {
			return fmt.Errorf("failed to copy S3 file to local: %v", err)
		}
	}

	for suffix, size := range ip.Config.WebpSizes {
		intermediateFilePath := filepath.Join(intermediateDir, fmt.Sprintf("%s-%s.png", sanitizedFileName, suffix))
		unwatermarkedIntermediateFilePath := intermediateFilePath

		// Convert SVG to PNG for this size
		if err := ip.ConvertSVGToPNG(sourceFilePath, intermediateFilePath, size); err != nil {
			return fmt.Errorf("SVG to PNG conversion failed for %s: %v", sourceFilePath, err)
		}

		targetFilePath := filepath.Join(targetDir, fmt.Sprintf("%s-%s.webp", sanitizedFileName, suffix))

		// If the size key is "watermark", add a watermark before creating the WebP file
		if suffix == "watermark" {
			watermarkedIntermediateFile := filepath.Join(intermediateDir, fmt.Sprintf("%s-%s-watermarked.png", sanitizedFileName, suffix))
			if err := ip.Watermark(intermediateFilePath, watermarkedIntermediateFile, size); err != nil {
				return fmt.Errorf("watermark creation failed for %s: %v", intermediateFilePath, err)
			}

			// Use the watermarked intermediate file for WebP conversion
			intermediateFilePath = watermarkedIntermediateFile
		}

		// Convert PNG (or watermarked PNG) to WebP
		if err := ip.RunFFmpeg(intermediateFilePath, targetFilePath); err != nil {
			return fmt.Errorf("WebP conversion failed for %s: %v", intermediateFilePath, err)
		}

		// Upload the WebP to S3 if not running locally
		if !ip.Config.IsLocal {
			svc := s3.New(ip.Session)
			outputKey := strings.Replace(file, ".svg", fmt.Sprintf("-%s.webp", suffix), 1)
			webpFile, err := os.Open(targetFilePath)
			if err != nil {
				return fmt.Errorf("failed to open WebP file for upload: %v", err)
			}
			defer webpFile.Close()

			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(ip.Config.TargetBucket),
				Key:         aws.String(outputKey),
				Body:        webpFile,
				ContentType: aws.String("image/webp"),
			})
			if err != nil {
				return fmt.Errorf("failed to upload WebP to S3: %v", err)
			}
		}

		// Cleanup intermediate files immediately after processing
		if ip.Config.AutoCleanup {
			os.Remove(intermediateFilePath)
			if suffix == "watermark" {
				os.Remove(filepath.Join(intermediateDir, fmt.Sprintf("%s-%s-watermarked.png", sanitizedFileName, suffix)))
			}
			os.Remove(unwatermarkedIntermediateFilePath)
		}
	}

	// Cleanup the source file for this specific file
	if ip.Config.AutoCleanup {
		os.Remove(sourceFilePath)
	}

	return nil
}

// RenameFolderWithTimestamp renames a folder with a timestamp
func (ip *ImageProcessor) RenameFolderWithTimestamp(folderPath string) error {
	// Get the absolute path of the folder
	absFolderPath, err := filepath.Abs(folderPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if the folder exists
	if _, err := os.Stat(absFolderPath); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", absFolderPath)
	}

	// Get the parent directory
	parentDir := filepath.Dir(absFolderPath)

	// Generate the timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")

	// Create the new folder name
	newFolderName := filepath.Join(parentDir, timestamp)

	// Rename the folder
	if err := os.Rename(absFolderPath, newFolderName); err != nil {
		return fmt.Errorf("failed to rename folder: %v", err)
	}

	fmt.Printf("Folder renamed from %s to %s\n", absFolderPath, newFolderName)
	return nil
}

// ProcessFiles processes all files
func (ip *ImageProcessor) ProcessFiles() error {
	files, err := ip.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	log.Printf("Files to process: %v", files)

	// Create a channel to queue up files for processing
	jobs := make(chan string, len(files))

	// Create a wait group to synchronize the workers
	var wg sync.WaitGroup

	// Start the workers
	for i := 1; i <= ip.Config.WorkerPoolSize; i++ {
		wg.Add(1)
		go ip.worker(i, jobs, &wg)
	}

	// Queue up the files for processing
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()

	return nil
}

func (ip *ImageProcessor) downloadWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range ip.DownloadQueue {
		log.Printf("Download Worker %d: Downloading file %s", id, file)

		localPath, err := ip.downloadFile(file) // Implement download logic
		if err != nil {
			log.Printf("Download Worker %d: Failed to download %s: %v", id, file, err)
			continue
		}

		// Add the downloaded file to the process queue
		ip.ProcessQueue <- localPath
		log.Printf("Download Worker %d: File %s added to processing queue", id, file)
	}

	log.Printf("Download Worker %d: Finished", id)
}

func (ip *ImageProcessor) downloadFile(file string) (string, error) {
	localPath := filepath.Join(ip.Config.WorkDir, "source", ip.UUID, filepath.Base(file))

	if ip.Config.IsLocal {
		// Copy from local source
		srcPath := filepath.Join(ip.Config.LocalSource, file)
		if err := ip.CopyFile(srcPath, localPath); err != nil {
			return "", fmt.Errorf("failed to copy local file: %v", err)
		}
	} else {
		// Download from S3
		svc := s3.New(ip.Session)
		s3Object, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(ip.Config.SourceBucket),
			Key:    aws.String(file),
		})
		if err != nil {
			return "", fmt.Errorf("failed to download file from S3: %v", err)
		}
		defer s3Object.Body.Close()

		outFile, err := os.Create(localPath)
		if err != nil {
			return "", fmt.Errorf("failed to create local file: %v", err)
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, s3Object.Body); err != nil {
			return "", fmt.Errorf("failed to copy S3 file to local: %v", err)
		}
	}

	return localPath, nil
}

func (ip *ImageProcessor) processWorker(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	for file := range ip.ProcessQueue {
		log.Printf("Processing Worker %d: Processing file %s", id, file)

		if err := ip.ProcessFile(file); err != nil {
			log.Printf("Processing Worker %d: Failed to process file %s: %v", id, file, err)
		} else {
			log.Printf("Processing Worker %d: Successfully processed file: %s", id, file)
		}
	}

	log.Printf("Processing Worker %d: Finished", id)
}

// Cleanup removes temporary directories and files
func (ip *ImageProcessor) Cleanup() {
	// Clean up empty directories.
	if ip.Config.AutoCleanup {
		os.Remove(filepath.Join(ip.Config.WorkDir, "source", ip.UUID))
		os.Remove(filepath.Join(ip.Config.WorkDir, "intermediate", ip.UUID))

		// Sleep 1 second
		time.Sleep(1 * time.Second)

		// Remove the directories if they are empty
		os.RemoveAll(ip.Config.WorkDir)
	}
}

// Run starts the image processing
func (ip *ImageProcessor) Run() error {
	if err := ip.SetupLogging(); err != nil {
		return fmt.Errorf("failed to set up logging: %v", err)
	}

	files, err := ip.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %v", err)
	}

	log.Printf("Files to process: %v", files)

	// Initialize queues
	ip.DownloadQueue = make(chan string, len(files))
	ip.ProcessQueue = make(chan string, len(files))

	// WaitGroups for synchronization
	var downloadWG sync.WaitGroup
	var processWG sync.WaitGroup

	// Start download workers
	for i := 0; i < ip.Config.DownloadWorkerPoolSize; i++ {
		downloadWG.Add(1)
		go ip.downloadWorker(i, &downloadWG)
	}

	// Start processing workers
	for i := 0; i < ip.Config.ProcessWorkerPoolSize; i++ {
		processWG.Add(1)
		go ip.processWorker(i, &processWG)
	}

	// Queue up files for downloading
	for _, file := range files {
		ip.DownloadQueue <- file
	}
	close(ip.DownloadQueue)

	// Wait for all downloads to complete
	downloadWG.Wait()

	// Close the process queue to signal no more files for processing
	close(ip.ProcessQueue)

	// Wait for all processing to complete
	processWG.Wait()

	// Cleanup temporary files and directories
	ip.Cleanup()

	return nil
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

// Creates new ImageProcessor instance
func NewImageProcessor(configFile string) *ImageProcessor {
	var sess *session.Session

	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
		return nil
	}

	roleArn := config.RoleArn
	region := config.Region

	if roleArn == "" {
		log.Println("RoleArn is required in the config file")
		return nil
	}

	if config.IsLocal {
		log.Println("Running in local mode")
	} else {
		sess, err = SessionWithRole(roleArn, region)
		if err != nil {
			log.Fatalf("Failed to create session with role: %v", err)
		}
		log.Println("Successfully assumed role and created session")
	}

	imageProcessor := &ImageProcessor{
		Config:  config,
		UUID:    uuid.New().String(),
		Session: sess,
	}

	return imageProcessor
}
