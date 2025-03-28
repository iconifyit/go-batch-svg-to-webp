package imageprocessor

// import (
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"
// 	"os/exec"
// 	"path/filepath"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/credentials"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/s3"
// 	"github.com/aws/aws-sdk-go/service/sts"
// 	"github.com/google/uuid"
// 	"gopkg.in/yaml.v2"

// 	imagefile "image-processor/src/image-file"
// )

// var imageFile *imagefile.ImageFile

// type ImageProcessor struct {
// 	UUID          string
// 	Config        *Config
// 	Session       *session.Session
// 	DownloadQueue chan string
// 	ProcessQueue  chan string
// }

// // Creates new ImageProcessor instance
// func NewImageProcessor(configFile string) *ImageProcessor {
// 	var sess *session.Session

// 	config, err := LoadConfig(configFile)
// 	if err != nil {
// 		log.Fatalf("Failed to load config: %v", err)
// 		return nil
// 	}

// 	roleArn := config.RoleArn
// 	region := config.Region

// 	if roleArn == "" {
// 		log.Println("RoleArn is required in the config file")
// 		return nil
// 	}

// 	if config.IsLocal {
// 		log.Println("Running in local mode")
// 	} else {
// 		sess, err = SessionWithRole(roleArn, region)
// 		if err != nil {
// 			log.Fatalf("Failed to create session with role: %v", err)
// 		}
// 		log.Println("Successfully assumed role and created session")
// 	}

// 	imageProcessor := &ImageProcessor{
// 		Config:  config,
// 		UUID:    uuid.New().String(),
// 		Session: sess,
// 	}

// 	return imageProcessor
// }

// // SessionWithRole creates a new session with the specified role
// func SessionWithRole(roleArn string, region string) (*session.Session, error) {
// 	// Start a base session (using your current user or environment credentials)
// 	baseSession, err := session.NewSession(&aws.Config{
// 		Region: aws.String(region),
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create an STS client
// 	stsClient := sts.New(baseSession)

// 	// Assume the role
// 	assumeRoleInput := &sts.AssumeRoleInput{
// 		RoleArn:         aws.String(roleArn),
// 		RoleSessionName: aws.String("svg-webp-app-session"),
// 	}

// 	assumeRoleOutput, err := stsClient.AssumeRole(assumeRoleInput)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Create a new session with the temporary credentials
// 	tempSession, err := session.NewSession(&aws.Config{
// 		Region: aws.String(region),
// 		Credentials: credentials.NewStaticCredentials(
// 			*assumeRoleOutput.Credentials.AccessKeyId,
// 			*assumeRoleOutput.Credentials.SecretAccessKey,
// 			*assumeRoleOutput.Credentials.SessionToken,
// 		),
// 	})
// 	if err != nil {
// 		log.Fatalf("Failed to create session with assumed role credentials: %v", err)
// 	}

// 	return tempSession, nil
// }

// // LoadConfig reads the configuration from the provided path.
// func LoadConfig(path string) (*Config, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	config := &Config{}
// 	if err := yaml.NewDecoder(file).Decode(config); err != nil {
// 		return nil, err
// 	}

// 	return config, nil
// }

// // SetupLogging configures logging based on the provided setting
// // SetupLogging configures logging based on the provided setting
// func (ip *ImageProcessor) SetupLogging() error {
// 	logSetting := ip.Config.LoggingOutput
// 	logFilePath := ip.Config.Logfile

// 	// Set log flags to include file and line number
// 	log.SetFlags(log.LstdFlags | log.Lshortfile)

// 	// Remove the existing log file
// 	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
// 		return fmt.Errorf("failed to remove existing log file: %v", err)
// 	}

