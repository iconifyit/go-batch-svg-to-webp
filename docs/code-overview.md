# Code Overview

This document provides a detailed technical overview of the `go-batch-svg-to-webp` project, a high-performance CLI tool for batch-converting SVG images to optimized WebP format.

## Table of Contents

- [Project Summary](#project-summary)
- [Technology Stack](#technology-stack)
- [Directory Structure](#directory-structure)
- [Core Components](#core-components)
  - [Entry Point](#entry-point)
  - [Image Processor](#image-processor)
  - [File Service](#file-service)
  - [Image File](#image-file)
  - [Database Service](#database-service)
  - [Models](#models)
  - [Common Utilities](#common-utilities)
  - [Mocks](#mocks)
- [Configuration](#configuration)
- [Concurrency Model](#concurrency-model)
- [Processing Pipeline](#processing-pipeline)
- [External Dependencies](#external-dependencies)
- [Build and Run](#build-and-run)

---

## Project Summary

This application batch-processes SVG images into optimized WebP format at multiple sizes. It was built for VectorIcons.com to retroactively generate WebP versions for 500,000+ existing images. The Go implementation with concurrent worker pools achieved a 16x performance improvement over single-threaded execution.

**Key capabilities:**
- Convert SVG to PNG using `rsvg-convert`
- Convert PNG to WebP using `ffmpeg`
- Apply watermarks to preview images
- Support both local filesystem and AWS S3 storage
- Configurable worker pools for download and processing operations
- PostgreSQL-based contributor validation

---

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.22 |
| Cloud Storage | AWS S3 |
| Authentication | AWS STS (IAM role assumption) |
| Database | PostgreSQL |
| ORM | GORM v1.25.12 |
| Image Processing | rsvg-convert (SVG to PNG), ffmpeg (PNG to WebP) |
| Configuration | YAML (gopkg.in/yaml.v2) |
| CLI Flags | spf13/pflag |
| Testing | testify v1.8.1 |
| UUID Generation | google/uuid |
| Environment | joho/godotenv |

---

## Directory Structure

```
go-batch-svg-to-webp/
├── main.go                      # CLI entry point
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── config-example.yml           # Configuration template
├── build.sh                     # Build script
├── run.sh                       # Run script with RAM disk setup
├── mock.sh                      # Mock generation script
├── setup.sh                     # Setup script
├── src/
│   ├── image-processor/         # Main orchestrator
│   │   ├── main.go              # ImageProcessor struct and Run() logic
│   │   ├── config.go            # Config struct and YAML parsing
│   │   ├── funcs.go             # Processing functions (SVG->PNG->WebP)
│   │   ├── IImageProcessor.go   # Interface definition
│   │   ├── main_test.go         # Unit tests
│   │   ├── funcs_test.go        # Function tests
│   │   └── config_test.go       # Config tests
│   ├── file-service/            # Storage abstraction layer
│   │   ├── IFileService.go      # Interface + factory function
│   │   ├── LocalFileService.go  # Local filesystem implementation
│   │   ├── S3FileService.go     # AWS S3 implementation
│   │   ├── ServiceInput.go      # Service initialization input
│   │   ├── TransferInput.go     # File transfer input
│   │   ├── ListFilesInput.go    # File listing input
│   │   ├── FileServiceConfig.go # Service configuration
│   │   └── *_test.go            # Tests
│   ├── image-file/              # Image metadata parser
│   │   ├── main.go              # ImageFile struct and path parsing
│   │   ├── IImageFile.go        # Interface definition
│   │   └── *_test.go            # Tests
│   ├── database/                # PostgreSQL integration
│   │   ├── main.go              # DatabaseService struct and connection
│   │   ├── IDatabaseService.go  # Interface definition
│   │   ├── QueryParams.go       # Query parameter struct
│   │   ├── fn-user.go           # User query functions
│   │   ├── fn-family.go         # Family query functions
│   │   ├── fn-set.go            # Set query functions
│   │   ├── fn-icon.go           # Icon query functions
│   │   ├── fn-illustration.go   # Illustration query functions
│   │   └── *_test.go            # Tests
│   ├── models/                  # GORM data models
│   │   ├── User.go              # User entity
│   │   ├── Family.go            # Family entity (product collection)
│   │   ├── Set.go               # Set entity (subset of family)
│   │   ├── Icon.go              # Icon entity
│   │   ├── Illustration.go      # Illustration entity
│   │   ├── Role.go              # User role entity
│   │   ├── Image.go             # Polymorphic image entity
│   │   ├── ImageType.go         # Image type enum
│   │   └── UserToRoles.go       # User-Role junction table
│   ├── common/                  # Shared utilities
│   │   └── main.go              # ToJSON, CopyFile, StringInSlice, etc.
│   └── mocks/                   # Test doubles
│       ├── MockImageFile.go
│       ├── MockDatabaseService.go
│       ├── MockLocalFileService.go
│       ├── MockImageProcessor.go
│       └── MockS3FileService.go
├── test/                        # Test fixtures (SVG files; not committed - supply your own)
├── aws/iam/policies/            # AWS IAM policy definitions
├── docs/                        # Documentation
└── .github/                     # GitHub workflows
```

---

## Core Components

### Entry Point

**File:** `main.go:13`

The CLI entry point parses command-line flags and initializes the `ImageProcessor`:

```go
func main() {
    pflag.StringVarP(&contributor, "contributor", "c", "", "Contributor name")
    pflag.StringVarP(&configFile, "file", "f", "", "Path to the configuration file")
    pflag.Parse()

    imageProcessor := ip.NewImageProcessor(contributor, configFile)
    imageProcessor.Run()
}
```

**CLI Usage:**
```bash
./image-processor -f config.yml -c contributor-name
```

---

### Image Processor

**File:** `src/image-processor/main.go:28`

The main orchestrator that manages the entire processing lifecycle.

**Struct:**
```go
type ImageProcessor struct {
    UUID          string                   // Unique run identifier
    Contributor   string                   // Vendor/contributor name
    Config        *Config                  // YAML configuration
    Session       *session.Session         // AWS session with assumed role
    DownloadQueue chan imagefile.ImageFile // Buffered channel for downloads
    ProcessQueue  chan imagefile.ImageFile // Buffered channel for processing
    FileService   fileservice.IFileService // Storage abstraction
}
```

**Key Methods:**

| Method | Location | Description |
|--------|----------|-------------|
| `NewImageProcessor()` | `main.go:63` | Constructor - loads config, assumes AWS role, validates contributor |
| `Run()` | `main.go:408` | Main execution - sets up logging, spawns workers, coordinates processing |
| `ProcessFiles()` | `main.go:352` | Legacy single-queue processing method |
| `ListFiles()` | `main.go:261` | Lists files from source (local or S3) |
| `IsValidContributor()` | `main.go:294` | Validates contributor against database |
| `SetupLogging()` | `main.go:180` | Configures logging output (console, file, or both) |
| `ShouldInclude()` | `main.go:217` | Filters files based on include/exclude prefixes |
| `Cleanup()` | `funcs.go:446` | Removes temporary files and directories |

**AWS Role Assumption:**

The processor uses AWS STS to assume an IAM role for S3 access:

```go
// SessionWithRole creates a new session with the specified role
func SessionWithRole(roleArn string, region string) (*session.Session, error)
```
Location: `main.go:125`

---

### File Service

**File:** `src/file-service/IFileService.go:7`

Abstract file operations using the Strategy pattern to support multiple storage backends.

**Interface:**
```go
type IFileService interface {
    Transfer(input TransferInput) error
    ListFiles(input ListFilesInput, callback func(*string) bool) ([]string, error)
    ToImageFiles(files []string) ([]*imagefile.ImageFile, error)
    Download(file *imagefile.ImageFile, dest string) (string, error)
}
```

**Factory Function:**
```go
func NewFileService(input ServiceInput) IFileService {
    if input.IsLocal {
        return &LocalFileService{...}
    }
    return &S3FileService{...}
}
```
Location: `IFileService.go:14`

**Implementations:**

| Implementation | File | Description |
|----------------|------|-------------|
| `LocalFileService` | `LocalFileService.go:15` | Direct filesystem operations using `os` and `filepath` |
| `S3FileService` | `S3FileService.go:17` | AWS SDK-based S3 operations |

**LocalFileService Methods:**
- `Transfer()` - Copies file using `io.Copy`
- `ListFiles()` - Uses `filepath.Walk` to traverse directories
- `Download()` - Copies from source to destination
- `ToImageFiles()` - Converts paths to ImageFile objects

**S3FileService Methods:**
- `Transfer()` - Uploads to S3 using `PutObject`
- `ListFiles()` - Paginates through S3 bucket using `ListObjectsV2Pages`
- `Download()` - Downloads from S3 using `GetObject`
- `Upload()` - Uploads file to S3 bucket
- `Exists()` - Checks if object exists using `HeadObject`

---

### Image File

**File:** `src/image-file/main.go:12`

Represents a parsed image file with metadata extracted from its path structure.

**Expected Path Structure:**
```
{contributor}/{type}/{familyUUID}/{setUUID}/{filename}.{ext}
Example: iconify/icons/2C5AC6FCEB8D/6591D6BE211A/coffee-cup.svg
```

**Struct:**
```go
type ImageFile struct {
    URL               string   // Original path/URL
    ObjectKey         string   // Normalized S3 key or relative path
    InputPath         string   // Input file path
    Contributor       string   // Vendor name
    ProductType       *string  // "icons" or "illustrations"
    FamilyUniqueID    string   // 12-char UUID
    SetUniqueID       *string  // 12-char UUID (optional)
    Filename          string   // Base filename
    Extension         string   // File extension
    Slug              string   // URL-safe name
    Stem              string   // Filename without size suffix
    IsValid           bool     // Validation status
    Error             string   // Error message if invalid
    OptimizedImageKey string   // Target WebP key
    FileType          string   // File type
}
```

**Path Patterns:**

The `NewImageFile()` function matches against multiple regex patterns:

| Pattern | Example |
|---------|---------|
| Web Set | `https://vectopus.com/iconify/icons/2C5AC6FCEB8D/6591D6BE211A/file.svg` |
| Web Family | `https://vectopus.com/iconify/previews/2C5AC6FCEB8D/file.svg` |
| Local Set | `iconify/icons/2C5AC6FCEB8D/6591D6BE211A/file.svg` |
| Local Family | `iconify/previews/2C5AC6FCEB8D/file.svg` |

**Key Methods:**
- `NewImageFile(inputPath)` - Constructor with path parsing (`main.go:87`)
- `buildObjectKey()` - Constructs S3/file key (`main.go:199`)
- `getOptimizedImageKey()` - Generates WebP output key (`main.go:188`)
- `stripSizeFromSlug()` - Removes size suffixes like `@2x` (`main.go:210`)
- `Exists()` - Checks if file exists (`main.go:272`)
- `ToJSON()` - Converts to JSON map (`main.go:229`)

---

### Database Service

**File:** `src/database/main.go:15`

PostgreSQL integration using GORM for contributor validation.

**Struct:**
```go
type DatabaseService struct {
    DB *gorm.DB
}
```

**Connection:**
```go
func NewDatabaseService() (*DatabaseService, error)
```
Location: `main.go:38`

Environment variables required (loaded via godotenv):
- `POSTGRES_HOST`
- `POSTGRES_USER`
- `POSTGRES_PASS`
- `POSTGRES_DB`
- `POSTGRES_PORT`

**Connection Pool Settings:**
- Max open connections: 10
- Max idle connections: 5
- Max connection lifetime: 30 minutes

**Interface (IDatabaseService):**

Location: `IDatabaseService.go:5`

```go
type IDatabaseService interface {
    GetFamilyById(id int) (*models.Family, error)
    GetFamily(params *QueryParams) (*models.Family, error)
    GetFamilies(params QueryParams) ([]models.Family, error)
    GetIconById(id int) (*models.Icon, error)
    GetIcon(params QueryParams) (*models.Icon, error)
    GetIcons(params QueryParams) ([]models.Icon, error)
    GetSetById(id int) (*models.Set, error)
    GetSet(params *QueryParams) (*models.Set, error)
    GetSets(params QueryParams) ([]models.Set, error)
    GetIllustrationById(id int) (*models.Illustration, error)
    GetIllustration(params QueryParams) (*models.Illustration, error)
    GetIllustrations(params QueryParams) ([]models.Illustration, error)
    GetUserById(id int) (*models.User, error)
    GetUser(params *QueryParams) (*models.User, error)
    GetUsers(params QueryParams) ([]models.User, error)
    Close() error
}
```

**Query Helper Functions:**
- `Where(column, value)` - Basic equality filter
- `WhereNot(column, value)` - Negation filter
- `WhereIn(column, values)` - IN clause
- `WhereLike(column, pattern)` - LIKE clause
- `WhereCustom(condition, args...)` - Custom SQL condition

---

### Models

**Directory:** `src/models/`

GORM data models representing database entities.

| Model | File | Description |
|-------|------|-------------|
| `User` | `User.go:8` | User entity with roles and images |
| `Family` | `Family.go:7` | Product collection (container for sets) |
| `Set` | `Set.go:7` | Subset of family (container for icons) |
| `Icon` | `Icon.go:7` | Individual icon entity |
| `Illustration` | `Illustration.go` | Illustration entity |
| `Role` | `Role.go` | User role entity |
| `Image` | `Image.go` | Polymorphic image entity |
| `ImageType` | `ImageType.go` | Image type constants |
| `UserToRoles` | `UserToRoles.go` | User-Role junction table |

**User Model (key fields):**
```go
type User struct {
    ID          int       `gorm:"primaryKey"`
    Email       string    `gorm:"unique;not null"`
    Username    string    `gorm:"not null"`
    UUID        string    `gorm:"type:uuid;default:gen_random_uuid()"`
    Roles       []Role    `gorm:"many2many:user_to_roles"`
    Images      []Image   `gorm:"polymorphic:Entity;polymorphicValue:user"`
    // ... other fields
}
```

---

### Common Utilities

**File:** `src/common/main.go`

Shared utility functions used across the application.

| Function | Location | Description |
|----------|----------|-------------|
| `ToJSON(data)` | `main.go:15` | Converts interface to pretty-printed JSON string |
| `StringInSlice(s, slice)` | `main.go:30` | Checks if string exists in slice |
| `CopyFile(src, dst)` | `main.go:40` | Copies file using `io.Copy` |
| `AsType(filepath, newType)` | `main.go:58` | Changes file extension |
| `AddSuffix(filepath, suffix)` | `main.go:66` | Adds suffix before extension |

---

### Mocks

**Directory:** `src/mocks/`

Test doubles for unit testing using interfaces.

| Mock | File | Mocks |
|------|------|-------|
| `MockImageFile` | `MockImageFile.go` | `IImageFile` interface |
| `MockDatabaseService` | `MockDatabaseService.go` | `IDatabaseService` interface |
| `MockLocalFileService` | `MockLocalFileService.go` | `IFileService` (local) |
| `MockS3FileService` | `MockS3FileService.go` | `IFileService` (S3) |
| `MockImageProcessor` | `MockImageProcessor.go` | `IImageProcessor` interface |

---

## Configuration

**File:** `src/image-processor/config.go:11`

YAML-based configuration with sensible defaults.

**Struct:**
```go
type Config struct {
    Contributor             string
    SourceBucket            string         `yaml:"source_bucket"`
    TargetBucket            string         `yaml:"target_bucket"`
    Include                 []string       `yaml:"include_prefixes"`
    Exclude                 []string       `yaml:"omit_prefixes"`
    Region                  string         `yaml:"aws_region"`
    DryRun                  bool           `yaml:"dry_run"`
    FFmpegPath              string         `yaml:"ffmpegPath"`
    IsLocal                 bool           `yaml:"is_local"`
    UploadToS3              bool           `yaml:"upload_to_s3"`
    LocalSource             string         `yaml:"local_source"`
    LocalTarget             string         `yaml:"local_target"`
    AutoCleanup             bool           `yaml:"auto_cleanup"`
    WebpSizes               map[string]int `yaml:"webp_sizes"`
    WatermarkPath           string         `yaml:"watermark_path"`
    RoleArn                 string         `yaml:"role_arn"`
    LoggingOutput           int            `yaml:"logging_output"`
    Logfile                 string         `yaml:"logfile"`
    WorkDir                 string         `yaml:"work_dir"`
    OutputDir               string         `yaml:"output_dir"`
    UseHardwareAcceleration bool           `yaml:"use_hardware_acceleration"`
    WorkerPoolSize          int            `yaml:"worker_pool_size"`
    DownloadWorkerPoolSize  int            `yaml:"download_worker_pool_size"`
    ProcessWorkerPoolSize   int            `yaml:"process_worker_pool_size"`
}
```

**Default Values (SetDefaults method at config.go:54):**

| Setting | Default |
|---------|---------|
| `IsLocal` | `false` |
| `UploadToS3` | `false` |
| `WorkDir` | `./tmp/work` |
| `OutputDir` | `./tmp/output` |
| `UseHardwareAcceleration` | `false` |
| `WorkerPoolSize` | `1` |
| `DownloadWorkerPoolSize` | `1` |
| `Logfile` | `./tmp/image-processor.log` |
| `AutoCleanup` | `false` |
| `Region` | `us-east-1` |

**Logging Output Options:**
- `0` - No output
- `1` - Console only
- `2` - File only
- `3` - Both console and file

**WebP Sizes Configuration:**
```yaml
webp_sizes:
  thumbnail: 128    # 128px thumbnail
  preview: 512      # 512px preview
  watermark: 512    # 512px with watermark
```

---

## Concurrency Model

The application uses a **producer-consumer pattern** with **dual worker pools**.

### Architecture

```
                    ┌─────────────────┐
                    │  Main Goroutine │
                    │   (Producer)    │
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ DownloadQueue   │
                    │ (Buffered Chan) │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ▼                    ▼                    ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│ Download      │    │ Download      │    │ Download      │
│ Worker 1      │    │ Worker 2      │    │ Worker N      │
└───────┬───────┘    └───────┬───────┘    └───────┬───────┘
        │                    │                    │
        └────────────────────┼────────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │  ProcessQueue   │
                    │ (Buffered Chan) │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ▼                    ▼                    ▼
┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│  Process      │    │  Process      │    │  Process      │
│  Worker 1     │    │  Worker 2     │    │  Worker M     │
└───────────────┘    └───────────────┘    └───────────────┘
```

### Implementation

**Worker Pool Setup (main.go:408):**

```go
func (ip *ImageProcessor) Run() error {
    // Initialize queues
    ip.DownloadQueue = make(chan imagefile.ImageFile, len(imageFiles))
    ip.ProcessQueue = make(chan imagefile.ImageFile, len(imageFiles))
    errorChan := make(chan error, 1)

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

    // Queue files for downloading
    go func() {
        for _, file := range imageFiles {
            ip.DownloadQueue <- file
        }
        close(ip.DownloadQueue)
    }()

    // Wait for downloads, then close process queue
    go func() {
        downloadWG.Wait()
        close(ip.ProcessQueue)
    }()

    // Wait for processing, then close error channel
    go func() {
        processWG.Wait()
        close(errorChan)
    }()

    // Monitor for errors
    for err := range errorChan {
        if err != nil {
            return err
        }
    }

    return nil
}
```

**Download Worker (funcs.go:389):**
```go
func (ip *ImageProcessor) downloadWorker(id int, errorChan chan<- error) {
    for file := range ip.DownloadQueue {
        localPath, err := ip.downloadFile(&file)
        if err != nil {
            errorChan <- fmt.Errorf("download worker %d: %v", id, err)
            return
        }
        ip.ProcessQueue <- file
    }
}
```

**Process Worker (funcs.go:433):**
```go
func (ip *ImageProcessor) processWorker(id int, errorChan chan<- error) {
    for file := range ip.ProcessQueue {
        if err := ip.ProcessFile(file); err != nil {
            errorChan <- fmt.Errorf("process worker %d: %v", id, err)
            return
        }
    }
}
```

---

## Processing Pipeline

Each image goes through a multi-stage pipeline:

### Stage 1: Download (funcs.go:407)

1. Receive `ImageFile` from `DownloadQueue`
2. Determine relative path
3. Create local path in working directory: `{workDir}/{uuid}/source/{objectKey}`
4. Copy from source (local) or download from S3
5. Push to `ProcessQueue`

### Stage 2: Process (funcs.go:137)

For each configured size (thumbnail, preview, watermark):

1. **Setup Directories:**
   - `{workDir}/{uuid}/source/` - Downloaded source files
   - `{workDir}/{uuid}/intermediate/` - PNG intermediates
   - `{workDir}/{uuid}/output/` - Final WebP files

2. **SVG to PNG Conversion (funcs.go:24):**
   ```go
   func (ip *ImageProcessor) ConvertSVGToPNG(svgPath, pngPath string, size int) error {
       cmd := exec.Command("rsvg-convert",
           "--background-color", "white",
           "-w", fmt.Sprint(size),
           "-h", fmt.Sprint(size),
           svgPath, "-o", pngPath)
       return cmd.Run()
   }
   ```

3. **Watermarking (funcs.go:73) - for "watermark" size only:**
   ```go
   func (ip *ImageProcessor) Watermark(inputFilePath, outputFilePath string, size int) error {
       // Convert watermark SVG to PNG
       // Apply overlay using ffmpeg: overlay=W-w-10:H-h-10
   }
   ```

4. **PNG to WebP Conversion (funcs.go:41):**
   ```go
   func (ip *ImageProcessor) RunFFmpeg(input, output string) error {
       args := []string{"-i", input, "-vf", "format=yuv420p", "-q:v", "75", output}
       if ip.Config.UseHardwareAcceleration {
           args = append([]string{"-hwaccel", "videotoolbox"}, args...)
       }
       cmd := exec.Command(ip.Config.FFmpegPath, args...)
       return cmd.Run()
   }
   ```

5. **Upload (if configured):**
   - Upload to S3 target bucket
   - Or save to local output directory

6. **Cleanup (if AutoCleanup enabled):**
   - Remove intermediate PNG files
   - Remove source files

---

## External Dependencies

### System Requirements

| Tool | Purpose | Installation |
|------|---------|--------------|
| `rsvg-convert` | SVG to PNG conversion | `brew install librsvg` (macOS) / `apt install librsvg2-bin` (Ubuntu) |
| `ffmpeg` | PNG to WebP conversion, watermarking | `brew install ffmpeg` (macOS) / `apt install ffmpeg` (Ubuntu) |

### Go Dependencies (go.mod)

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/aws/aws-sdk-go` | v1.55.5 | AWS S3 and STS operations |
| `github.com/google/uuid` | v1.6.0 | UUID generation |
| `github.com/joho/godotenv` | v1.5.1 | .env file loading |
| `github.com/spf13/pflag` | v1.0.5 | CLI flag parsing |
| `github.com/stretchr/testify` | v1.8.1 | Test assertions |
| `gopkg.in/yaml.v2` | v2.4.0 | YAML configuration parsing |
| `gorm.io/driver/postgres` | v1.5.11 | PostgreSQL driver for GORM |
| `gorm.io/gorm` | v1.25.12 | ORM for database operations |

---

## Build and Run

### Build

```bash
./build.sh
# or
go build -o image-processor main.go
```

### Run (with RAM disk for performance)

The `run.sh` script:
1. Builds the binary
2. Creates a 4GB RAM disk for temporary files
3. Updates config.yml with RAM disk paths
4. Runs the image processor
5. Restores original config

```bash
./run.sh
```

### Manual Run

```bash
./image-processor -f config.yml -c contributor-name
```

### Environment Variables

Create a `.env` file with database credentials:
```
POSTGRES_HOST=localhost
POSTGRES_USER=postgres
POSTGRES_PASS=password
POSTGRES_DB=vectoricons
POSTGRES_PORT=5432
```

---

## Performance Characteristics

| Configuration | Files | Time | Throughput |
|---------------|-------|------|------------|
| Single-threaded | 4,500 | 7m 17s | 10.4 files/sec |
| 10 Workers | 4,500 | 27s | 166.67 files/sec |
| Extrapolated (500K) | 500,000 | ~50 min | ~166 files/sec |

**Optimizations:**
- Dual worker pools for I/O and CPU-bound operations
- Buffered channels prevent worker starvation
- Optional hardware acceleration (VideoToolbox on macOS)
- RAM disk usage for temporary files (via run.sh)
- Connection pooling for database operations
