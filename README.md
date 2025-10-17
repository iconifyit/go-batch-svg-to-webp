# High-Performance Batch Image Processor

A production-grade CLI tool built in Go to batch-convert 500,000+ SVG images to optimized WebP format at multiple sizes. Originally developed for [VectorIcons.com](https://vectoricons.com), a multi-vendor marketplace for vector illustrations and icons.

## The Problem

When VectorIcons launched, only PNG conversions were performed at upload time to save on development time. Later, we needed to generate multiple WebP versions retroactively for CDN delivery, previews, and browser compatibility across 500,000+ existing images.

Initial estimates using Node.js suggested 11.5 days of processing time. Go's concurrency model reduced this to **45 minutes** - a **16x performance improvement**.

---

## Technology Stack

- **Language:** Go 1.22 with goroutines and channels
- **Cloud:** AWS S3, AWS STS (IAM role assumption)
- **Database:** PostgreSQL with GORM
- **Image Processing:** rsvg-convert (SVG→PNG), ffmpeg (PNG→WebP, watermarking)
- **Configuration:** YAML-based with runtime validation

---

## Performance Results

### Single-Threaded Baseline

I first built a single-threaded version to establish a baseline, processing 4,500 test files:

| Metric | Value |
|--------|-------|
| Files processed | 4,500 files |
| Total time | 7 minutes 17 seconds |
| Throughput | **10.4 files/sec** |
| Extrapolated for 500K files | **13.4 hours** |

### Concurrent Implementation (10 Workers)

After adding Go's worker pool pattern with 10 concurrent goroutines:

| Metric | Value |
|--------|-------|
| Files processed | 4,500 files |
| Total time | 27 seconds |
| Throughput | **166.67 files/sec** |
| Extrapolated for 500K files | **50 minutes** |

### Language Comparison

| Language | Est. Time per Image | Total Runtime (500K files) | vs Go |
|----------|---------------------|----------------------------|-------|
| **Go** | ~1ms | **~45 minutes** | 1x |
| Python | ~50ms | ~41.6 hours | 55x slower |
| Node.js | ~200ms | ~11.5 days | 368x slower |

**Result:** Go's lightweight goroutines and native parallelism delivered a **16x improvement** over single-threaded execution.

---

## Architecture

The application follows a **producer-consumer pattern** with **dual worker pools**, using Go's concurrency primitives (goroutines, channels, and WaitGroups) to achieve high throughput.

```mermaid
graph TB
    START([CLI Entry Point<br/>--prefix --config]) --> CONFIG[Load YAML Config]

    CONFIG --> VALIDATE[Validate Contributor<br/>in PostgreSQL]
    VALIDATE --> ROLE[Assume AWS IAM Role<br/>via STS]
    ROLE --> FILESERVICE{Initialize<br/>FileService}

    FILESERVICE -->|IsLocal=true| LOCAL[LocalFileService]
    FILESERVICE -->|IsLocal=false| S3FS[S3FileService]

    LOCAL --> LISTFILES[List and Parse Files]
    S3FS --> LISTFILES

    LISTFILES --> IMAGEFILES[Create ImageFile Objects<br/>Parse Path Structure]
    IMAGEFILES --> QUEUES[Initialize Buffered Channels]

    QUEUES --> DOWNLOADQ[Download Queue<br/>chan ImageFile]
    QUEUES --> PROCESSQ[Process Queue<br/>chan ImageFile]

    DOWNLOADQ --> DLW1[Download Worker 1]
    DOWNLOADQ --> DLW2[Download Worker 2]
    DOWNLOADQ --> DLW3[Download Worker 3]
    DOWNLOADQ --> DLWN[Download Worker N]

    DLW1 --> FETCH1[Fetch from Source]
    DLW2 --> FETCH2[Fetch from Source]
    DLW3 --> FETCH3[Fetch from Source]
    DLWN --> FETCHN[Fetch from Source]

    FETCH1 --> PROCESSQ
    FETCH2 --> PROCESSQ
    FETCH3 --> PROCESSQ
    FETCHN --> PROCESSQ

    PROCESSQ --> PW1[Process Worker 1]
    PROCESSQ --> PW2[Process Worker 2]
    PROCESSQ --> PW3[Process Worker 3]
    PROCESSQ --> PWN[Process Worker N]

    PW1 --> SIZES[For Each Size<br/>thumbnail, preview, watermark]
    PW2 --> SIZES
    PW3 --> SIZES
    PWN --> SIZES
    SIZES --> SVG2PNG[SVG to PNG<br/>rsvg-convert]

    SVG2PNG --> CHECKWM{Watermark<br/>Size?}
    CHECKWM -->|Yes| APPLYWM[Apply Watermark<br/>ffmpeg overlay filter]
    CHECKWM -->|No| PNG2WEBP[PNG to WebP<br/>ffmpeg -q:v 75]
    APPLYWM --> PNG2WEBP

    PNG2WEBP --> UPLOADCHECK{Upload<br/>to S3?}
    UPLOADCHECK -->|Yes| S3UPLOAD[S3 PutObject<br/>Target Bucket]
    UPLOADCHECK -->|No| LOCALSAVE[Save to Local<br/>Output Directory]

    S3UPLOAD --> CLEANUP[Cleanup Temp Files<br/>source, intermediate]
    LOCALSAVE --> CLEANUP

    CLEANUP --> DONE([Processing Complete<br/>Log Results])

    style START fill:#e1f5ff,stroke:#0066cc
    style FILESERVICE fill:#fff4e1,stroke:#cc9900
    style DOWNLOADQ fill:#f0e1ff,stroke:#9933cc
    style PROCESSQ fill:#f0e1ff,stroke:#9933cc
    style DLWORKERS fill:#FF9900,stroke:#cc7a00,color:#fff
    style PROCWORKERS fill:#FF9900,stroke:#cc7a00,color:#fff
    style SIZES fill:#e1ffe1,stroke:#009933
    style VALIDATE fill:#ffe1e1,stroke:#cc0000
```

---

## How It Works

### Dual Worker Pool Architecture

The processor uses two independent worker pools to parallelize I/O-bound (downloading) and CPU-bound (processing) operations:

1. **Download Workers** fetch files from source (S3 or local filesystem) and add them to the Process Queue
2. **Process Workers** pull from the queue and execute the image transformation pipeline
3. **Buffered Channels** connect the pools, enabling continuous processing without blocking

### Processing Pipeline

For each image, the processor generates multiple sizes (thumbnail: 128px, preview: 512px, watermark: 512px):

1. **SVG → PNG Conversion** using `rsvg-convert` at target dimensions
2. **Optional Watermarking** using ffmpeg's overlay filter for preview images
3. **PNG → WebP Conversion** using ffmpeg with quality optimization (`-q:v 75`)
4. **Upload** to S3 bucket or save to local output directory

### Concurrency Model

```go
// Buffered channels for work distribution
DownloadQueue := make(chan ImageFile, len(files))
ProcessQueue  := make(chan ImageFile, len(files))

// Configurable worker pools
for i := 0; i < downloadWorkers; i++ {
    go downloadWorker(DownloadQueue, ProcessQueue)
}
for i := 0; i < processWorkers; i++ {
    go processWorker(ProcessQueue)
}

// Synchronization with WaitGroups
downloadWG.Wait() // Wait for all downloads
close(ProcessQueue)
processWG.Wait()  // Wait for all processing
```

---

## Key Features

- **Highly Concurrent:** Configurable worker pools optimize for I/O and CPU workloads
- **Storage Flexibility:** Supports both local filesystem and AWS S3 (via strategy pattern)
- **Production-Ready:** Database validation, comprehensive logging, automatic cleanup
- **Config-Driven:** YAML configuration with sensible defaults
- **Hardware Acceleration:** Optional VideoToolbox support for ffmpeg on macOS
- **Modular Design:** Interface-based architecture with dependency injection

---

## Design Patterns

- **Producer-Consumer:** Decouples file discovery from processing via buffered channels
- **Strategy Pattern:** Abstract file service enables runtime switching between local/S3 backends
- **Worker Pool:** Limits concurrency to prevent resource exhaustion
- **Pipeline:** Sequential transformation stages (SVG → PNG → WebP) with conditional watermarking

---

## Installation & Usage

### Prerequisites

```bash
# macOS
brew install librsvg ffmpeg

# Ubuntu
sudo apt-get install librsvg2-bin ffmpeg
```

### Build

```bash
go build -o image-processor ./src/image-processor
```

### Configuration

Create a `config.yaml` file:

```yaml
aws_region: us-east-1
source_bucket: png-image-source-bucket
target_bucket: webp-output-target-bucket
role_arn: arn:aws:iam::111111111111:role/svg-webp-app-role

# Worker pool configuration
download_worker_pool_size: 5
process_worker_pool_size: 10

# Output sizes
webp_sizes:
  thumbnail: 128
  preview: 512
  watermark: 512

# Optional
watermark_path: /path/to/watermark.svg
use_hardware_acceleration: true
auto_cleanup: true
```

### Run

```bash
./image-processor --prefix=contributor-name --config=config.yaml
```

---

## Project Structure

```
go-batch-svg-to-webp/
├── src/
│   ├── image-processor/    # Main orchestrator & CLI
│   ├── file-service/       # Storage abstraction (Local/S3)
│   ├── image-file/         # Image metadata parser
│   ├── database/           # PostgreSQL integration
│   ├── models/             # GORM data models
│   └── common/             # Shared utilities
├── test/                   # Test fixtures
└── config-example.yml      # Configuration template
```

---

## License

MIT License. See LICENSE for details.

## Disclaimer

This software is provided "as is" without warranty of any kind. You are responsible for testing it in your environment and ensuring it meets your needs.