// 	switch logSetting {
// 	case 0: // No output
// 		log.SetOutput(io.Discard)
// 	case 1: // Log to console
// 		log.SetOutput(os.Stdout)
// 	case 2: // Log to file
// 		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 		if err != nil {
// 			return fmt.Errorf("failed to open log file: %v", err)
// 		}
// 		log.SetOutput(logFile)
// 	case 3: // Log to both console and file
// 		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 		if err != nil {
// 			return fmt.Errorf("failed to open log file: %v", err)
// 		}
// 		multiWriter := io.MultiWriter(os.Stdout, logFile)
// 		log.SetOutput(multiWriter)
// 	default:
// 		return fmt.Errorf("invalid log setting: %d", logSetting)
// 	}
// 	return nil
// }

// // ShouldInclude checks if a file should be included based on the include and exclude lists
// func (ip *ImageProcessor) ShouldInclude(filePath *string) bool {
// 	include := ip.Config.Include
// 	exclude := ip.Config.Exclude

// 	if filePath == nil {
// 		return false
// 	}

// 	log.Printf("ShouldInclude file: %s", *filePath)
// 	log.Printf("Include: %v", include)
// 	log.Printf("Exclude: %v\n", exclude)

// 	fileName := filepath.Base(*filePath)

// 	// Exclude hidden files (e.g., .DS_Store or files starting with '.')
// 	if strings.HasPrefix(fileName, ".") {
// 		return false
// 	}

// 	// Check for exclusion
// 	for _, prefix := range exclude {
// 		log.Printf("Checking prefix: %s - %s", prefix, *filePath)
// 		if strings.HasPrefix(*filePath, prefix) {
// 			return false // Exclude the file if it matches any prefix in `exclude`
// 		}
// 	}

// 	// If include list is empty, include all files (except excluded ones)
// 	if len(include) == 0 {
// 		return true
// 	}

// 	// Check for inclusion
// 	for _, prefix := range include {
// 		if strings.HasPrefix(*filePath, prefix) {
// 			return true
// 		}
// 	}

// 	// If the file does not match any inclusion rule, exclude it
// 	return false
// }

// // ListFiles lists files either locally or from S3
// // func (ip *ImageProcessor) ListFiles() ([]string, error) {
// // 	var files []string

// // 	if ip.Config.IsLocal {
// // 		err := filepath.Walk(ip.Config.LocalSource, func(path string, info os.FileInfo, err error) error {
// // 			if err != nil {
// // 				return err
// // 			}
// // 			if !info.IsDir() {
// // 				relPath, _ := filepath.Rel(ip.Config.LocalSource, path)
// // 				if ip.ShouldInclude(&relPath) {
// // 					files = append(files, relPath)
// // 				}
// // 			}
// // 			return nil
// // 		})
// // 		return files, err
// // 	}

// // 	svc := s3.New(ip.Session)
// // 	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
// // 		Bucket: aws.String(ip.Config.SourceBucket),
// // 	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
// // 		for _, obj := range page.Contents {
// // 			log.Printf("\nObject Key: %s", *obj.Key)
// // 			if ip.ShouldInclude(obj.Key) {
// // 				files = append(files, *obj.Key)
// // 			}
// // 		}
// // 		return true
// // 	})

// // 	return files, err
// // }

// func (ip *ImageProcessor) ListFiles() ([]string, error) {
// 	var files []string

// 	if ip.Config.IsLocal {
// 		// Output absolute paths for local files
// 		err := filepath.Walk(ip.Config.LocalSource, func(path string, info os.FileInfo, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			if !info.IsDir() {
// 				if ip.ShouldInclude(&path) {
// 					imageFile, err := imagefile.NewImageFile(
// 						map[string]string{
// 							"work_dir":   ip.Config.WorkDir,
// 							"output_dir": ip.Config.OutputDir,
// 							"uuid":       ip.UUID,
// 							"is_local":   fmt.Sprint(ip.Config.IsLocal),
// 						},
// 						"",
// 						path,
// 					)
// 					// Check if error is of type imagefile.ContributorNotFoundError
// 					if err != nil {
// 						log.Printf("Error: %+v", err)
// 					} else {
// 						log.Printf("ImageFile: %+v", imageFile)
// 						files = append(files, path) // Absolute paths
// 					}
// 				}
// 			}
// 			return nil
// 		})
// 		return files, err
// 	}

