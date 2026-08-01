package imageprocessor

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	fileservice "github.com/iconifyit/go-batch-svg-to-webp/src/file-service"
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
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

// TestDownloadFile verifies the downloaded file lands at
// <work_dir>/<uuid>/source/<object_key> - the exact path ProcessFile reads
// from - and that download errors propagate instead of being dropped.
func TestDownloadFile(t *testing.T) {
	// Scenario: local-mode download of a seeded source SVG into the work dir.
	sourceRoot := t.TempDir()
	workDir := t.TempDir()
	objectKey := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg"

	fullSource := filepath.Join(sourceRoot, objectKey)
	if err := os.MkdirAll(filepath.Dir(fullSource), 0755); err != nil {
		t.Fatalf("failed to create fixture dirs: %v", err)
	}
	if err := os.WriteFile(fullSource, []byte(realisticSVG), 0644); err != nil {
		t.Fatalf("failed to seed source: %v", err)
	}

	img, err := imagefile.NewImageFile(objectKey)
	if err != nil || img == nil {
		t.Fatalf("fixture parse failed: %v", err)
	}

	ip := &ImageProcessor{
		UUID:        "02b5e8da-a37b-4666-9892-44706466438e",
		Config:      &Config{IsLocal: true, WorkDir: workDir, LocalSource: sourceRoot},
		FileService: &fileservice.LocalFileService{SourceRoot: sourceRoot},
	}

	localPath, err := ip.downloadFile(img)
	if err != nil {
		t.Fatalf("downloadFile() error = %v", err)
	}
	want := filepath.Join(workDir, ip.UUID, "source", objectKey)
	if localPath != want {
		t.Errorf("downloadFile() = %q, want %q", localPath, want)
	}
	content, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("downloaded file missing at expected path: %v", err)
	}
	if string(content) != realisticSVG {
		t.Errorf("downloaded content differs from source")
	}

	// Scenario: the source file is missing - the error must propagate.
	missing, err := imagefile.NewImageFile("iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg")
	if err != nil || missing == nil {
		t.Fatalf("fixture parse failed: %v", err)
	}
	if _, err := ip.downloadFile(missing); err == nil {
		t.Error("downloadFile() with missing source: expected error, got nil")
	}
}

// TestDownloadFile_RejectsTraversal verifies an object key with parent
// traversal segments is rejected before any file is written, so a crafted
// key cannot escape the run's work directory.
func TestDownloadFile_RejectsTraversal(t *testing.T) {
	// Scenario: an untrusted key attempts to climb out of the source dir.
	ip := &ImageProcessor{
		UUID:        "02b5e8da-a37b-4666-9892-44706466438e",
		Config:      &Config{IsLocal: true, WorkDir: t.TempDir(), LocalSource: t.TempDir()},
		FileService: &fileservice.LocalFileService{SourceRoot: t.TempDir()},
	}
	evil := &imagefile.ImageFile{
		ObjectKey: "../../icons/2C11DB2D5F79/B24091F3DF3E/evil.svg",
		IsValid:   true,
	}
	if _, err := ip.downloadFile(evil); err == nil {
		t.Fatal("downloadFile() with traversal key: expected error, got nil")
	}
}

// TestCleanup verifies AutoCleanup removes this run's source and
// intermediate directories while preserving the output directory, and that
// cleanup is a no-op when AutoCleanup is disabled.
func TestCleanup(t *testing.T) {
	// Scenario: a completed run with files in all three per-run directories.
	seedRun := func(t *testing.T) (string, string) {
		t.Helper()
		workDir := t.TempDir()
		runUUID := "02b5e8da-a37b-4666-9892-44706466438e"
		for _, dir := range []string{"source", "intermediate", "output"} {
			path := filepath.Join(workDir, runUUID, dir, "iconify", "icons")
			if err := os.MkdirAll(path, 0755); err != nil {
				t.Fatalf("failed to seed %s: %v", dir, err)
			}
			if err := os.WriteFile(filepath.Join(path, "coffee-cup.webp"), []byte("RIFFxxxxWEBP"), 0644); err != nil {
				t.Fatalf("failed to seed file in %s: %v", dir, err)
			}
		}
		return workDir, runUUID
	}

	workDir, runUUID := seedRun(t)
	ip := &ImageProcessor{UUID: runUUID, Config: &Config{WorkDir: workDir, AutoCleanup: true}}
	ip.Cleanup()

	for _, dir := range []string{"source", "intermediate"} {
		if _, err := os.Stat(filepath.Join(workDir, runUUID, dir)); !os.IsNotExist(err) {
			t.Errorf("%s dir still exists after Cleanup with AutoCleanup enabled", dir)
		}
	}
	if _, err := os.Stat(filepath.Join(workDir, runUUID, "output", "iconify", "icons", "coffee-cup.webp")); err != nil {
		t.Errorf("output was removed by Cleanup - results must be preserved: %v", err)
	}

	// Scenario: AutoCleanup disabled - nothing is removed.
	workDir, runUUID = seedRun(t)
	ip = &ImageProcessor{UUID: runUUID, Config: &Config{WorkDir: workDir, AutoCleanup: false}}
	ip.Cleanup()
	for _, dir := range []string{"source", "intermediate", "output"} {
		if _, err := os.Stat(filepath.Join(workDir, runUUID, dir)); err != nil {
			t.Errorf("%s dir missing after no-op Cleanup: %v", dir, err)
		}
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
