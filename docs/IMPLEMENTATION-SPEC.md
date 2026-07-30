# Implementation Specification: Go App Safety Updates

This document details the exact changes to be made to the Go application to implement Sprint 1 (Safety Features) and Sprint 2 (Staging Infrastructure) from the ROADMAP.

## Table of Contents

- [Overview](#overview)
- [S3 Object Metadata Strategy](#s3-object-metadata-strategy)
- [File Changes Summary](#file-changes-summary)
- [Detailed Changes](#detailed-changes)
  - [1. Config Updates](#1-config-updates)
  - [2. Database Service Updates](#2-database-service-updates)
  - [3. File Service Updates](#3-file-service-updates)
  - [4. Image Processor Updates](#4-image-processor-updates)
  - [5. New Statistics Module](#5-new-statistics-module)
  - [6. New Manifest Module](#6-new-manifest-module)
- [S3 Upload with Metadata](#s3-upload-with-metadata)
- [Testing Strategy](#testing-strategy)
- [Migration Path](#migration-path)

---

## Overview

### Goals

1. **Never process files that already have WebP versions** (check DB + production S3)
2. **Upload to staging bucket only** (never directly to production)
3. **Attach metadata to each WebP** so Lambda knows what to do with it
4. **Generate batch manifest** for audit trail and recovery
5. **Track statistics** for monitoring and reporting

### Non-Goals (Handled by Lambda later)

- Creating database records
- Copying files to production bucket
- Deleting staging files after processing

---

## S3 Object Metadata Strategy

S3 allows custom metadata on objects (up to 2KB total). We'll attach metadata to each uploaded WebP file that tells the Lambda everything it needs to know.

### Metadata Fields

| Key | Example Value | Description |
|-----|---------------|-------------|
| `x-amz-meta-batch-id` | `2024-01-15-iconify-001` | Unique batch identifier |
| `x-amz-meta-source-key` | `iconify/icons/ABC123/DEF456/cup.svg` | Original SVG object key |
| `x-amz-meta-entity-type` | `icon` | Entity type: `icon`, `illustration`, `family`, `set` |
| `x-amz-meta-entity-contributor` | `iconify` | Contributor username |
| `x-amz-meta-entity-family-id` | `ABC123` | Family unique ID (12-char) |
| `x-amz-meta-entity-set-id` | `DEF456` | Set unique ID (12-char, optional) |
| `x-amz-meta-webp-size` | `preview` | Size variant: `thumbnail`, `preview`, `watermark` |
| `x-amz-meta-webp-dimensions` | `512` | Pixel dimensions |
| `x-amz-meta-target-bucket` | `vectoricons-public` | Final destination bucket |
| `x-amz-meta-target-key` | `iconify/icons/ABC123/DEF456/cup-preview.webp` | Final destination key |
| `x-amz-meta-processed-at` | `2024-01-15T10:30:00Z` | Processing timestamp |
| `x-amz-meta-image-name` | `cup-preview` | Name for DB record |
| `x-amz-meta-image-visibility` | `public` | Visibility for DB record |
| `x-amz-meta-image-access` | `all` | Access level for DB record |

### Why Object Metadata vs Companion JSON Files

| Approach | Pros | Cons |
|----------|------|------|
| **Object Metadata** | Atomic (metadata travels with file), simpler Lambda, fewer S3 operations | 2KB limit, harder to inspect |
| **Companion JSON** | Unlimited size, easy to inspect | Two files per WebP, race conditions possible, more S3 operations |

**Decision:** Use object metadata. Our metadata fits well under 2KB, and the atomic nature prevents orphaned files.

### Lambda Processing Flow

```
1. S3 Event: ObjectCreated on vectoricons-webp-staging/*.webp
2. Lambda reads object metadata via HeadObject
3. Lambda extracts target-bucket, target-key, entity-* fields
4. Lambda looks up entity ID in DB using entity-* fields
5. Lambda copies object to target-bucket/target-key
6. Lambda creates Image record in DB
7. Lambda deletes object from staging bucket
```

---

## File Changes Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `src/image-processor/config.go` | Modify | Add new config fields |
| `src/image-processor/main.go` | Modify | Add existence check, stats tracking |
| `src/image-processor/funcs.go` | Modify | Update ProcessFile, add metadata upload |
| `src/database/fn-image.go` | **New** | Add GetImageByObjectKey method |
| `src/database/IDatabaseService.go` | Modify | Add new interface methods |
| `src/image-processor/stats.go` | **New** | Processing statistics |
| `src/image-processor/manifest.go` | **New** | Batch manifest generation |
| `src/image-processor/metadata.go` | **New** | S3 metadata builder |
| `src/file-service/S3FileService.go` | Modify | Add UploadWithMetadata method |
| `src/file-service/IS3FileService.go` | Modify | Add new interface method |
| `config-production.yml` | **New** | Production configuration template |

---

## Detailed Changes

### 1. Config Updates

**File:** `src/image-processor/config.go`

#### Current State (lines 11-36)

```go
type Config struct {
    Contributor             string
    SourceBucket            string         `yaml:"source_bucket"`
    TargetBucket            string         `yaml:"target_bucket"`
    // ... existing fields
}
```

#### Proposed Changes

Add new fields after existing ones:

```go
type Config struct {
    // === EXISTING FIELDS (unchanged) ===
    Contributor             string         `yaml:"contributor"`
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

    // === NEW FIELDS ===

    // Staging bucket - Go app uploads here (Lambda moves to production)
    StagingBucket string `yaml:"staging_bucket"`

    // Production bucket - for existence checks only (read-only access)
    ProductionBucket string `yaml:"production_bucket"`

    // Safety checks
    SkipExistingWebP     bool `yaml:"skip_existing_webp"`
    CheckDBBeforeProcess bool `yaml:"check_db_before_process"`
    CheckS3BeforeProcess bool `yaml:"check_s3_before_process"`

    // Batch control
    BatchID           string `yaml:"batch_id"`
    MaxFilesToProcess int    `yaml:"max_files_to_process"`

    // Manifest generation
    GenerateManifest bool   `yaml:"generate_manifest"`
    ManifestDir      string `yaml:"manifest_dir"`
}
```

#### Update SetDefaults Method

Add defaults for new fields in `SetDefaults()` (around line 54):

```go
func (config *Config) SetDefaults() {
    // ... existing defaults ...

    // NEW: Staging defaults
    if config.StagingBucket == "" {
        config.StagingBucket = "vectoricons-webp-staging"
    }

    // NEW: Safety defaults - all enabled by default for safety
    if !config.SkipExistingWebP {
        config.SkipExistingWebP = true
    }
    if !config.CheckDBBeforeProcess {
        config.CheckDBBeforeProcess = true
    }
    if !config.CheckS3BeforeProcess {
        config.CheckS3BeforeProcess = true
    }

    // NEW: Batch defaults
    if config.BatchID == "" {
        config.BatchID = fmt.Sprintf("%s-%s-%s",
            time.Now().Format("2006-01-02"),
            config.Contributor,
            uuid.New().String()[:8],
        )
    }
    if config.MaxFilesToProcess == 0 {
        config.MaxFilesToProcess = -1 // -1 means no limit
    }

    // NEW: Manifest defaults
    if config.ManifestDir == "" {
        config.ManifestDir = "./manifests"
    }
    config.GenerateManifest = true // Always generate for safety
}
```

---

### 2. Database Service Updates

#### New File: `src/database/fn-image.go`

```go
package database

import (
    "github.com/iconifyit/go-batch-svg-to-webp/src/models"
    "gorm.io/gorm"
)

// GetImageByObjectKey retrieves an image by its S3 object key
func (svc *DatabaseService) GetImageByObjectKey(objectKey string) (*models.Image, error) {
    var image models.Image
    err := svc.DB.Where("object_key = ?", objectKey).First(&image).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, nil // Not found is not an error for existence checks
        }
        return nil, err
    }
    return &image, nil
}

// GetImagesByObjectKeys retrieves multiple images by their S3 object keys
// Useful for batch existence checks
func (svc *DatabaseService) GetImagesByObjectKeys(objectKeys []string) ([]models.Image, error) {
    var images []models.Image
    err := svc.DB.Where("object_key IN ?", objectKeys).Find(&images).Error
    if err != nil {
        return nil, err
    }
    return images, nil
}

// ImageExistsByObjectKey checks if an image with the given object key exists
// More efficient than GetImageByObjectKey when you only need existence
func (svc *DatabaseService) ImageExistsByObjectKey(objectKey string) (bool, error) {
    var count int64
    err := svc.DB.Model(&models.Image{}).Where("object_key = ?", objectKey).Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

// WebPExistsForSVG checks if any WebP images exist for a given SVG source
// Checks all size variants (thumbnail, preview, watermark)
func (svc *DatabaseService) WebPExistsForSVG(baseObjectKey string, sizes []string) (map[string]bool, error) {
    result := make(map[string]bool)

    // Build list of potential WebP keys
    var webpKeys []string
    for _, size := range sizes {
        // Convert "iconify/icons/ABC/DEF/cup.svg" to "iconify/icons/ABC/DEF/cup-preview.webp"
        webpKey := strings.TrimSuffix(baseObjectKey, filepath.Ext(baseObjectKey)) + "-" + size + ".webp"
        webpKeys = append(webpKeys, webpKey)
        result[webpKey] = false // Initialize as not found
    }

    // Batch query for all variants
    images, err := svc.GetImagesByObjectKeys(webpKeys)
    if err != nil {
        return nil, err
    }

    // Mark found images
    for _, img := range images {
        result[img.ObjectKey] = true
    }

    return result, nil
}
```

#### Update Interface: `src/database/IDatabaseService.go`

Add new methods to the interface:

```go
type IDatabaseService interface {
    // ... existing methods ...

    // NEW: Image existence checks
    GetImageByObjectKey(objectKey string) (*models.Image, error)
    GetImagesByObjectKeys(objectKeys []string) ([]models.Image, error)
    ImageExistsByObjectKey(objectKey string) (bool, error)
    WebPExistsForSVG(baseObjectKey string, sizes []string) (map[string]bool, error)
}
```

---

### 3. File Service Updates

#### Update: `src/file-service/S3FileService.go`

Add new method for uploading with metadata:

```go
// WebPMetadata contains all metadata to attach to uploaded WebP files
type WebPMetadata struct {
    BatchID           string `json:"batch_id"`
    SourceKey         string `json:"source_key"`
    EntityType        string `json:"entity_type"`
    EntityContributor string `json:"entity_contributor"`
    EntityFamilyID    string `json:"entity_family_id"`
    EntitySetID       string `json:"entity_set_id"`
    WebPSize          string `json:"webp_size"`
    WebPDimensions    int    `json:"webp_dimensions"`
    TargetBucket      string `json:"target_bucket"`
    TargetKey         string `json:"target_key"`
    ProcessedAt       string `json:"processed_at"`
    ImageName         string `json:"image_name"`
    ImageVisibility   string `json:"image_visibility"`
    ImageAccess       string `json:"image_access"`
}

// UploadWithMetadata uploads a file to S3 with custom metadata attached
func (svc *S3FileService) UploadWithMetadata(localPath, objectKey string, metadata WebPMetadata) error {
    client := s3.New(svc.Session)

    fileBuffer, err := os.Open(localPath)
    if err != nil {
        return fmt.Errorf("failed to open local file: %v", err)
    }
    defer fileBuffer.Close()

    // Build metadata map for S3
    // Note: S3 automatically prepends "x-amz-meta-" to these keys
    s3Metadata := map[string]*string{
        "batch-id":           aws.String(metadata.BatchID),
        "source-key":         aws.String(metadata.SourceKey),
        "entity-type":        aws.String(metadata.EntityType),
        "entity-contributor": aws.String(metadata.EntityContributor),
        "entity-family-id":   aws.String(metadata.EntityFamilyID),
        "entity-set-id":      aws.String(metadata.EntitySetID),
        "webp-size":          aws.String(metadata.WebPSize),
        "webp-dimensions":    aws.String(fmt.Sprintf("%d", metadata.WebPDimensions)),
        "target-bucket":      aws.String(metadata.TargetBucket),
        "target-key":         aws.String(metadata.TargetKey),
        "processed-at":       aws.String(metadata.ProcessedAt),
        "image-name":         aws.String(metadata.ImageName),
        "image-visibility":   aws.String(metadata.ImageVisibility),
        "image-access":       aws.String(metadata.ImageAccess),
    }

    _, err = client.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(svc.TargetBucket),
        Key:         aws.String(objectKey),
        Body:        fileBuffer,
        ContentType: aws.String("image/webp"),
        Metadata:    s3Metadata,
    })
    if err != nil {
        return fmt.Errorf("failed to upload file to S3 with metadata: %v", err)
    }

    return nil
}

// HeadObject checks if an object exists and returns its metadata
func (svc *S3FileService) HeadObject(bucket, objectKey string) (bool, map[string]*string, error) {
    client := s3.New(svc.Session)

    result, err := client.HeadObject(&s3.HeadObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(objectKey),
    })
    if err != nil {
        // Check if it's a "not found" error
        if aerr, ok := err.(awserr.Error); ok {
            if aerr.Code() == "NotFound" || aerr.Code() == "NoSuchKey" {
                return false, nil, nil
            }
        }
        return false, nil, err
    }

    return true, result.Metadata, nil
}
```

#### Update Interface: `src/file-service/IS3FileService.go`

```go
type IS3FileService interface {
    // ... existing methods ...

    // NEW: Upload with metadata
    UploadWithMetadata(localPath, objectKey string, metadata WebPMetadata) error

    // NEW: Head object for existence checks
    HeadObject(bucket, objectKey string) (bool, map[string]*string, error)
}
```

---

### 4. Image Processor Updates

#### New File: `src/image-processor/metadata.go`

```go
package imageprocessor

import (
    "time"

    fileservice "github.com/iconifyit/go-batch-svg-to-webp/src/file-service"
    imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

// BuildWebPMetadata constructs the metadata to attach to an uploaded WebP file
func (ip *ImageProcessor) BuildWebPMetadata(
    imgFile imagefile.ImageFile,
    sizeName string,
    sizePixels int,
    targetKey string,
) fileservice.WebPMetadata {
    // Determine entity type from product type
    entityType := "icon" // default
    if imgFile.ProductType != nil {
        switch *imgFile.ProductType {
        case "icons":
            entityType = "icon"
        case "illustrations":
            entityType = "illustration"
        case "previews":
            entityType = "family"
        }
    }

    // Build image name from filename and size
    imageName := imgFile.Stem + "-" + sizeName

    // Get set ID (may be nil for family previews)
    setID := ""
    if imgFile.SetUniqueID != nil {
        setID = *imgFile.SetUniqueID
    }

    return fileservice.WebPMetadata{
        BatchID:           ip.Config.BatchID,
        SourceKey:         imgFile.ObjectKey,
        EntityType:        entityType,
        EntityContributor: imgFile.Contributor,
        EntityFamilyID:    imgFile.FamilyUniqueID,
        EntitySetID:       setID,
        WebPSize:          sizeName,
        WebPDimensions:    sizePixels,
        TargetBucket:      ip.Config.ProductionBucket,
        TargetKey:         targetKey,
        ProcessedAt:       time.Now().UTC().Format(time.RFC3339),
        ImageName:         imageName,
        ImageVisibility:   "public",
        ImageAccess:       "all",
    }
}
```

#### New File: `src/image-processor/existence.go`

```go
package imageprocessor

import (
    "log"
    "strings"

    imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
    dbsvc "github.com/iconifyit/go-batch-svg-to-webp/src/database"
)

// ExistenceCheckResult contains the result of checking if WebP versions exist
type ExistenceCheckResult struct {
    AnyExist     bool
    AllExist     bool
    ExistingKeys []string
    MissingKeys  []string
}

// CheckWebPExists checks both DB and S3 for existing WebP versions
func (ip *ImageProcessor) CheckWebPExists(imgFile imagefile.ImageFile) (*ExistenceCheckResult, error) {
    result := &ExistenceCheckResult{
        AnyExist:     false,
        AllExist:     true,
        ExistingKeys: []string{},
        MissingKeys:  []string{},
    }

    // Build list of WebP keys we would create
    sizes := make([]string, 0, len(ip.Config.WebpSizes))
    for sizeName := range ip.Config.WebpSizes {
        sizes = append(sizes, sizeName)
    }

    // Generate expected WebP object keys
    baseKey := strings.TrimSuffix(imgFile.ObjectKey, "."+imgFile.Extension)
    webpKeys := make([]string, 0, len(sizes))
    for _, size := range sizes {
        webpKey := baseKey + "-" + size + ".webp"
        webpKeys = append(webpKeys, webpKey)
    }

    // Check DB if enabled
    if ip.Config.CheckDBBeforeProcess {
        dbExists, err := ip.checkDBForWebP(webpKeys)
        if err != nil {
            return nil, fmt.Errorf("DB existence check failed: %v", err)
        }
        for key, exists := range dbExists {
            if exists {
                result.AnyExist = true
                result.ExistingKeys = append(result.ExistingKeys, key+" (DB)")
            }
        }
    }

    // Check S3 production bucket if enabled
    if ip.Config.CheckS3BeforeProcess {
        s3Exists, err := ip.checkS3ForWebP(webpKeys)
        if err != nil {
            return nil, fmt.Errorf("S3 existence check failed: %v", err)
        }
        for key, exists := range s3Exists {
            if exists {
                result.AnyExist = true
                // Avoid duplicates if already found in DB
                alreadyListed := false
                for _, existing := range result.ExistingKeys {
                    if strings.HasPrefix(existing, key) {
                        alreadyListed = true
                        break
                    }
                }
                if !alreadyListed {
                    result.ExistingKeys = append(result.ExistingKeys, key+" (S3)")
                }
            }
        }
    }

    // Determine missing keys
    for _, key := range webpKeys {
        found := false
        for _, existing := range result.ExistingKeys {
            if strings.HasPrefix(existing, key) {
                found = true
                break
            }
        }
        if !found {
            result.MissingKeys = append(result.MissingKeys, key)
            result.AllExist = false
        }
    }

    if len(result.MissingKeys) == 0 {
        result.AllExist = true
    }

    return result, nil
}

// checkDBForWebP queries the database for existing Image records
func (ip *ImageProcessor) checkDBForWebP(webpKeys []string) (map[string]bool, error) {
    result := make(map[string]bool)
    for _, key := range webpKeys {
        result[key] = false
    }

    db, err := dbsvc.NewDatabaseService()
    if err != nil {
        return nil, fmt.Errorf("failed to create database service: %v", err)
    }
    defer db.Close()

    images, err := db.GetImagesByObjectKeys(webpKeys)
    if err != nil {
        return nil, err
    }

    for _, img := range images {
        result[img.ObjectKey] = true
    }

    return result, nil
}

// checkS3ForWebP checks the production S3 bucket for existing files
func (ip *ImageProcessor) checkS3ForWebP(webpKeys []string) (map[string]bool, error) {
    result := make(map[string]bool)
    for _, key := range webpKeys {
        result[key] = false
    }

    // Use the S3FileService to check each key
    // Note: For large batches, consider using batch HeadObject or ListObjects
    for _, key := range webpKeys {
        exists, _, err := ip.checkS3ObjectExists(ip.Config.ProductionBucket, key)
        if err != nil {
            log.Printf("Warning: S3 existence check failed for %s: %v", key, err)
            continue // Don't fail the whole batch for one check
        }
        result[key] = exists
    }

    return result, nil
}

// checkS3ObjectExists checks if a single object exists in S3
func (ip *ImageProcessor) checkS3ObjectExists(bucket, key string) (bool, map[string]*string, error) {
    svc := s3.New(ip.Session)

    result, err := svc.HeadObject(&s3.HeadObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        if aerr, ok := err.(awserr.Error); ok {
            if aerr.Code() == "NotFound" || aerr.Code() == "NoSuchKey" {
                return false, nil, nil
            }
        }
        return false, nil, err
    }

    return true, result.Metadata, nil
}
```

#### Update: `src/image-processor/funcs.go`

Modify `ProcessFile` to use existence checks and upload with metadata:

**Location:** Around line 137

```go
// ProcessFile processes a single file
func (ip *ImageProcessor) ProcessFile(imgFile imagefile.ImageFile) error {
    log.Println("\n\n------------------------------------------------------")
    log.Printf("ProcessFile - file : %s", imgFile.ObjectKey)

    // =========================================================================
    // NEW: Check if WebP versions already exist
    // =========================================================================
    if ip.Config.SkipExistingWebP {
        existenceResult, err := ip.CheckWebPExists(imgFile)
        if err != nil {
            return fmt.Errorf("existence check failed for %s: %v", imgFile.ObjectKey, err)
        }

        if existenceResult.AnyExist {
            log.Printf("SKIP: WebP already exists for %s", imgFile.ObjectKey)
            log.Printf("  Existing: %v", existenceResult.ExistingKeys)
            ip.Stats.IncrementSkipped(imgFile.ObjectKey, "WebP already exists")
            return nil
        }
    }

    // ... existing directory setup code (lines 172-217) ...

    processSourceDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "source")
    processIntermediateDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "intermediate")
    processOutputDir := filepath.Join(ip.Config.WorkDir, ip.UUID, "output")

    // ... existing directory creation code ...

    // =========================================================================
    // Process each size for WebP conversion
    // =========================================================================
    for sizeName, sizePixels := range ip.Config.WebpSizes {
        ext := filepath.Ext(imgFile.ObjectKey)
        intermediatePNGPath := fmt.Sprintf("%s-%s.png", intermediateFilePath, sizeName)

        // Convert SVG to PNG
        if err := ip.ConvertSVGToPNG(sourceFilePath, intermediatePNGPath, sizePixels); err != nil {
            ip.Stats.IncrementFailed(imgFile.ObjectKey, fmt.Sprintf("SVG to PNG failed: %v", err))
            return fmt.Errorf("SVG to PNG conversion failed for %s: %v", sourceFilePath, err)
        }

        // Construct target WebP file path
        targetRelativePath := strings.Replace(imgFile.ObjectKey, ext, "", -1)
        localWebpPath := filepath.Join(processOutputDir, fmt.Sprintf("%s-%s.webp", targetRelativePath, sizeName))

        // Ensure output directory exists
        if err := os.MkdirAll(filepath.Dir(localWebpPath), os.ModePerm); err != nil {
            return fmt.Errorf("failed to create output subdirectory: %v", err)
        }

        // Add watermark if necessary
        if sizeName == "watermark" {
            watermarkedIntermediateFile := fmt.Sprintf("%s-watermarked.png", intermediateFilePath)
            if err := ip.Watermark(intermediatePNGPath, watermarkedIntermediateFile, sizePixels); err != nil {
                ip.Stats.IncrementFailed(imgFile.ObjectKey, fmt.Sprintf("Watermark failed: %v", err))
                return fmt.Errorf("watermark creation failed for %s: %v", intermediatePNGPath, err)
            }
            intermediatePNGPath = watermarkedIntermediateFile
        }

        // Convert PNG to WebP
        if err := ip.RunFFmpeg(intermediatePNGPath, localWebpPath); err != nil {
            ip.Stats.IncrementFailed(imgFile.ObjectKey, fmt.Sprintf("WebP conversion failed: %v", err))
            return fmt.Errorf("WebP conversion failed for %s: %v", intermediatePNGPath, err)
        }

        // =====================================================================
        // NEW: Upload to STAGING bucket with metadata
        // =====================================================================
        if ip.Config.UploadToS3 {
            // Build the staging object key (same structure as production)
            stagingKey := strings.Replace(imgFile.ObjectKey, ext, fmt.Sprintf("-%s.webp", sizeName), 1)

            // Build the target key (where Lambda will copy to)
            targetKey := stagingKey // Same key, different bucket

            // Build metadata for Lambda
            metadata := ip.BuildWebPMetadata(imgFile, sizeName, sizePixels, targetKey)

            // Upload to staging bucket with metadata
            if err := ip.uploadToStagingWithMetadata(localWebpPath, stagingKey, metadata); err != nil {
                ip.Stats.IncrementFailed(imgFile.ObjectKey, fmt.Sprintf("S3 upload failed: %v", err))
                return fmt.Errorf("failed to upload WebP to staging: %v", err)
            }

            log.Printf("Uploaded to staging: s3://%s/%s", ip.Config.StagingBucket, stagingKey)
            ip.Stats.IncrementWebPCreated()
        }
    }

    ip.Stats.IncrementProcessed(imgFile.ObjectKey)

    // Cleanup source file
    if ip.Config.AutoCleanup {
        os.Remove(sourceFilePath)
    }

    return nil
}

// uploadToStagingWithMetadata uploads a WebP file to the staging bucket with metadata
func (ip *ImageProcessor) uploadToStagingWithMetadata(localPath, objectKey string, metadata fileservice.WebPMetadata) error {
    svc := s3.New(ip.Session)

    fileBuffer, err := os.Open(localPath)
    if err != nil {
        return fmt.Errorf("failed to open local file: %v", err)
    }
    defer fileBuffer.Close()

    // Build S3 metadata map
    s3Metadata := map[string]*string{
        "batch-id":           aws.String(metadata.BatchID),
        "source-key":         aws.String(metadata.SourceKey),
        "entity-type":        aws.String(metadata.EntityType),
        "entity-contributor": aws.String(metadata.EntityContributor),
        "entity-family-id":   aws.String(metadata.EntityFamilyID),
        "entity-set-id":      aws.String(metadata.EntitySetID),
        "webp-size":          aws.String(metadata.WebPSize),
        "webp-dimensions":    aws.String(fmt.Sprintf("%d", metadata.WebPDimensions)),
        "target-bucket":      aws.String(metadata.TargetBucket),
        "target-key":         aws.String(metadata.TargetKey),
        "processed-at":       aws.String(metadata.ProcessedAt),
        "image-name":         aws.String(metadata.ImageName),
        "image-visibility":   aws.String(metadata.ImageVisibility),
        "image-access":       aws.String(metadata.ImageAccess),
    }

    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(ip.Config.StagingBucket),
        Key:         aws.String(objectKey),
        Body:        fileBuffer,
        ContentType: aws.String("image/webp"),
        Metadata:    s3Metadata,
    })
    if err != nil {
        return fmt.Errorf("failed to upload to staging: %v", err)
    }

    return nil
}
```

---

### 5. New Statistics Module

#### New File: `src/image-processor/stats.go`

```go
package imageprocessor

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
)

// ProcessingStats tracks statistics for a batch run
type ProcessingStats struct {
    mu sync.Mutex

    BatchID     string    `json:"batch_id"`
    Contributor string    `json:"contributor"`
    StartedAt   time.Time `json:"started_at"`
    CompletedAt time.Time `json:"completed_at,omitempty"`

    // Counters
    FilesFound     int `json:"files_found"`
    FilesProcessed int `json:"files_processed"`
    FilesSkipped   int `json:"files_skipped"`
    FilesFailed    int `json:"files_failed"`
    WebPCreated    int `json:"webp_created"`

    // Details
    SkippedFiles []SkippedFile `json:"skipped_files,omitempty"`
    FailedFiles  []FailedFile  `json:"failed_files,omitempty"`
}

// SkippedFile records why a file was skipped
type SkippedFile struct {
    ObjectKey string `json:"object_key"`
    Reason    string `json:"reason"`
}

// FailedFile records why a file failed processing
type FailedFile struct {
    ObjectKey string `json:"object_key"`
    Error     string `json:"error"`
}

// NewProcessingStats creates a new stats tracker
func NewProcessingStats(batchID, contributor string) *ProcessingStats {
    return &ProcessingStats{
        BatchID:      batchID,
        Contributor:  contributor,
        StartedAt:    time.Now().UTC(),
        SkippedFiles: []SkippedFile{},
        FailedFiles:  []FailedFile{},
    }
}

// SetFilesFound sets the total number of files to process
func (s *ProcessingStats) SetFilesFound(count int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.FilesFound = count
}

// IncrementProcessed increments the processed counter
func (s *ProcessingStats) IncrementProcessed(objectKey string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.FilesProcessed++
}

// IncrementSkipped increments the skipped counter and records reason
func (s *ProcessingStats) IncrementSkipped(objectKey, reason string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.FilesSkipped++
    s.SkippedFiles = append(s.SkippedFiles, SkippedFile{
        ObjectKey: objectKey,
        Reason:    reason,
    })
}

// IncrementFailed increments the failed counter and records error
func (s *ProcessingStats) IncrementFailed(objectKey, errMsg string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.FilesFailed++
    s.FailedFiles = append(s.FailedFiles, FailedFile{
        ObjectKey: objectKey,
        Error:     errMsg,
    })
}

// IncrementWebPCreated increments the WebP files created counter
func (s *ProcessingStats) IncrementWebPCreated() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.WebPCreated++
}

// Complete marks the stats as complete
func (s *ProcessingStats) Complete() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.CompletedAt = time.Now().UTC()
}

// Duration returns the processing duration
func (s *ProcessingStats) Duration() time.Duration {
    if s.CompletedAt.IsZero() {
        return time.Since(s.StartedAt)
    }
    return s.CompletedAt.Sub(s.StartedAt)
}

// Summary returns a formatted summary string
func (s *ProcessingStats) Summary() string {
    s.mu.Lock()
    defer s.mu.Unlock()

    return fmt.Sprintf(`
================================================================================
BATCH PROCESSING COMPLETE
================================================================================
Batch ID:        %s
Contributor:     %s
Duration:        %s

Files Found:     %d
Files Processed: %d
Files Skipped:   %d
Files Failed:    %d
WebP Created:    %d

Success Rate:    %.2f%%
================================================================================
`,
        s.BatchID,
        s.Contributor,
        s.Duration().Round(time.Second),
        s.FilesFound,
        s.FilesProcessed,
        s.FilesSkipped,
        s.FilesFailed,
        s.WebPCreated,
        float64(s.FilesProcessed)/float64(s.FilesFound)*100,
    )
}

// ToJSON returns the stats as JSON
func (s *ProcessingStats) ToJSON() (string, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    jsonBytes, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return "", err
    }
    return string(jsonBytes), nil
}

// WriteToFile writes the stats to a JSON file
func (s *ProcessingStats) WriteToFile(filepath string) error {
    jsonStr, err := s.ToJSON()
    if err != nil {
        return err
    }

    if err := os.MkdirAll(filepath[:len(filepath)-len("/stats.json")], os.ModePerm); err != nil {
        return err
    }

    return os.WriteFile(filepath, []byte(jsonStr), 0644)
}

// LogProgress logs current progress
func (s *ProcessingStats) LogProgress() {
    s.mu.Lock()
    defer s.mu.Unlock()

    total := s.FilesProcessed + s.FilesSkipped + s.FilesFailed
    if s.FilesFound > 0 {
        pct := float64(total) / float64(s.FilesFound) * 100
        log.Printf("Progress: %d/%d (%.1f%%) - Processed: %d, Skipped: %d, Failed: %d",
            total, s.FilesFound, pct, s.FilesProcessed, s.FilesSkipped, s.FilesFailed)
    }
}
```

---

### 6. New Manifest Module

#### New File: `src/image-processor/manifest.go`

```go
package imageprocessor

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "time"
)

// BatchManifest contains complete information about a batch run
type BatchManifest struct {
    BatchID     string         `json:"batch_id"`
    Contributor string         `json:"contributor"`
    Config      ManifestConfig `json:"config"`
    Statistics  *ProcessingStats `json:"statistics"`
    CreatedAt   time.Time      `json:"created_at"`
}

// ManifestConfig contains the config values used for this batch
type ManifestConfig struct {
    SourceBucket     string         `json:"source_bucket"`
    StagingBucket    string         `json:"staging_bucket"`
    ProductionBucket string         `json:"production_bucket"`
    WebPSizes        map[string]int `json:"webp_sizes"`
    MaxFiles         int            `json:"max_files"`
    SkipExisting     bool           `json:"skip_existing"`
    CheckDB          bool           `json:"check_db"`
    CheckS3          bool           `json:"check_s3"`
}

// NewBatchManifest creates a new manifest for a batch run
func NewBatchManifest(config *Config, stats *ProcessingStats) *BatchManifest {
    return &BatchManifest{
        BatchID:     config.BatchID,
        Contributor: config.Contributor,
        Config: ManifestConfig{
            SourceBucket:     config.SourceBucket,
            StagingBucket:    config.StagingBucket,
            ProductionBucket: config.ProductionBucket,
            WebPSizes:        config.WebpSizes,
            MaxFiles:         config.MaxFilesToProcess,
            SkipExisting:     config.SkipExistingWebP,
            CheckDB:          config.CheckDBBeforeProcess,
            CheckS3:          config.CheckS3BeforeProcess,
        },
        Statistics: stats,
        CreatedAt:  time.Now().UTC(),
    }
}

// WriteManifest writes the manifest to the configured directory
func (ip *ImageProcessor) WriteManifest() error {
    if !ip.Config.GenerateManifest {
        return nil
    }

    manifest := NewBatchManifest(ip.Config, ip.Stats)

    jsonBytes, err := json.MarshalIndent(manifest, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal manifest: %v", err)
    }

    // Create manifest directory
    manifestDir := filepath.Join(ip.Config.ManifestDir, ip.Config.BatchID)
    if err := os.MkdirAll(manifestDir, os.ModePerm); err != nil {
        return fmt.Errorf("failed to create manifest directory: %v", err)
    }

    // Write manifest file
    manifestPath := filepath.Join(manifestDir, "manifest.json")
    if err := os.WriteFile(manifestPath, jsonBytes, 0644); err != nil {
        return fmt.Errorf("failed to write manifest: %v", err)
    }

    log.Printf("Manifest written to: %s", manifestPath)
    return nil
}
```

---

## S3 Upload with Metadata

### Complete Flow

```
1. ProcessFile() called with ImageFile
2. Check existence (DB + S3) → skip if exists
3. Convert SVG → PNG → WebP (local files)
4. For each size variant:
   a. Build WebPMetadata struct
   b. Upload to staging bucket with metadata attached
5. Update statistics
6. Clean up local files
```

### Metadata Visibility

The Lambda can retrieve metadata using `HeadObject`:

```javascript
const response = await s3.headObject({
  Bucket: 'vectoricons-webp-staging',
  Key: 'iconify/icons/ABC123/DEF456/cup-preview.webp'
}).promise();

const metadata = response.Metadata;
// metadata['batch-id'] = '2024-01-15-iconify-abc12345'
// metadata['target-bucket'] = 'vectoricons-public'
// metadata['entity-type'] = 'icon'
// etc.
```

---

## Testing Strategy

### Unit Tests

1. **Config Tests** (`config_test.go`)
   - Test new field defaults
   - Test YAML parsing with new fields

2. **Existence Check Tests** (`existence_test.go`)
   - Mock DB responses
   - Mock S3 HeadObject responses
   - Test skip logic

3. **Metadata Tests** (`metadata_test.go`)
   - Test metadata building
   - Test all entity types

4. **Stats Tests** (`stats_test.go`)
   - Test thread-safety
   - Test JSON output

### Integration Tests

1. **Local Processing**
   - Process test SVGs locally
   - Verify WebP output
   - Verify no S3 calls when `is_local: true`

2. **Staging Upload**
   - Upload to test staging bucket
   - Verify metadata attached
   - Verify correct content type

3. **Existence Check Integration**
   - Pre-populate test DB with Image records
   - Verify files are skipped

### End-to-End Test

1. Create test SVG in source bucket
2. Run processor with `max_files_to_process: 1`
3. Verify WebP in staging bucket
4. Verify metadata readable
5. Manually trigger Lambda (or mock)
6. Verify WebP in production bucket
7. Verify DB record created

---

## Migration Path

### Phase 1: Code Changes (No Production Impact)

1. Add new config fields (backwards compatible)
2. Add new database methods
3. Add new file service methods
4. Add statistics module
5. Add manifest module
6. Update ProcessFile with existence checks
7. Run all unit tests

### Phase 2: Local Testing

1. Test with `is_local: true`
2. Verify existence checks work
3. Verify statistics tracking
4. Verify manifest generation

### Phase 3: Staging Bucket Setup

1. Create `vectoricons-webp-staging` bucket
2. Configure IAM policies
3. Test upload with metadata

### Phase 4: Small Production Test

1. Run with `max_files_to_process: 100`
2. Verify staging uploads
3. Verify metadata
4. Manual Lambda test

### Phase 5: Full Production Run

1. Increase batch size incrementally
2. Monitor CloudWatch metrics
3. Process all backlog files

---

## Open Questions Resolved

| Question | Decision |
|----------|----------|
| Metadata storage | S3 object metadata (not companion files) |
| Skip logic | Skip if ANY variant exists (not just all) |
| Batch ID format | `{date}-{contributor}-{uuid8}` |
| Statistics persistence | JSON file in manifest directory |

---

## Next Steps

1. Review this specification
2. Confirm metadata fields meet Lambda requirements
3. Confirm entity type mapping logic
4. Begin implementation of Sprint 1