// 	// For S3, use object keys as relative paths
// 	svc := s3.New(ip.Session)
// 	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
// 		Bucket: aws.String(ip.Config.SourceBucket),
// 	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
// 		for _, obj := range page.Contents {
// 			log.Printf("\nObject Key: %s", *obj.Key)
// 			if ip.ShouldInclude(obj.Key) {
// 				files = append(files, *obj.Key)
// 			}
// 		}
// 		return true
// 	})

// 	return files, err
// }

// // listContributors lists contributors either locally or from S3
// func (ip *ImageProcessor) ListContributors() ([]string, error) {
// 	var contributors []string

// 	if ip.Config.IsLocal {
// 		entries, err := os.ReadDir(ip.Config.LocalSource)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to read local source directory: %v", err)
// 		}
// 		for _, entry := range entries {
// 			if entry.IsDir() {
// 				contributors = append(contributors, entry.Name())
// 			}
// 		}
// 	} else {
// 		svc := s3.New(ip.Session)
// 		resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
// 			Bucket:    aws.String(ip.Config.SourceBucket),
// 			Delimiter: aws.String("/"),
// 		})
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to list contributors from S3: %v", err)
// 		}
// 		for _, prefix := range resp.CommonPrefixes {
// 			contributors = append(contributors, aws.StringValue(prefix.Prefix))
// 		}
// 	}

// 	return contributors, nil
// }

// func (ip *ImageProcessor) ListFilesForContributor(contributor string) ([]string, error) {
// 	var files []string
// 	contributorPath := filepath.Join(ip.Config.LocalSource, contributor)

// 	if ip.Config.IsLocal {
// 		err := filepath.Walk(contributorPath, func(path string, info os.FileInfo, err error) error {
// 			if err != nil {
// 				return err
// 			}
// 			if !info.IsDir() {
// 				relPath, err := filepath.Rel(contributorPath, path)
// 				if err != nil {
// 					return fmt.Errorf("failed to compute relative path: %v", err)
// 				}
// 				files = append(files, relPath)
// 			}
// 			return nil
// 		})
// 		return files, err
// 	}

// 	svc := s3.New(ip.Session)
// 	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
// 		Bucket: aws.String(ip.Config.SourceBucket),
// 		Prefix: aws.String(contributor + "/"),
// 	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
// 		for _, obj := range page.Contents {
// 			relPath := strings.TrimPrefix(aws.StringValue(obj.Key), contributor+"/")
// 			files = append(files, relPath)
// 		}
// 		return true
// 	})

// 	return files, err
// }

// // CopyFile copies a file from src to dst
// func (ip *ImageProcessor) CopyFile(src, dst string) error {
// 	in, err := os.Open(src)
// 	if err != nil {
// 		return err
// 	}
// 	defer in.Close()

// 	out, err := os.Create(dst)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()

// 	_, err = io.Copy(out, in)
// 	return err
// }

// // ConvertSVGToPNG converts an SVG file to a PNG file
// func (ip *ImageProcessor) ConvertSVGToPNG(svgPath, pngPath string, size int) error {
// 	cmd := exec.Command(
// 		"rsvg-convert",
// 		"--background-color",
// 		"white",
// 		"-w",
// 		fmt.Sprint(size),
// 		"-h",
// 		fmt.Sprint(size),
// 		svgPath,
// 		"-o",
// 		pngPath,
// 	)
// 	return cmd.Run()
// }

// // RunFFmpeg runs an FFmpeg command to convert a video file to WebP
// func (ip *ImageProcessor) RunFFmpeg(input, output string) error {
// 	var args []string

