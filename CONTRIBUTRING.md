# Contributing to Image Processor CLI

First off, thanks for your interest in contributing! This project is designed to be fast, modular, and useful for batch processing image files at scale using Go. Whether you're fixing bugs, improving performance, or adding support for new formats—we welcome your input.

## Getting Started

1. **Fork the repository** and clone your fork locally.

   git clone https://github.com/your-username/go-batch-svg-to-webp.git  
   cd go-batch-svg-to-webp

2. **Create a new branch** for your feature or fix:

   git checkout -b feature/your-feature-name

3. **Install system dependencies** (required by the processor):
   - rsvg-convert
   - ffmpeg

   On macOS:

   brew install librsvg ffmpeg

4. **Build the CLI tool:**

   go build -o image-processor ./src/image-processor

5. **Test your changes locally.** Feel free to add or improve logging.

---

## Pull Request Guidelines

- All changes must go through a pull request (no direct pushes to `develop`).
- Keep commits focused and atomic. Use meaningful commit messages.
- Write clear descriptions in your PRs: what was changed and why.
- If adding a new feature, please document how it works.
- If fixing a bug, describe how you reproduced it.

---

## Code Style

This is a Go project, so:
- Use go fmt ./... before committing
- Avoid unnecessary dependencies
- Keep code modular and testable

---

## Feature Ideas

We're especially open to:
- AVIF format support
- Dry-run mode for debugging
- Plugin architecture for new conversion targets
- Docker support for easier cross-platform use

---

## Questions?

Use GitHub Discussions for general Q&A, or open an issue if you believe something is broken.

Thanks for contributing to the project!