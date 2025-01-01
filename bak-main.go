package main

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
	"gopkg.in/yaml.v2"
	// Correctly import the uuid package
)

// Config holds the YAML configuration structure
type Config struct {
	SourceBucket   string         `yaml:"source_bucket"`
	TargetBucket   string         `yaml:"target_bucket"`
	Include        []string       `yaml:"include_prefixes"`
	Exclude        []string       `yaml:"omit_prefixes"`
	Region         string         `yaml:"aws_region"`
	DryRun         bool           `yaml:"dry_run"`
	FFmpegPath     string         `yaml:"ffmpegPath"`
	IsLocal        bool           `yaml:"is_local"`
	LocalSource    string         `yaml:"local_source"`
	LocalTarget    string         `yaml:"local_target"`
	AutoCleanup    bool           `yaml:"auto_cleanup"`
	WebpSizes      map[string]int `yaml:"webp_sizes"`
	WatermarkPath  string         `yaml:"watermark_path"`
	WorkerPoolSize int            `yaml:"worker_pool_size"`
	RoleArn        string         `yaml:"role_arn"`
	LoggingOutput  int            `yaml:"logging_output"`
	Logfile        string         `yaml:"logfile"`
	WorkDir        string         `yaml:"work_dir"`
}

// LoadConfig loads the configuration from the YAML file
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	if err := yaml.NewDecoder(file).Decode(config); err != nil {
		return nil, err
	}
	return config, nil
}

// SetupLogging configures logging based on the provided setting
func SetupLogging(logSetting int, logFilePath string) error {
	// Remove the existing log file
	if err := os.Remove(logFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing log file: %v", err)
	}

	switch logSetting {
	case 1: // Log to console
		log.SetOutput(os.Stdout)
	case 2: // Log to file
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		log.SetOutput(logFile)
	case 3: // Log to both console and file
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		multiWriter := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(multiWriter)
	default:
		return fmt.Errorf("invalid log setting: %d", logSetting)
	}
	return nil
}

// ShouldInclude checks if a file should be included based on the include and exclude lists
func ShouldInclude(filePath *string, include []string, exclude []string) bool {
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
func ListFiles(sess *session.Session, config *Config) ([]string, error) {
	var files []string

	if config.IsLocal {
		err := filepath.Walk(config.LocalSource, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, _ := filepath.Rel(config.LocalSource, path)
				if ShouldInclude(&relPath, config.Include, config.Exclude) {
					files = append(files, relPath)
				}
			}
			return nil
		})
		return files, err
	}

	svc := s3.New(sess)
	err := svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(config.SourceBucket),
	}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			log.Printf("\nObject Key: %s", *obj.Key)
			if ShouldInclude(obj.Key, config.Include, config.Exclude) {
				files = append(files, *obj.Key)
			}
		}
		return true
	})

	return files, err
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
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
func ConvertSVGToPNG(svgPath, pngPath string, size int) error {
	cmd := exec.Command("rsvg-convert", "--background-color", "white", "-w", fmt.Sprint(size), "-h", fmt.Sprint(size), svgPath, "-o", pngPath)
	return cmd.Run()
}

// RunFFmpeg runs an FFmpeg command to convert a video file to WebP
func RunFFmpeg(ffmpegPath, input, output string) error {
	log.Printf("\nffmpeg -i %s -vf format=yuv420p -q:v 75 %s\n", input, output)
	cmd := exec.Command(ffmpegPath, "-i", input, "-vf", "format=yuv420p", "-q:v", "75", output)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg command failed: %v, output: %s", err, string(outputBytes))
	}
	return nil
}

// Watermark adds a watermark to an image
func Watermark(config *Config, inputFilePath, outputFilePath string, size int) error {
	ffmpegPath := config.FFmpegPath
	watermarkSVGPath := config.WatermarkPath

	// Create a unique temporary file for the watermark PNG
	tempDir := filepath.Dir(config.WorkDir) // watermarkSVGPath
	tempWatermarkPNG := filepath.Join(tempDir, fmt.Sprintf("watermark-%s.png", uuid.New().String()))

	log.Printf("Converting watermark SVG to PNG: %s -> %s\n", watermarkSVGPath, tempWatermarkPNG)
	cmd := exec.Command("rsvg-convert", "-w", fmt.Sprint(size), "-h", fmt.Sprint(size), "-o", tempWatermarkPNG, watermarkSVGPath)
	if outputBytes, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to convert watermark SVG to PNG: %v, output: %s", err, string(outputBytes))
	}

	// Overlay the watermark PNG onto the input image
	cmd = exec.Command(ffmpegPath,
		"-i", inputFilePath,
		"-i", tempWatermarkPNG,
		"-filter_complex", "overlay=W-w-10:H-h-10", // Position the watermark at bottom-right
		"-q:v", "75",
		outputFilePath,
	)
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
func LoadWatermark(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read watermark file %s: %v", path, err)
	}
	return string(data), nil
}