// 	if ip.Config.UseHardwareAcceleration {
// 		log.Printf("\nUsing hardware acceleration: ffmpeg -hwaccel videotoolbox -i %s -vf format=yuv420p -q:v 75 %s\n", input, output)
// 		args = []string{
// 			"-hwaccel", "videotoolbox", // Use VideoToolbox for hardware acceleration
// 			"-i", input,
// 			"-vf", "format=yuv420p",
// 			"-q:v", "75",
// 			output,
// 		}
// 	} else {
// 		log.Printf("\nUsing software encoding: ffmpeg -i %s -vf format=yuv420p -q:v 75 %s\n", input, output)
// 		args = []string{
// 			"-i", input,
// 			"-vf", "format=yuv420p",
// 			"-q:v", "75",
// 			output,
// 		}
// 	}

// 	cmd := exec.Command(ip.Config.FFmpegPath, args...)
// 	outputBytes, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("ffmpeg command failed: %v, output: %s", err, string(outputBytes))
// 	}

// 	return nil
// }

// // Watermark adds a watermark to an image
// func (ip *ImageProcessor) Watermark(inputFilePath, outputFilePath string, size int) error {
// 	ffmpegPath := ip.Config.FFmpegPath
// 	watermarkSVGPath := ip.Config.WatermarkPath

// 	// Create a unique temporary file for the watermark PNG
// 	tempDir := filepath.Dir(ip.Config.WorkDir)
// 	tempWatermarkPNG := filepath.Join(tempDir, fmt.Sprintf("watermark-%s.png", uuid.New().String()))

// 	log.Printf("Converting watermark SVG to PNG: %s -> %s\n", watermarkSVGPath, tempWatermarkPNG)

// 	// Convert watermark SVG to PNG
// 	cmd := exec.Command("rsvg-convert", "-w", fmt.Sprint(size), "-h", fmt.Sprint(size), "-o", tempWatermarkPNG, watermarkSVGPath)
// 	if outputBytes, err := cmd.CombinedOutput(); err != nil {
// 		return fmt.Errorf("failed to convert watermark SVG to PNG: %v, output: %s", err, string(outputBytes))
// 	}

// 	// Build the ffmpeg command
// 	var args []string
// 	if ip.Config.UseHardwareAcceleration {
// 		log.Printf("Using hardware acceleration for watermark: ffmpeg -hwaccel videotoolbox -i %s -i %s -filter_complex overlay=W-w-10:H-h-10 -q:v 75 %s\n", inputFilePath, tempWatermarkPNG, outputFilePath)
// 		args = []string{
// 			"-hwaccel", "videotoolbox", // Apply hardware acceleration for decoding
// 			"-i", inputFilePath,
// 			"-i", tempWatermarkPNG,
// 			"-filter_complex", "overlay=W-w-10:H-h-10", // Apply watermark overlay
// 			"-q:v", "75",
// 			outputFilePath,
// 		}
// 	} else {
// 		log.Printf("Using software encoding for watermark: ffmpeg -i %s -i %s -filter_complex overlay=W-w-10:H-h-10 -q:v 75 %s\n", inputFilePath, tempWatermarkPNG, outputFilePath)
// 		args = []string{
// 			"-i", inputFilePath,
// 			"-i", tempWatermarkPNG,
// 			"-filter_complex", "overlay=W-w-10:H-h-10", // Apply watermark overlay
// 			"-q:v", "75",
// 			outputFilePath,
// 		}
// 	}

// 	// Execute the ffmpeg command
// 	cmd = exec.Command(ffmpegPath, args...)
// 	outputBytes, err := cmd.CombinedOutput()
// 	if err != nil {
// 		return fmt.Errorf("failed to add watermark: %v, output: %s", err, string(outputBytes))
// 	}

// 	// Clean up the temporary watermark PNG file
// 	if err := os.Remove(tempWatermarkPNG); err != nil {
// 		log.Printf("Failed to clean up watermark PNG: %v", err)
// 	}
// 	return nil
// }

