# Image Processor CLI (Go)

This repository contains a high-performance image processing tool built in Go. It was originally developed to batch-convert over 500,000 SVG images to optimized WebP format for [Vectopus.com](https://vectopus.com), a multi-vendor marketplace for vector illustrations and icons.

## 🚀 Overview

When Vectopus launched, only PNG conversions were performed at upload time to save on development time. Later, we needed to generate multiple WebP versions retroactively for CDN delivery, previews, and browser compatibility.

This CLI tool:
- Processes millions of images across nested directories
- Converts SVG → PNG using rsvg-convert
- Optionally adds watermarks
- Converts PNG → WebP using ffmpeg
- Uploads results to AWS S3 or saves them locally

Go was chosen for its speed and concurrency model, reducing processing time from an estimated 11.5 days (Node.js) to just 45 minutes.

For a detailed view of the assumptions and calculations used to estimate performance, see the [image-processing-benchmark.md](image-processing-benchmark.md) file.

## ⚙️ Features

- Highly concurrent: configurable worker pool
- Supports both local and S3 file systems
- Automatic directory creation and cleanup
- Config-driven: supports YAML or JSON config files
- Robust logging and error reporting
- Modular design with interchangeable file service backends

## NOTE: 
This is a work-in-progress and should **not** be used for production installations. It was built for a specific task. Some items may not be relevant to your setup and purposes (e.g., contributor). Where-ever you see `contributor` you can most likely change it to `prefix`. 

## 📦 Installation

Clone the repository and build the CLI:

```bash
go build -o image-processor ./src/image-processor
```

Ensure you have the following system dependencies installed:

- rsvg-convert
- ffmpeg

On macOS:

```bash
brew install librsvg ffmpeg
```
On Ubuntu:

```bash
sudo apt-get install librsvg2-bin ffmpeg
```

## 🔧 Configuration

Create a config file in YAML or JSON format. Example:

config.yaml

```yaml
# Target AWS Region
aws_region: us-east-1

# Source bucket name
# source_bucket: vectopus-webp-test
source_bucket: png-image-source-bucket

# Target bucket name
target_bucket: webp-output-target-bucket
# target_bucket: image-engine-public-prod

# Prefixes to include
include_prefixes:
  - bucket-prefix-one
  - bucket-prefix-two
  - bucket-prefix-three

Prefixes to exclude
omit_prefixes:
  - omit-me-one
  - omit-me-two
  - omit-me-three

# Dry run?
dry_run: true

# Local?
is_local: true

# Upload results to s3?
upload_to_s3: true

local_source: /path/to/local/test/input
local_target: /path/to/local/test/output

# Archive structure
archive_structure: contributor,family,set,icons

# Garbage collection
auto_cleanup: false

# ffMpeg path
ffmpegPath: /opt/homebrew/bin/ffmpeg

# Webp sizes
webp_sizes:
  thumbnail: 128
  preview: 512
  watermark: 512

# Watermark
watermark_path: /path/to/local/test/assets/watermark.svg

# AWS Role arn
role_arn: arn:aws:iam::111111111111:role/svg-webp-app-role

# Logging output
# 0 = no output
# 1 = output to console
# 2 = output to file
# 3 = output to both
logging_output: 3

# Log file path
logfile: ./output.log

# Work dir
work_dir: ""

# Output dir
output_dir: ./test/output

# Worker Count
worker_pool_size: 10
download_worker_pool_size: 5
process_worker_pool_size: 10

# Use Hardware Acceleration
use_hardware_acceleration: true
``` 

## 🖼️ Example Usage

Run the processor with:

```bash
./image-processor --contributor=iconify --config=config.yaml
```

This will:

1. Load all SVG files from the configured source
2. Spawn N workers to convert and process them
3. Upload the resulting WebP images to S3 (or save locally)

## 🧠 Architecture

The processor uses a producer-consumer pattern:

- A single producer enumerates all files and adds them to a buffered job queue
- A configurable number of workers pull from the queue and execute ProcessFile()
- Each ProcessFile() call:
  - Ensures directories exist
  - Converts SVG → PNG
  - Optionally applies a watermark
  - Converts to WebP
  - Uploads result to S3 or local output

A sync.WaitGroup blocks the main thread until all workers complete, and a shared error channel captures any issues.

## 📈 Performance

| Language | Est. Time per Image | Total Runtime (500k files x 5 variants) |
|----------|---------------------|-----------------------------------------|
| Node.js  | ~200ms              | ~11.5 days                              |
| Python   | ~50ms               | ~41.6 hours                             |
| Go       | ~1ms                | ~45 minutes                             |

Go's lightweight goroutines and native parallelism make it an ideal tool for high-throughput CLI applications like this.

## 🛠️ Roadmap

- [ ] Add support for AVIF conversion  
- [ ] Optional caching of intermediate files  
- [ ] CLI flag to dry-run or list targets without processing  
- [ ] Plugin system for new output formats or storage backends  
- [ ] Dockerfile for deployment convenience  

## 🤝 Contributing

Contributions are welcome! To get started:

1. Fork this repo  
2. Create a new branch (git checkout -b feature-name)  
3. Commit your changes  
4. Open a pull request  

Feel free to file issues or suggest enhancements as well.

## 📄 License

MIT License. See LICENSE for details.

## Disclaimer

This software is provided “as is” without warranty of any kind. You are responsible for testing it in your environment and ensuring it meets your needs. The authors and maintainers are not liable for any loss of data, outages, or other damage resulting from use.
