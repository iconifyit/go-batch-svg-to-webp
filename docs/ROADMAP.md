# Roadmap: Safe Batch WebP Conversion

This document outlines the implementation plan for safely converting the remaining ~300,000+ SVG images to WebP format while protecting existing data and infrastructure.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Current State Analysis](#current-state-analysis)
- [Target Architecture](#target-architecture)
- [Phase 1: Pre-Processing Safety Checks](#phase-1-pre-processing-safety-checks)
- [Phase 2: Configuration Updates](#phase-2-configuration-updates)
- [Phase 3: Staging Bucket Setup](#phase-3-staging-bucket-setup)
- [Phase 4: Lambda Loader](#phase-4-lambda-loader)
- [Phase 5: Validation & Monitoring](#phase-5-validation--monitoring)
- [Risk Mitigation](#risk-mitigation)
- [Rollback Plan](#rollback-plan)

---

## Architecture Overview

### Current Flow (Problematic)

```
SVG Source → Go App → Direct Upload to Production S3
                   → No DB entry created
                   → No existence check
                   → Risk of overwrites
```

### Target Flow (Safe ETL Pipeline)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        EXTRACT + TRANSFORM (Go App)                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐    ┌─────────────────┐    ┌────────────────────────┐  │
│  │ SVG Source   │───▶│ Existence Check │───▶│ Skip if WebP exists    │  │
│  │ (S3 Bucket)  │    │ (DB + S3 Query) │    │ in DB or prod bucket   │  │
│  └──────────────┘    └─────────────────┘    └───────────┬────────────┘  │
│                                                         │               │
│                                          ┌──────────────▼──────────────┐│
│                                          │     Process SVG → WebP      ││
│                                          │  (rsvg-convert + ffmpeg)    ││
│                                          └──────────────┬──────────────┘│
│                                                         │               │
│                                          ┌──────────────▼──────────────┐│
│                                          │  Upload to STAGING bucket   ││
│                                          │  with metadata manifest     ││
│                                          └──────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            LOAD (Lambda)                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────────┐    ┌─────────────────┐    ┌────────────────────┐  │
│  │ S3 Event Trigger │───▶│ Lambda Function │───▶│ 1. Validate WebP   │  │
│  │ (ObjectCreated)  │    │                 │    │ 2. Create DB entry │  │
│  └──────────────────┘    └─────────────────┘    │ 3. Copy to prod S3 │  │
│                                                  │ 4. Delete staging  │  │
│                                                  └────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Current State Analysis

### What Works
- SVG → PNG → WebP conversion pipeline
- Concurrent worker pools (download + process)
- Local and S3 file service abstraction
- Contributor validation against DB
- Configurable output sizes (thumbnail, preview, watermark)

### What's Missing
| Gap | Risk | Priority |
|-----|------|----------|
| No existence check before processing | Wasted compute, potential overwrites | **Critical** |
| No DB entry creation | Orphaned files, broken references | **Critical** |
| Direct upload to production bucket | Data integrity risk | **High** |
| No manifest/audit trail | Cannot verify or rollback | **High** |
| Dry-run not fully implemented | Cannot safely test | **Medium** |

### Data Protection Requirements
1. **Never delete** existing files in production S3
2. **Never overwrite** existing WebP files
3. **Never modify** existing DB records
4. **Always verify** before any write operation
5. **Always log** every action for audit trail

---

## Target Architecture

### S3 Buckets

| Bucket | Purpose | Access |
|--------|---------|--------|
| `vectoricons-svg-source` | Source SVG files | Read-only for Go app |
| `vectoricons-webp-staging` | **NEW** - Staging for processed WebP | Write for Go app, Read for Lambda |
| `vectoricons-public` | Production WebP hosting | Write for Lambda only |

### Database Tables

| Table | Purpose | Go App Access | Lambda Access |
|-------|---------|---------------|---------------|
| `users` | Contributor validation | Read | None |
| `images` | Image metadata | Read (existence check) | Write (new entries) |
| `icons` | Icon entities | Read (for EntityID) | Read |
| `illustrations` | Illustration entities | Read (for EntityID) | Read |
| `families` | Family entities | Read | Read |
| `sets` | Set entities | Read | Read |

---

## Phase 1: Pre-Processing Safety Checks

### 1.1 Add Existence Check to Go App

**File:** `src/image-processor/funcs.go`

Add method to check if WebP already exists:

```go
// CheckWebPExists checks both DB and S3 for existing WebP
func (ip *ImageProcessor) CheckWebPExists(imgFile imagefile.ImageFile) (bool, error) {
    // 1. Check DB for existing Image record with FileType=webp
    // 2. Check production S3 bucket for existing object
    // Return true if either exists (skip processing)
}
```

**New files needed:**
- `src/database/fn-image.go` - Add `GetImageByObjectKey()` method

### 1.2 Add Skip Logic to ProcessFile

**File:** `src/image-processor/funcs.go`

Modify `ProcessFile()` to check existence first:

```go
func (ip *ImageProcessor) ProcessFile(imgFile imagefile.ImageFile) error {
    // NEW: Check if WebP already exists
    exists, err := ip.CheckWebPExists(imgFile)
    if err != nil {
        return fmt.Errorf("existence check failed: %v", err)
    }
    if exists {
        log.Printf("SKIP: WebP already exists for %s", imgFile.ObjectKey)
        ip.Stats.Skipped++
        return nil
    }

    // ... existing processing logic ...
}
```

### 1.3 Add Processing Statistics

Track and report:
- Files found
- Files skipped (already exist)
- Files processed
- Files failed
- Processing time

---

## Phase 2: Configuration Updates

### 2.1 New Config Fields

**File:** `src/image-processor/config.go`

```go
type Config struct {
    // ... existing fields ...

    // NEW: Staging configuration
    StagingBucket        string `yaml:"staging_bucket"`
    ProductionBucket     string `yaml:"production_bucket"`

    // NEW: Safety flags
    SkipExistingWebP     bool   `yaml:"skip_existing_webp"`
    CheckDBBeforeProcess bool   `yaml:"check_db_before_process"`
    CheckS3BeforeProcess bool   `yaml:"check_s3_before_process"`

    // NEW: Manifest generation
    GenerateManifest     bool   `yaml:"generate_manifest"`
    ManifestPath         string `yaml:"manifest_path"`

    // NEW: Batch control
    MaxFilesToProcess    int    `yaml:"max_files_to_process"`
    BatchID              string `yaml:"batch_id"`
}
```

### 2.2 Production Config Template

**File:** `config-production.yml`

```yaml
# AWS Configuration
aws_region: us-east-1
role_arn: arn:aws:iam::ACCOUNT_ID:role/svg-webp-processor-role

# Source Configuration
source_bucket: vectoricons-svg-source
is_local: false

# Staging Configuration (NEW - Go app uploads here)
staging_bucket: vectoricons-webp-staging
upload_to_s3: true

# Production Configuration (Lambda copies to here)
production_bucket: vectoricons-public

# Safety Checks (NEW)
skip_existing_webp: true
check_db_before_process: true
check_s3_before_process: true

# Processing Configuration
webp_sizes:
  thumbnail: 128
  preview: 512
  watermark: 512

watermark_path: s3://vectoricons-assets/watermark.svg

# Worker Configuration
download_worker_pool_size: 5
process_worker_pool_size: 10

# Batch Control (NEW)
max_files_to_process: 10000  # Process in batches for safety
batch_id: ""  # Auto-generated if empty
generate_manifest: true
manifest_path: ./manifests/

# Cleanup
auto_cleanup: true

# Logging
logging_output: 3  # Both console and file
logfile: ./logs/batch-processor.log
```

---

## Phase 3: Staging Bucket Setup

### 3.1 Create Staging Bucket

```bash
aws s3 mb s3://vectoricons-webp-staging --region us-east-1
```

### 3.2 Bucket Structure

```
vectoricons-webp-staging/
├── manifests/
│   └── {batch_id}/
│       └── manifest.json          # List of all files in this batch
├── webp/
│   └── {contributor}/
│       └── {type}/
│           └── {familyUUID}/
│               └── {setUUID}/
│                   ├── {filename}-thumbnail.webp
│                   ├── {filename}-preview.webp
│                   └── {filename}-watermark.webp
└── metadata/
    └── {contributor}/
        └── {type}/
            └── {familyUUID}/
                └── {setUUID}/
                    └── {filename}.json    # Metadata for Lambda
```

### 3.3 Metadata File Structure

Each WebP upload should have a companion metadata file:

```json
{
  "batch_id": "2024-01-15-batch-001",
  "source_object_key": "iconify/icons/2C5AC6FCEB8D/6591D6BE211A/coffee-cup.svg",
  "webp_files": [
    {
      "size": "thumbnail",
      "object_key": "iconify/icons/2C5AC6FCEB8D/6591D6BE211A/coffee-cup-thumbnail.webp",
      "dimensions": 128
    },
    {
      "size": "preview",
      "object_key": "iconify/icons/2C5AC6FCEB8D/6591D6BE211A/coffee-cup-preview.webp",
      "dimensions": 512
    },
    {
      "size": "watermark",
      "object_key": "iconify/icons/2C5AC6FCEB8D/6591D6BE211A/coffee-cup-watermark.webp",
      "dimensions": 512
    }
  ],
  "entity_type": "icon",
  "entity_lookup": {
    "contributor": "iconify",
    "family_unique_id": "2C5AC6FCEB8D",
    "set_unique_id": "6591D6BE211A",
    "filename": "coffee-cup"
  },
  "processed_at": "2024-01-15T10:30:00Z"
}
```

### 3.4 IAM Policy for Go App

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadSourceBucket",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::vectoricons-svg-source",
        "arn:aws:s3:::vectoricons-svg-source/*"
      ]
    },
    {
      "Sid": "WriteStagingBucket",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::vectoricons-webp-staging",
        "arn:aws:s3:::vectoricons-webp-staging/*"
      ]
    },
    {
      "Sid": "ReadProductionBucketForExistenceCheck",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:HeadObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::vectoricons-public",
        "arn:aws:s3:::vectoricons-public/*"
      ]
    }
  ]
}
```

---

## Phase 4: Lambda Loader

### 4.1 Lambda Function Overview

**Purpose:** Process staged WebP files and load them into production

**Trigger:** S3 ObjectCreated event on `vectoricons-webp-staging/metadata/*.json`

**Actions:**
1. Read metadata JSON
2. Validate all WebP files exist in staging
3. For each WebP file:
   - Check production S3 doesn't already have it
   - Check DB doesn't already have Image record
   - Copy WebP to production bucket
   - Create Image record in DB
4. Delete staging files (webp + metadata)
5. Log success/failure to CloudWatch

### 4.2 Lambda Pseudocode

```javascript
// lambda/webp-loader/index.js

exports.handler = async (event) => {
  const s3Event = event.Records[0].s3;
  const metadataKey = s3Event.object.key;

  // 1. Read metadata
  const metadata = await readMetadataFromS3(metadataKey);

  // 2. Validate WebP files exist in staging
  for (const webpFile of metadata.webp_files) {
    const exists = await s3HeadObject(STAGING_BUCKET, webpFile.object_key);
    if (!exists) {
      throw new Error(`Missing WebP file: ${webpFile.object_key}`);
    }
  }

  // 3. Check production doesn't already have these files
  for (const webpFile of metadata.webp_files) {
    const prodExists = await s3HeadObject(PRODUCTION_BUCKET, webpFile.object_key);
    if (prodExists) {
      console.log(`SKIP: Already exists in production: ${webpFile.object_key}`);
      continue;
    }

    // 4. Look up entity ID from DB
    const entityId = await lookupEntityId(metadata.entity_lookup);
    if (!entityId) {
      throw new Error(`Entity not found for: ${JSON.stringify(metadata.entity_lookup)}`);
    }

    // 5. Copy to production
    await s3CopyObject(
      STAGING_BUCKET, webpFile.object_key,
      PRODUCTION_BUCKET, webpFile.object_key
    );

    // 6. Create DB record
    await createImageRecord({
      entity_id: entityId,
      entity_type: metadata.entity_type,
      file_type: 'webp',
      object_key: webpFile.object_key,
      name: `${metadata.entity_lookup.filename}-${webpFile.size}`,
      url: `https://cdn.vectoricons.com/${webpFile.object_key}`,
      visibility: 'public',
      access: 'all'
    });
  }

  // 7. Clean up staging
  await deleteFromStaging(metadata);

  return { status: 'success', processed: metadata.webp_files.length };
};
```

### 4.3 Lambda IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ReadStagingBucket",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:HeadObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::vectoricons-webp-staging",
        "arn:aws:s3:::vectoricons-webp-staging/*"
      ]
    },
    {
      "Sid": "WriteProductionBucket",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:HeadObject"
      ],
      "Resource": [
        "arn:aws:s3:::vectoricons-public",
        "arn:aws:s3:::vectoricons-public/*"
      ]
    },
    {
      "Sid": "DatabaseAccess",
      "Effect": "Allow",
      "Action": [
        "rds-db:connect"
      ],
      "Resource": "arn:aws:rds-db:us-east-1:ACCOUNT_ID:dbuser:*/lambda_user"
    }
  ]
}
```

### 4.4 S3 Event Configuration

```json
{
  "LambdaFunctionConfigurations": [
    {
      "Id": "WebPLoaderTrigger",
      "LambdaFunctionArn": "arn:aws:lambda:us-east-1:ACCOUNT_ID:function:webp-loader",
      "Events": ["s3:ObjectCreated:*"],
      "Filter": {
        "Key": {
          "FilterRules": [
            {
              "Name": "prefix",
              "Value": "metadata/"
            },
            {
              "Name": "suffix",
              "Value": ".json"
            }
          ]
        }
      }
    }
  ]
}
```

---

## Phase 5: Validation & Monitoring

### 5.1 Pre-Run Validation Checklist

Before running batch processing:

- [ ] Verify source bucket accessible
- [ ] Verify staging bucket exists and writable
- [ ] Verify production bucket accessible (read-only check)
- [ ] Verify DB connection works
- [ ] Verify contributor exists in DB
- [ ] Run dry-run on small sample (100 files)
- [ ] Verify Lambda is deployed and triggered correctly
- [ ] Verify CloudWatch alarms are configured

### 5.2 Batch Processing Manifest

Each batch run should generate:

```json
{
  "batch_id": "2024-01-15-batch-001",
  "started_at": "2024-01-15T10:00:00Z",
  "completed_at": "2024-01-15T10:45:00Z",
  "contributor": "iconify",
  "config": {
    "source_bucket": "vectoricons-svg-source",
    "staging_bucket": "vectoricons-webp-staging",
    "max_files": 10000
  },
  "statistics": {
    "files_found": 10000,
    "files_skipped_existing": 2500,
    "files_processed": 7500,
    "files_failed": 3,
    "webp_files_created": 22500
  },
  "failed_files": [
    {
      "object_key": "iconify/icons/ABC123/DEF456/broken-file.svg",
      "error": "rsvg-convert failed: invalid SVG"
    }
  ]
}
```

### 5.3 CloudWatch Metrics

| Metric | Description |
|--------|-------------|
| `WebPProcessor/FilesProcessed` | Count of files processed |
| `WebPProcessor/FilesSkipped` | Count of files skipped (already exist) |
| `WebPProcessor/FilesFailed` | Count of processing failures |
| `WebPProcessor/ProcessingTime` | Time per file |
| `WebPLoader/FilesLoaded` | Lambda: files copied to production |
| `WebPLoader/DBRecordsCreated` | Lambda: DB records created |
| `WebPLoader/Errors` | Lambda: processing errors |

### 5.4 CloudWatch Alarms

| Alarm | Threshold | Action |
|-------|-----------|--------|
| High failure rate | >1% files failed | SNS notification |
| Lambda errors | >0 errors in 5 min | SNS notification |
| Staging bucket growth | >100GB | SNS notification |
| Processing stopped | 0 files in 30 min | SNS notification |

---

## Risk Mitigation

### Data Protection Measures

| Risk | Mitigation |
|------|------------|
| Overwrite existing WebP | Existence check before processing; HeadObject before copy |
| Corrupt production data | Staging bucket + Lambda validation |
| Orphaned DB records | Lambda creates record AFTER successful S3 copy |
| Orphaned S3 files | Lambda creates S3 copy BEFORE DB record (recoverable) |
| Wrong entity association | Lambda validates entity exists before creating record |
| Partial batch failure | Manifest tracks all files; can resume from failure point |

### Safety Flags in Go App

```go
// These flags MUST be true for production runs
type SafetyConfig struct {
    SkipExistingWebP     bool  // Skip if WebP exists anywhere
    CheckDBBeforeProcess bool  // Query DB for existing Image record
    CheckS3BeforeProcess bool  // HeadObject on production bucket
    UseStaginBucket      bool  // Never write directly to production
    GenerateManifest     bool  // Always create audit trail
}
```

### Never Do List

1. **NEVER** give Go app write access to production bucket
2. **NEVER** give Go app write access to database
3. **NEVER** skip existence checks in production
4. **NEVER** process without manifest generation
5. **NEVER** delete from staging until Lambda confirms success

---

## Rollback Plan

### If Go App Produces Bad Output

1. Stop Go app processing
2. Do NOT trigger Lambda (or disable S3 event)
3. Review staging bucket contents
4. Delete bad files from staging
5. Fix issue and reprocess

### If Lambda Creates Bad DB Records

1. Disable Lambda trigger
2. Query `images` table for recent records with batch_id
3. Soft-delete affected records (`is_deleted = true`)
4. Do NOT delete S3 files (they're valid)
5. Fix Lambda logic
6. Reprocess affected files

### If Lambda Copies Wrong Files to Production

1. Disable Lambda trigger
2. Query manifest for affected files
3. S3 versioning allows recovery of previous versions
4. Or delete new files if they didn't exist before
5. Soft-delete DB records

---

## Implementation Order

### Sprint 1: Safety First
1. [ ] Add existence check to Go app (DB query)
2. [ ] Add existence check to Go app (S3 HeadObject)
3. [ ] Add skip logic to ProcessFile
4. [ ] Add processing statistics
5. [ ] Add manifest generation
6. [ ] Update config with new fields
7. [ ] Test with dry-run on 100 files

### Sprint 2: Staging Infrastructure
1. [ ] Create staging S3 bucket
2. [ ] Configure IAM policies
3. [ ] Update Go app to upload to staging
4. [ ] Add metadata file generation
5. [ ] Test staging upload with 1000 files

### Sprint 3: Lambda Loader
1. [ ] Create Lambda function
2. [ ] Configure S3 event trigger
3. [ ] Configure DB access (RDS Proxy recommended)
4. [ ] Test with manual trigger
5. [ ] Test with S3 event trigger
6. [ ] Add CloudWatch metrics

### Sprint 4: Production Run
1. [ ] Set up CloudWatch alarms
2. [ ] Run batch of 10,000 files
3. [ ] Validate results
4. [ ] Scale up batch size
5. [ ] Monitor and iterate

---

## Success Criteria

- [ ] Zero existing files overwritten
- [ ] Zero existing DB records modified
- [ ] 100% of processed files have DB records
- [ ] 100% of DB records have valid S3 files
- [ ] Complete audit trail in manifests
- [ ] <1% failure rate
- [ ] All 300,000+ backlog files processed

---

## Open Questions

1. **Batch size per run?** Recommend 10,000 files per batch for manageability
2. **Concurrent batches?** Start with 1, scale to 3-5 if Lambda handles load
3. **Priority order?** Process by contributor? By date? By product type?
4. **Notification preferences?** Email? Slack? PagerDuty for critical errors?
5. **S3 versioning on production bucket?** Recommended for rollback capability