// // LoadWatermark loads the watermark SVG content
// func (ip *ImageProcessor) LoadWatermark() (string, error) {
// 	path := ip.Config.WatermarkPath
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read watermark file %s: %v", path, err)
// 	}
// 	return string(data), nil
// }

// // Worker function to process files
// func (ip *ImageProcessor) worker(id int, jobs <-chan string, wg *sync.WaitGroup, errorChan chan<- error) {
// 	defer wg.Done()

// 	for file := range jobs {
// 		log.Printf("Worker %d: Processing file %s", id, file)

// 		// Process the file
// 		if err := ip.ProcessFile(file); err != nil {
// 			// Send the error to the error channel and stop this worker
// 			errorChan <- fmt.Errorf("worker %d: error processing file %s: %v", id, file, err)
// 			return
// 		}

// 		log.Printf("Worker %d: Successfully processed file: %s", id, file)
// 	}
// }

// // ProcessFile processes a single file
// func (ip *ImageProcessor) ProcessFile(file string) error {
// 	// Base directories for source, intermediate, and target
// 	sourceDir := filepath.Join(ip.Config.WorkDir, "source", ip.UUID)
// 	intermediateDir := filepath.Join(ip.Config.WorkDir, "intermediate", ip.UUID)
// 	targetDir := filepath.Join(ip.Config.OutputDir, ip.UUID)

// 	// Ensure base directories exist
// 	if err := os.MkdirAll(sourceDir, os.ModePerm); err != nil {
// 		return fmt.Errorf("failed to create source directory: %v", err)
// 	}
// 	if err := os.MkdirAll(intermediateDir, os.ModePerm); err != nil {
// 		return fmt.Errorf("failed to create intermediate directory: %v", err)
// 	}
// 	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
// 		return fmt.Errorf("failed to create target directory: %v", err)
// 	}

// 	// Construct paths using the relative `file`
// 	sourceFilePath := filepath.Join(sourceDir, file)
// 	intermediateFilePath := filepath.Join(intermediateDir, file)

// 	// Ensure directories for the file's relative path exist
// 	if err := os.MkdirAll(filepath.Dir(sourceFilePath), os.ModePerm); err != nil {
// 		return fmt.Errorf("failed to create source subdirectory: %v", err)
// 	}
// 	if err := os.MkdirAll(filepath.Dir(intermediateFilePath), os.ModePerm); err != nil {
// 		return fmt.Errorf("failed to create intermediate subdirectory: %v", err)
// 	}

// 	// Handle local file copying or S3 download
// 	if ip.Config.IsLocal {
// 		inputFile := filepath.Join(ip.Config.LocalSource, file)
// 		if err := ip.CopyFile(inputFile, sourceFilePath); err != nil {
// 			return fmt.Errorf("failed to copy local file %s to %s: %v", inputFile, sourceFilePath, err)
// 		}
// 	} else {
// 		svc := s3.New(ip.Session)
// 		s3Object, err := svc.GetObject(&s3.GetObjectInput{
// 			Bucket: aws.String(ip.Config.SourceBucket),
// 			Key:    aws.String(file),
// 		})
// 		if err != nil {
// 			return fmt.Errorf("failed to download file from S3: %v", err)
// 		}
// 		defer s3Object.Body.Close()

// 		outFile, err := os.Create(sourceFilePath)
// 		if err != nil {
// 			return fmt.Errorf("failed to create local file %s: %v", sourceFilePath, err)
// 		}
// 		defer outFile.Close()

// 		if _, err := io.Copy(outFile, s3Object.Body); err != nil {
// 			return fmt.Errorf("failed to copy S3 file to local: %v", err)
// 		}
// 	}

// 	// Process each size for WebP conversion
// 	for suffix, size := range ip.Config.WebpSizes {
// 		intermediatePNGPath := fmt.Sprintf("%s-%s.png", intermediateFilePath, suffix)
// 		unwatermarkedPNGPath := intermediatePNGPath

