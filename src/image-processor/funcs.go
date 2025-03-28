package imageprocessor

import (
	"fmt"
	fn "image-processor/src/common"
	fileservice "image-processor/src/file-service"
	imagefile "image-processor/src/image-file"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

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

// ProcessFile processes a single file
func (ip *ImageProcessor) ProcessFile(imgFile imagefile.ImageFile) error {
	log.Println("\n\n------------------------------------------------------")
	log.Printf("ProcessFile - file : %s", imgFile.ObjectKey)
	log.Printf("Processing file: %v", fn.ToJSON(imgFile))

	// n := 0

	// foo := map[string]string{
	// 	"URL":               "/Users/scott/github/vectopus-code/image-processor/test/input/iconify/icons/2C5AC6FCEB8D/6591D6BE211A/icons-icons-handg-gesture-coffee-cup.svg",
	// 	"ObjectKey":         "iconify/icons/2C5AC6FCEB8D/6591D6BE211A/icons-icons-handg-gesture-coffee-cup.svg",
	// 	"InputPath":         "/Users/scott/github/vectopus-code/image-processor/test/input/iconify/icons/2C5AC6FCEB8D/6591D6BE211A/icons-icons-handg-gesture-coffee-cup.svg",
	// 	"Contributor":       "iconify",
	// 	"ProductType":       "icons",
	// 	"FamilyUniqueID":    "2C5AC6FCEB8D",
	// 	"SetUniqueID":       "6591D6BE211A",
	// 	"Filename":          "icons-icons-handg-gesture-coffee-cup.svg",
	// 	"Extension":         "svg",
	// 	"Slug":              "icons-icons-handg-gesture-coffee-cup",
	// 	"Stem":              "icons-icons-handg-gesture-coffee-cup",
	// 	"IsValid":           "true",
	// 	"Error":             "",
	// 	"OptimizedImageKey": "iconify/icons/2C5AC6FCEB8D/6591D6BE211A/icons-icons-handg-gesture-coffee-cup.webp",
	// 	"FileType":          "svg",
	// }

	// log.Printf("foo : %s %d", n, len(foo))

	// file := imgFile.ObjectKey
	// log.Printf("exImageFile : %v", exImageFile)

	// Base directories for source, intermediate, and target
	// sourceDir := filepath.Join(ip.Config.WorkDir, ip.UUID,"source")
	// intermediateDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "intermediate")
	// targetDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "output")

	processSourceDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "source")
	processIntermediateDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "intermediate")
	processOutputDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "output")

	log.Println("\n\n------------------------------------------------------")
	log.Printf("processSourceDir : %s", processSourceDir)
	log.Printf("processIntermediateDir : %s", processIntermediateDir)
	log.Printf("processOutputDir : %s", processOutputDir)

	// processSourceDir : /Volumes/image-processor-ramdisk/a81cb2d2-f46b-4acb-b555-f261d26259c0/source
	// processIntermediateDir : /Volumes/image-processor-ramdisk/a81cb2d2-f46b-4acb-b555-f261d26259c0/intermediate
	// processOutputDir : /Volumes/image-processor-ramdisk/a81cb2d2-f46b-4acb-b555-f261d26259c0/output

	// =========================================================================
	// Ensure base directories exist
	// =========================================================================
	if err := os.MkdirAll(processSourceDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create source directory: %v", err)
	}

	if err := os.MkdirAll(processIntermediateDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create intermediate directory: %v", err)
	}

	if err := os.MkdirAll(processOutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Construct paths using the relative `file`
	sourceFilePath := filepath.Join(processSourceDir, imgFile.ObjectKey)
	intermediateFilePath := filepath.Join(processIntermediateDir, imgFile.ObjectKey)
	webpOutputFilePath := filepath.Join(processOutputDir, imgFile.ObjectKey)

	log.Println("\n\n------------------------------------------------------")
	log.Printf("sourceFilePath : %s", sourceFilePath)
	log.Printf("intermediateFilePath : %s", intermediateFilePath)
	log.Printf("webpOutputFilePath : %s", webpOutputFilePath)

	// // Ensure directories for the file's relative path exist
	if err := os.MkdirAll(filepath.Dir(sourceFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create source subdirectory: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(intermediateFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create intermediate subdirectory: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(webpOutputFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output subdirectory: %v", err)
	}

	// return nil

	// Handle local file copying or S3 download
	// if ip.Config.IsLocal {
	// 	inputFile := filepath.Join(ip.Config.LocalSource, imgFile.ObjectKey)
	// 	if err := ip.CopyFile(inputFile, sourceFilePath); err != nil {
	// 		return fmt.Errorf("failed to copy local file %s to %s: %v", inputFile, sourceFilePath, err)
	// 	}
	// } else {
	// 	svc := s3.New(ip.Session)
	// 	s3Object, err := svc.GetObject(&s3.GetObjectInput{
	// 		Bucket: aws.String(ip.Config.SourceBucket),
	// 		Key:    aws.String(imgFile.ObjectKey),
	// 	})
	// 	if err != nil {
	// 		return fmt.Errorf("failed to download file from S3: %v", err)
	// 	}
	// 	defer s3Object.Body.Close()

	// 	outFile, err := os.Create(sourceFilePath)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to create local file %s: %v", sourceFilePath, err)
	// 	}
	// 	defer outFile.Close()

	// 	if _, err := io.Copy(outFile, s3Object.Body); err != nil {
	// 		return fmt.Errorf("failed to copy S3 file to local: %v", err)
	// 	}
	// }

	// log.Printf("ip.Session : %+v", ip.Session)

	// fsvc := fileservice.NewS3FileService(&fileservice.ServiceInput{
	// 	UUID:    ip.UUID,
	// 	IsLocal: false,
	// 	Session: ip.Session,
	// })

	// log.Printf("fsvc : %s", fn.ToJSON(fsvc))

	// s3Client := s3.New(

	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String(ip.Config.Region),
	})
	// if err != nil {
	// 	return nil, err
	// }

	s3Client := s3.New(sess)

	// // Process each size for WebP conversion
	for suffix, size := range ip.Config.WebpSizes {
		ext := filepath.Ext(imgFile.ObjectKey)
		intermediatePNGPath := fmt.Sprintf("%s-%s.png", intermediateFilePath, suffix)

		// Convert SVG to PNG
		if err := ip.ConvertSVGToPNG(sourceFilePath, intermediatePNGPath, size); err != nil {
			return fmt.Errorf("SVG to PNG conversion failed for %s: %v", sourceFilePath, err)
		}

		// Construct target WebP file path
		targetRelativePath := strings.Replace(imgFile.ObjectKey, ext, "", -1)
		targetWebpFilePath := filepath.Join(processOutputDir, fmt.Sprintf("%s-%s.webp", targetRelativePath, suffix))

		// Add watermark if necessary
		if suffix == "watermark" {
			watermarkedIntermediateFile := fmt.Sprintf("%s-watermarked.png", intermediateFilePath)
			if err := ip.Watermark(intermediatePNGPath, watermarkedIntermediateFile, size); err != nil {
				return fmt.Errorf("watermark creation failed for %s: %v", intermediatePNGPath, err)
			}
			intermediatePNGPath = watermarkedIntermediateFile
		}

		// Convert PNG to WebP
		if err := ip.RunFFmpeg(intermediatePNGPath, targetWebpFilePath); err != nil {
			return fmt.Errorf("WebP conversion failed for %s: %v", intermediatePNGPath, err)
		}

		// Upload WebP to S3 if not local
		outputKey := strings.Replace(imgFile.ObjectKey, ext, fmt.Sprintf("-%s.webp", suffix), 1)

		log.Printf("outputKey : %s", outputKey)
		log.Printf("targetWebpFilePath : %s", targetWebpFilePath)

		if ip.Config.UploadToS3 {
			output, err := s3Client.PutObject(&s3.PutObjectInput{
				Bucket: aws.String(ip.Config.TargetBucket),
				Key:    aws.String(outputKey),
				Body:   aws.ReadSeekCloser(strings.NewReader(targetWebpFilePath)),
			})
			if err != nil {
				log.Fatalf("failed to upload file, %v", err)
			}
			log.Printf("S3 Upload result : %s", fn.ToJSON(output))
		}

		// If not local, upload the WebP file to S3
		if !ip.Config.IsLocal {
			transferInput := fileservice.TransferInput{
				Bucket:         ip.Config.TargetBucket,
				SourceFilePath: webpOutputFilePath,
				TargetFilePath: outputKey,
			}
			if err := ip.FileService.Transfer(transferInput); err != nil {
				return fmt.Errorf("failed to upload WebP to S3: %v", err)
			}
		}

		// Cleanup intermediate files
		// if ip.Config.AutoCleanup {
		// 	os.Remove(intermediatePNGPath)
		// 	if suffix == "watermark" {
		// 		os.Remove(fmt.Sprintf("%s-watermarked.png", intermediateFilePath))
		// 	}
		// 	os.Remove(unwatermarkedPNGPath)
	}

	// Cleanup source file
	if ip.Config.AutoCleanup {
		os.Remove(sourceFilePath)
	}

	return nil
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

	parentDir := filepath.Dir(absFolderPath)
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	newFolderName := filepath.Join(parentDir, timestamp)

	// Rename the folder
	if err := os.Rename(absFolderPath, newFolderName); err != nil {
		return fmt.Errorf("failed to rename folder: %v", err)
	}

	fmt.Printf("Folder renamed from %s to %s\n", absFolderPath, newFolderName)
	return nil
}

func (ip *ImageProcessor) downloadWorker(id int, errorChan chan<- error) {
	defer log.Printf("Download Worker %d: Finished", id)

	for file := range ip.DownloadQueue {
		log.Printf("Download Worker %d: Downloading file %s", id, file.ObjectKey)

		localPath, err := ip.downloadFile(&file)
		if err != nil {
			errorChan <- fmt.Errorf("download worker %d: Failed to download %s: %v", id, file.ObjectKey, err)
			return
		}

		log.Printf("ip.ProcessQueue <- localPath : %s", localPath)
		ip.ProcessQueue <- file
		log.Printf("Download Worker %d: File %s added to processing queue", id, file.ObjectKey)
	}
}

func (ip *ImageProcessor) downloadFile(file *imagefile.ImageFile) (string, error) {
	var relativePath string

	if ip.Config.IsLocal {
		relativePath, _ = filepath.Rel(ip.Config.LocalSource, file.ObjectKey)
	} else {
		relativePath = file.ObjectKey
	}

	fmt.Printf("file : %s\n", file.ObjectKey)
	fmt.Printf("relativePath : %s\n", relativePath)

	// Construct the local path in the working directory
	localPath := filepath.Join(ip.Config.WorkDir, ip.UUID, "source", relativePath)

	// log.Printf("relativePath : %s", relativePath)
	log.Printf("ip.Config.WorkDir : %s", ip.Config.WorkDir)
	log.Printf("localPath : %s", localPath)
	log.Printf("IP SourceDir : %s", ip.Config.LocalSource)
	log.Printf("Downloading file : %s to %s", file.ObjectKey, filepath.Join(localPath, file.ObjectKey))

	ip.FileService.Download(file, filepath.Join(localPath, file.ObjectKey))

	return localPath, nil
}

func (ip *ImageProcessor) processWorker(id int, errorChan chan<- error) {
	defer log.Printf("Processing Worker %d: Finished", id)
	for file := range ip.ProcessQueue {
		log.Printf("Processing Worker %d: Processing file %s", id, file.ObjectKey)
		if err := ip.ProcessFile(file); err != nil {
			errorChan <- fmt.Errorf("processing Worker %d: Failed to process file %s: %v", id, file.ObjectKey, err)
			return
		}
		log.Printf("Processing Worker %d: Successfully processed file: %s", id, file.ObjectKey)
	}
}

// Cleanup removes temporary directories and files
func (ip *ImageProcessor) Cleanup() {
	if ip.Config.AutoCleanup {
		os.Remove(filepath.Join(ip.Config.WorkDir, "source", ip.UUID))
		os.Remove(filepath.Join(ip.Config.WorkDir, "intermediate", ip.UUID))
		time.Sleep(1 * time.Second)
		os.RemoveAll(ip.Config.WorkDir)
	}
}