// Worker function to process files
func worker(id int, jobs <-chan string, wg *sync.WaitGroup, sess *session.Session, config *Config, runUUID, watermarkSVG string) {
	defer wg.Done()

	for file := range jobs {
		log.Printf("Worker %d: Processing file %s", id, file)
		if err := ProcessFile(sess, config, runUUID, file, watermarkSVG); err != nil {
			log.Printf("Worker %d: Failed to process file %s: %v", id, file, err)
		} else {
			log.Printf("Worker %d: Successfully processed file: %s", id, file)
		}
	}
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

// ProcessFile processes a single file
func ProcessFile(sess *session.Session, config *Config, runUUID string, file string, watermarkSVG string) error {
	sourceDir := filepath.Join("./tmp/source", runUUID)
	intermediateDir := filepath.Join("./tmp/intermediate", runUUID)
	targetDir := filepath.Join("./test/output", runUUID)

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
	if config.IsLocal {
		inputFile := filepath.Join(config.LocalSource, file)
		if err := CopyFile(inputFile, sourceFilePath); err != nil {
			return fmt.Errorf("failed to copy local file %s to %s: %v", inputFile, sourceFilePath, err)
		}
	} else {
		svc := s3.New(sess)
		s3Object, err := svc.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(config.SourceBucket),
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

	for suffix, size := range config.WebpSizes {
		intermediateFilePath := filepath.Join(intermediateDir, fmt.Sprintf("%s-%s.png", sanitizedFileName, suffix))
		unwatermarkedIntermediateFilePath := intermediateFilePath

		// Convert SVG to PNG for this size
		if err := ConvertSVGToPNG(sourceFilePath, intermediateFilePath, size); err != nil {
			return fmt.Errorf("SVG to PNG conversion failed for %s: %v", sourceFilePath, err)
		}

		targetFilePath := filepath.Join(targetDir, fmt.Sprintf("%s-%s.webp", sanitizedFileName, suffix))

		// If the size key is "watermark", add a watermark before creating the WebP file
		if suffix == "watermark" {
			watermarkedIntermediateFile := filepath.Join(intermediateDir, fmt.Sprintf("%s-%s-watermarked.png", sanitizedFileName, suffix))
			if err := Watermark(config, intermediateFilePath, watermarkedIntermediateFile, size); err != nil {
				return fmt.Errorf("watermark creation failed for %s: %v", intermediateFilePath, err)
			}

			// Use the watermarked intermediate file for WebP conversion
			intermediateFilePath = watermarkedIntermediateFile
		}

		// Convert PNG (or watermarked PNG) to WebP
		if err := RunFFmpeg(config.FFmpegPath, intermediateFilePath, targetFilePath); err != nil {
			return fmt.Errorf("WebP conversion failed for %s: %v", intermediateFilePath, err)
		}

		// Upload the WebP to S3 if not running locally
		if !config.IsLocal {
			svc := s3.New(sess)
			outputKey := strings.Replace(file, ".svg", fmt.Sprintf("-%s.webp", suffix), 1)
			webpFile, err := os.Open(targetFilePath)
			if err != nil {
				return fmt.Errorf("failed to open WebP file for upload: %v", err)
			}
			defer webpFile.Close()

			_, err = svc.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(config.TargetBucket),
				Key:         aws.String(outputKey),
				Body:        webpFile,
				ContentType: aws.String("image/webp"),
			})
			if err != nil {
				return fmt.Errorf("failed to upload WebP to S3: %v", err)
			}
		}

		// Cleanup intermediate files immediately after processing
		if config.AutoCleanup {
			os.Remove(intermediateFilePath)
			if suffix == "watermark" {
				os.Remove(filepath.Join(intermediateDir, fmt.Sprintf("%s-%s-watermarked.png", sanitizedFileName, suffix)))
			}
			os.Remove(unwatermarkedIntermediateFilePath)
		}
	}

	// Cleanup the source file for this specific file
	if config.AutoCleanup {
		os.Remove(sourceFilePath)
	}

	return nil
}

// Main function
func main2() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <config.yml>")
	}

	var sess *session.Session
	var err error

	configFile := os.Args[1]
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config: %+v", config)

	// Setup logging
	if err := SetupLogging(config.LoggingOutput, config.Logfile); err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	runUUID := uuid.New().String()
	log.Printf("runUUID: %s\n", runUUID)

	roleArn := config.RoleArn
	region := config.Region

	if roleArn == "" {
		log.Println("RoleArn is required in the config file")
		return
	}

	if config.IsLocal {
		log.Println("Running in local mode")
	} else {
		sess, err = SessionWithRole(roleArn, region)
		if err != nil {
			log.Fatalf("Failed to create session with role: %v", err)
		}

		// Now use `sess` in your app as the session with assumed role credentials
		log.Println("Successfully assumed role and created session")
	}

	// Load watermark SVG
	watermarkSVG, err := os.ReadFile(config.WatermarkPath)
	if err != nil {
		log.Fatalf("Failed to read watermark file %s: %v", config.WatermarkPath, err)
	}
	watermarkContent := string(watermarkSVG)

	// Fetch files to process
	log.Println("Fetching files to process...")
	files, err := ListFiles(sess, config)
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}

	if len(files) == 0 {
		log.Println("No files to process.")
		return
	}

	log.Printf("Processing %d files with %d workers...", len(files), config.WorkerPoolSize)

	// Create channels for jobs
	jobs := make(chan string, len(files))

	// WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < config.WorkerPoolSize; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg, sess, config, runUUID, watermarkContent)
	}

	// Send jobs to workers
	for _, file := range files {
		jobs <- file
	}
	close(jobs) // Close the channel to signal workers no more jobs

	// Wait for all workers to finish
	wg.Wait()

	// Clean up empty directories.
	if config.AutoCleanup {
		os.Remove(filepath.Join(config.WorkDir, "source", runUUID))
		os.Remove(filepath.Join(config.WorkDir, "intermediate", runUUID))

		// Sleep 1 second
		time.Sleep(1 * time.Second)

		// Remove the directories if they are empty
		os.RemoveAll(config.WorkDir)
	}

	log.Println("Processing complete!")
}