// 		// Convert SVG to PNG
// 		if err := ip.ConvertSVGToPNG(sourceFilePath, intermediatePNGPath, size); err != nil {
// 			return fmt.Errorf("SVG to PNG conversion failed for %s: %v", sourceFilePath, err)
// 		}

// 		// Construct target WebP file path
// 		targetRelativePath := strings.Replace(file, ".svg", "", -1)
// 		targetFilePath := filepath.Join(targetDir, fmt.Sprintf("%s-%s.webp", targetRelativePath, suffix))

// 		// Add watermark if necessary
// 		if suffix == "watermark" {
// 			watermarkedIntermediateFile := fmt.Sprintf("%s-watermarked.png", intermediateFilePath)
// 			if err := ip.Watermark(intermediatePNGPath, watermarkedIntermediateFile, size); err != nil {
// 				return fmt.Errorf("watermark creation failed for %s: %v", intermediatePNGPath, err)
// 			}
// 			intermediatePNGPath = watermarkedIntermediateFile
// 		}

// 		// Convert PNG to WebP
// 		if err := ip.RunFFmpeg(intermediatePNGPath, targetFilePath); err != nil {
// 			return fmt.Errorf("WebP conversion failed for %s: %v", intermediatePNGPath, err)
// 		}

// 		// Upload WebP to S3 if not local
// 		if !ip.Config.IsLocal {
// 			svc := s3.New(ip.Session)
// 			outputKey := strings.Replace(file, ".svg", fmt.Sprintf("-%s.webp", suffix), 1)
// 			webpFile, err := os.Open(targetFilePath)
// 			if err != nil {
// 				return fmt.Errorf("failed to open WebP file for upload: %v", err)
// 			}
// 			defer webpFile.Close()

// 			_, err = svc.PutObject(&s3.PutObjectInput{
// 				Bucket:      aws.String(ip.Config.TargetBucket),
// 				Key:         aws.String(outputKey),
// 				Body:        webpFile,
// 				ContentType: aws.String("image/webp"),
// 			})
// 			if err != nil {
// 				return fmt.Errorf("failed to upload WebP to S3: %v", err)
// 			}
// 		}

// 		// Cleanup intermediate files
// 		if ip.Config.AutoCleanup {
// 			os.Remove(intermediatePNGPath)
// 			if suffix == "watermark" {
// 				os.Remove(fmt.Sprintf("%s-watermarked.png", intermediateFilePath))
// 			}
// 			os.Remove(unwatermarkedPNGPath)
// 		}
// 	}

// 	// Cleanup source file
// 	if ip.Config.AutoCleanup {
// 		os.Remove(sourceFilePath)
// 	}

// 	return nil
// }

// // RenameFolderWithTimestamp renames a folder with a timestamp
// func (ip *ImageProcessor) RenameFolderWithTimestamp(folderPath string) error {
// 	// Get the absolute path of the folder
// 	absFolderPath, err := filepath.Abs(folderPath)
// 	if err != nil {
// 		return fmt.Errorf("failed to get absolute path: %v", err)
// 	}

// 	// Check if the folder exists
// 	if _, err := os.Stat(absFolderPath); os.IsNotExist(err) {
// 		return fmt.Errorf("folder does not exist: %s", absFolderPath)
// 	}

// 	// Get the parent directory
// 	parentDir := filepath.Dir(absFolderPath)

// 	// Generate the timestamp
// 	timestamp := time.Now().Format("2006-01-02-15-04-05")

// 	// Create the new folder name
// 	newFolderName := filepath.Join(parentDir, timestamp)

// 	// Rename the folder
// 	if err := os.Rename(absFolderPath, newFolderName); err != nil {
// 		return fmt.Errorf("failed to rename folder: %v", err)
// 	}

// 	fmt.Printf("Folder renamed from %s to %s\n", absFolderPath, newFolderName)
// 	return nil
// }

// // ProcessFiles processes all files
// func (ip *ImageProcessor) ProcessFiles() error {
// 	files, err := ip.ListFiles()
// 	if err != nil {
// 		return fmt.Errorf("failed to list files: %v", err)
// 	}

