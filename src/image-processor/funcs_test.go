package imageprocessor

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

// realisticSVG is a minimal but valid icon SVG used as a conversion fixture.
const realisticSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
  <circle cx="12" cy="12" r="10" fill="#333"/>
  <rect x="8" y="8" width="8" height="8" fill="#fff"/>
</svg>`

// TestImageProcessor_CopyFile verifies byte-for-byte copying into the work
// directory.
func TestImageProcessor_CopyFile(t *testing.T) {
	// Scenario: copy a downloaded SVG into the processing source dir.
	dir := t.TempDir()
	src := filepath.Join(dir, "coffee-cup.svg")
	dst := filepath.Join(dir, "coffee-cup-copy.svg")
	if err := os.WriteFile(src, []byte(realisticSVG), 0644); err != nil {
		t.Fatalf("failed to seed source: %v", err)
	}

	ip := &ImageProcessor{Config: &Config{}}
	if err := ip.CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read copy: %v", err)
	}
	if string(got) != realisticSVG {
		t.Errorf("copied content differs from source")
	}
}

// TestImageProcessor_CopyFile_MissingSource verifies the error contract.
func TestImageProcessor_CopyFile_MissingSource(t *testing.T) {
	// Scenario: the source file was already cleaned up.
	dir := t.TempDir()
	ip := &ImageProcessor{Config: &Config{}}
	if err := ip.CopyFile(filepath.Join(dir, "missing.svg"), filepath.Join(dir, "out.svg")); err == nil {
		t.Fatal("CopyFile() with missing source: expected error, got nil")
	}
}

// TestRenameFolderWithTimestamp verifies the folder is renamed to a
// timestamp-formatted name in the same parent, preserving contents.
func TestRenameFolderWithTimestamp(t *testing.T) {
	// Scenario: a completed output folder is archived under a timestamp name.
	// The timestamp itself is wall-clock dependent, so the assertion pins the
	// documented name format rather than an exact time.
	parent := t.TempDir()
	folder := filepath.Join(parent, "output")
	if err := os.MkdirAll(folder, 0755); err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}
	inner := filepath.Join(folder, "coffee-cup-thumbnail.webp")
	if err := os.WriteFile(inner, []byte("RIFFxxxxWEBP"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	ip := &ImageProcessor{Config: &Config{}}
	if err := ip.RenameFolderWithTimestamp(folder); err != nil {
		t.Fatalf("RenameFolderWithTimestamp() error = %v", err)
	}

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Errorf("original folder still exists after rename")
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatalf("failed to read parent dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("parent contains %d entries, want exactly 1 renamed folder", len(entries))
	}
	timestampName := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}$`)
	renamed := entries[0].Name()
	if !timestampName.MatchString(renamed) {
		t.Errorf("renamed folder = %q, want YYYY-MM-DD-HH-MM-SS format", renamed)
	}

	// The folder's contents must survive the rename.
	if _, err := os.Stat(filepath.Join(parent, renamed, "coffee-cup-thumbnail.webp")); err != nil {
		t.Errorf("renamed folder lost its contents: %v", err)
	}
}

// TestRenameFolderWithTimestamp_MissingFolder verifies the error contract.
func TestRenameFolderWithTimestamp_MissingFolder(t *testing.T) {
	// Scenario: the folder to archive does not exist.
	ip := &ImageProcessor{Config: &Config{}}
	if err := ip.RenameFolderWithTimestamp(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("RenameFolderWithTimestamp() with missing folder: expected error, got nil")
	}
}

// TestConvertSVGToPNG is an integration test for the rsvg-convert pipeline
// stage. Skips when rsvg-convert is not installed.
func TestConvertSVGToPNG(t *testing.T) {
	// Scenario: a 24x24 icon SVG is rasterized to a 128px PNG.
	if _, err := exec.LookPath("rsvg-convert"); err != nil {
		t.Skip("integration test: rsvg-convert not installed")
	}

	dir := t.TempDir()
	svgPath := filepath.Join(dir, "coffee-cup.svg")
	pngPath := filepath.Join(dir, "coffee-cup-thumbnail.png")
	if err := os.WriteFile(svgPath, []byte(realisticSVG), 0644); err != nil {
		t.Fatalf("failed to seed SVG: %v", err)
	}

	ip := &ImageProcessor{Config: &Config{}}
	if err := ip.ConvertSVGToPNG(svgPath, pngPath, 128); err != nil {
		t.Fatalf("ConvertSVGToPNG() error = %v", err)
	}

	content, err := os.ReadFile(pngPath)
	if err != nil {
		t.Fatalf("PNG not created: %v", err)
	}
	if !bytes.HasPrefix(content, []byte("\x89PNG")) {
		t.Errorf("output is not a PNG (magic bytes = %x)", content[:min(4, len(content))])
	}
}

// TestRunFFmpeg is an integration test for the PNG-to-WebP pipeline stage,
// chained after a real rsvg-convert rasterization. Skips when either binary
// is not installed.
func TestRunFFmpeg(t *testing.T) {
	// Scenario: the intermediate 128px PNG is converted to WebP.
	if _, err := exec.LookPath("rsvg-convert"); err != nil {
		t.Skip("integration test: rsvg-convert not installed")
	}
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("integration test: ffmpeg not installed")
	}

	dir := t.TempDir()
	svgPath := filepath.Join(dir, "coffee-cup.svg")
	pngPath := filepath.Join(dir, "coffee-cup-thumbnail.png")
	webpPath := filepath.Join(dir, "coffee-cup-thumbnail.webp")
	if err := os.WriteFile(svgPath, []byte(realisticSVG), 0644); err != nil {
		t.Fatalf("failed to seed SVG: %v", err)
	}

	ip := &ImageProcessor{Config: &Config{FFmpegPath: ffmpegPath, UseHardwareAcceleration: false}}
	if err := ip.ConvertSVGToPNG(svgPath, pngPath, 128); err != nil {
		t.Fatalf("ConvertSVGToPNG() error = %v", err)
	}
	if err := ip.RunFFmpeg(pngPath, webpPath); err != nil {
		t.Fatalf("RunFFmpeg() error = %v", err)
	}

	content, err := os.ReadFile(webpPath)
	if err != nil {
		t.Fatalf("WebP not created: %v", err)
	}
	// WebP container: "RIFF" .... "WEBP"
	if !bytes.HasPrefix(content, []byte("RIFF")) || len(content) < 12 || string(content[8:12]) != "WEBP" {
		t.Errorf("output is not a WebP container (first 12 bytes = %x)", content[:min(12, len(content))])
	}
}