// 	log.Printf("Files to process: %v", files)

// 	// Create a channel to queue up files for processing
// 	jobs := make(chan string, len(files))
// 	errorChan := make(chan error, 1) // Buffer size 1 for fail-fast error handling

// 	// Create a wait group to synchronize the workers
// 	var wg sync.WaitGroup

// 	// Start the workers
// 	for i := 1; i <= ip.Config.WorkerPoolSize; i++ {
// 		wg.Add(1)
// 		go ip.worker(i, jobs, &wg, errorChan) // Call worker with all required arguments
// 	}

// 	// Queue up the files for processing in a separate goroutine
// 	go func() {
// 		for _, file := range files {
// 			select {
// 			case jobs <- file:
// 				// File added to jobs queue
// 			case <-errorChan:
// 				// Stop queuing files if an error is encountered
// 				close(jobs)
// 				return
// 			}
// 		}
// 		close(jobs) // Close jobs channel when all files are queued
// 	}()

// 	// Wait for workers in a separate goroutine
// 	done := make(chan struct{})
// 	go func() {
// 		wg.Wait()
// 		close(done) // Signal that all workers are done
// 	}()

// 	// Monitor for errors or completion
// 	select {
// 	case err := <-errorChan:
// 		// A worker encountered a critical error
// 		log.Printf("Critical error encountered: %v", err)
// 		close(jobs) // Ensure no further jobs are processed
// 		return err
// 	case <-done:
// 		// All workers completed successfully
// 		log.Println("All workers completed successfully.")
// 		return nil
// 	}
// }

// func (ip *ImageProcessor) downloadWorker(id int, errorChan chan<- error) {
// 	defer log.Printf("Download Worker %d: Finished", id)

// 	for file := range ip.DownloadQueue {
// 		log.Printf("Download Worker %d: Downloading file %s", id, file)

// 		localPath, err := ip.downloadFile(file) // Implement download logic
// 		if err != nil {
// 			// Send error to errorChan and stop processing
// 			errorChan <- fmt.Errorf("download worker %d: Failed to download %s: %v", id, file, err)
// 			return
// 		}

// 		// Add the downloaded file to the process queue
// 		ip.ProcessQueue <- localPath
// 		log.Printf("Download Worker %d: File %s added to processing queue", id, file)
// 	}
// }

// func (ip *ImageProcessor) downloadFile(file string) (string, error) {
// 	var relativePath string

// 	if ip.Config.IsLocal {
// 		// Compute relative path only once
// 		relativePath, _ = filepath.Rel(ip.Config.LocalSource, file)
// 	} else {
// 		// For S3 files, the relative path is the object key
// 		relativePath = file
// 	}

// 	// Construct the local path in the working directory
// 	localPath := filepath.Join(ip.Config.WorkDir, "source", ip.UUID, relativePath)

// 	// Ensure the directory structure exists for the local path
// 	if err := os.MkdirAll(filepath.Dir(localPath), os.ModePerm); err != nil {
// 		return "", fmt.Errorf("failed to create directory for local file: %v", err)
// 	}

// 	if ip.Config.IsLocal {
// 		// Copy local files
// 		if err := ip.CopyFile(file, localPath); err != nil {
// 			return "", fmt.Errorf("failed to copy local file: %v", err)
// 		}
// 	} else {
// 		// Download from S3
// 		svc := s3.New(ip.Session)
// 		s3Object, err := svc.GetObject(&s3.GetObjectInput{
// 			Bucket: aws.String(ip.Config.SourceBucket),
// 			Key:    aws.String(file),
// 		})
// 		if err != nil {
// 			return "", fmt.Errorf("failed to download file from S3: %v", err)
// 		}
// 		defer s3Object.Body.Close()

// 		outFile, err := os.Create(localPath)
// 		if err != nil {
// 			return "", fmt.Errorf("failed to create local file: %v", err)
// 		}
// 		defer outFile.Close()

// 		if _, err := io.Copy(outFile, s3Object.Body); err != nil {
// 			return "", fmt.Errorf("failed to copy S3 file to local: %v", err)
// 		}
// 	}

// 	return localPath, nil
// }

// func (ip *ImageProcessor) processWorker(id int, errorChan chan<- error) {
// 	defer log.Printf("Processing Worker %d: Finished", id)

// 	for file := range ip.ProcessQueue {
// 		log.Printf("Processing Worker %d: Processing file %s", id, file)

// 		if err := ip.ProcessFile(file); err != nil {
// 			// Send error to errorChan and stop processing
// 			errorChan <- fmt.Errorf("processing Worker %d: Failed to process file %s: %v", id, file, err)
// 			return
// 		}

// 		log.Printf("Processing Worker %d: Successfully processed file: %s", id, file)
// 	}
// }

// // Cleanup removes temporary directories and files
// func (ip *ImageProcessor) Cleanup() {
// 	// Clean up empty directories.
// 	if ip.Config.AutoCleanup {
// 		os.Remove(filepath.Join(ip.Config.WorkDir, "source", ip.UUID))
// 		os.Remove(filepath.Join(ip.Config.WorkDir, "intermediate", ip.UUID))

// 		// Sleep 1 second
// 		time.Sleep(1 * time.Second)

// 		// Remove the directories if they are empty
// 		os.RemoveAll(ip.Config.WorkDir)
// 	}
// }

// // Run starts the image processing
// func (ip *ImageProcessor) Run() error {
// 	if err := ip.SetupLogging(); err != nil {
// 		return fmt.Errorf("failed to set up logging: %v", err)
// 	}

// 	files, err := ip.ListFiles()
// 	if err != nil {
// 		return fmt.Errorf("failed to list files: %v", err)
// 	}

// 	log.Printf("Files to process: %d", len(files))

// 	// Initialize queues
// 	ip.DownloadQueue = make(chan string, len(files))
// 	ip.ProcessQueue = make(chan string, len(files))

// 	// Define error channel for fail-fast mechanism
// 	errorChan := make(chan error, 1) // Buffer size 1 to handle first error

// 	// WaitGroups for synchronization
// 	var downloadWG sync.WaitGroup
// 	var processWG sync.WaitGroup

// 	// Start download workers
// 	for i := 0; i < ip.Config.DownloadWorkerPoolSize; i++ {
// 		downloadWG.Add(1)
// 		go func(workerID int) {
// 			defer downloadWG.Done()
// 			ip.downloadWorker(workerID, errorChan)
// 		}(i)
// 	}

// 	// Start processing workers
// 	for i := 0; i < ip.Config.ProcessWorkerPoolSize; i++ {
// 		processWG.Add(1)
// 		go func(workerID int) {
// 			defer processWG.Done()
// 			ip.processWorker(workerID, errorChan)
// 		}(i)
// 	}

// 	// Queue up files for downloading
// 	go func() {
// 		for _, file := range files {
// 			ip.DownloadQueue <- file
// 		}
// 		close(ip.DownloadQueue)
// 	}()

// 	// Wait for downloads in a separate goroutine
// 	go func() {
// 		downloadWG.Wait()
// 		close(ip.ProcessQueue) // Close process queue only after downloads are done
// 	}()

// 	// Wait for processing workers to complete in a separate goroutine
// 	go func() {
// 		processWG.Wait()
// 		close(errorChan) // Signal no more errors can occur
// 	}()

// 	// Monitor for errors or completion
// 	for err := range errorChan {
// 		if err != nil {
// 			log.Printf("Critical error encountered: %v", err)
// 			return err // Exit immediately on the first error
// 		}
// 	}

// 	// Cleanup temporary files and directories
// 	ip.Cleanup()

// 	log.Println("Image processor completed successfully.")
// 	return nil
// }
