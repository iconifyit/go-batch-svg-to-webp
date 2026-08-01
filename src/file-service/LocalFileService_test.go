package fileservice

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// seedSourceTree creates a realistic local source layout under a temp root:
// two valid icon SVGs and one non-image file. It returns the root and the
// two valid relative keys.
func seedSourceTree(t *testing.T) (string, []string) {
	t.Helper()
	root := t.TempDir()

	validKeys := []string{
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg",
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg",
	}
	svg := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`
	for _, key := range validKeys {
		full := filepath.Join(root, key)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("failed to create fixture dirs: %v", err)
		}
		if err := os.WriteFile(full, []byte(svg), 0644); err != nil {
			t.Fatalf("failed to seed %s: %v", key, err)
		}
	}
	// A file that is not an image source (must be filtered by image parsing).
	if err := os.WriteFile(filepath.Join(root, "iconify", "manifest.json"), []byte(`{"batch":"2024-01-15"}`), 0644); err != nil {
		t.Fatalf("failed to seed manifest: %v", err)
	}
	return root, validKeys
}

// TestLocalFileService_Transfer verifies a local copy preserves content.
func TestLocalFileService_Transfer(t *testing.T) {
	// Scenario: a finished WebP is transferred from work dir to target dir.
	dir := t.TempDir()
	src := filepath.Join(dir, "coffee-cup-thumbnail.webp")
	dst := filepath.Join(dir, "coffee-cup-thumbnail-final.webp")
	content := []byte("RIFFxxxxWEBPVP8 ")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("failed to seed source: %v", err)
	}

	svc := &LocalFileService{}
	if err := svc.Transfer(TransferInput{SourceFilePath: src, TargetFilePath: dst}); err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("transferred file missing: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("transferred content = %q, want %q", got, content)
	}
}

// TestLocalFileService_Transfer_MissingSource verifies the error contract
// and that the error names both paths.
func TestLocalFileService_Transfer_MissingSource(t *testing.T) {
	// Scenario: the source WebP was never produced.
	dir := t.TempDir()
	src := filepath.Join(dir, "missing.webp")
	svc := &LocalFileService{}
	err := svc.Transfer(TransferInput{SourceFilePath: src, TargetFilePath: filepath.Join(dir, "out.webp")})
	if err == nil {
		t.Fatal("Transfer() with missing source: expected error, got nil")
	}
	if !strings.Contains(err.Error(), src) {
		t.Errorf("error %q does not name the source path", err)
	}
}

// TestLocalFileService_ListFiles_NoFilter verifies that a nil filter lists
// every file in the tree, image or not.
func TestLocalFileService_ListFiles_NoFilter(t *testing.T) {
	// Scenario: walk the source tree with no filter - all 3 files returned.
	root, validKeys := seedSourceTree(t)

	svc := &LocalFileService{}
	files, err := svc.ListFiles(ListFilesInput{SourceRoot: root}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(root, validKeys[0]),
		filepath.Join(root, validKeys[1]),
		filepath.Join(root, "iconify", "manifest.json"),
	}
	sort.Strings(files)
	sort.Strings(want)
	if len(files) != len(want) {
		t.Fatalf("ListFiles() returned %d files %v, want %d", len(files), files, len(want))
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("ListFiles()[%d] = %q, want %q", i, files[i], want[i])
		}
	}
}

// TestLocalFileService_ListFiles_WithFilter verifies that with a filter the
// listing keeps only paths that both pass the filter and parse as images.
func TestLocalFileService_ListFiles_WithFilter(t *testing.T) {
	// Scenario: an accept-all filter still drops manifest.json because it is
	// not a parseable image path.
	root, validKeys := seedSourceTree(t)

	svc := &LocalFileService{}
	acceptAll := func(path *string) bool { return true }
	files, err := svc.ListFiles(ListFilesInput{SourceRoot: root}, acceptAll)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	want := []string{
		filepath.Join(root, validKeys[0]),
		filepath.Join(root, validKeys[1]),
	}
	sort.Strings(files)
	sort.Strings(want)
	if len(files) != len(want) {
		t.Fatalf("ListFiles() = %v, want the 2 valid image paths", files)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("ListFiles()[%d] = %q, want %q", i, files[i], want[i])
		}
	}

	// Scenario: a reject-all filter yields no files.
	rejectAll := func(path *string) bool { return false }
	files, err = svc.ListFiles(ListFilesInput{SourceRoot: root}, rejectAll)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Errorf("ListFiles() with reject-all filter = %v, want empty", files)
	}
}

// TestLocalFileService_ToImageFiles verifies conversion keeps valid image
// paths and silently drops non-image paths.
func TestLocalFileService_ToImageFiles(t *testing.T) {
	// Scenario: two valid icon paths mixed with a manifest file.
	svc := &LocalFileService{}
	imageFiles, err := svc.ToImageFiles([]string{
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg",
		"iconify/manifest.json",
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg",
	})
	if err != nil {
		t.Fatalf("ToImageFiles() error = %v", err)
	}
	if len(imageFiles) != 2 {
		t.Fatalf("ToImageFiles() returned %d files, want 2", len(imageFiles))
	}
	if imageFiles[0].ObjectKey != "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg" {
		t.Errorf("first ObjectKey = %q", imageFiles[0].ObjectKey)
	}
	if imageFiles[1].Stem != "robot" {
		t.Errorf("second Stem = %q, want %q (size suffix stripped)", imageFiles[1].Stem, "robot")
	}
}

// TestLocalFileService_Download verifies the source file is copied to the
// destination path, creating intermediate directories.
func TestLocalFileService_Download(t *testing.T) {
	// Scenario: "download" (local copy) of a source SVG into the work dir.
	root, validKeys := seedSourceTree(t)
	workDir := t.TempDir()

	img, err := (&LocalFileService{}).ToImageFiles([]string{validKeys[0]})
	if err != nil || len(img) != 1 {
		t.Fatalf("fixture parse failed: %v", err)
	}

	svc := &LocalFileService{SourceRoot: root}
	dest := filepath.Join(workDir, "source", img[0].ObjectKey)
	got, err := svc.Download(img[0], dest)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if got != dest {
		t.Errorf("Download() returned %q, want %q", got, dest)
	}

	srcContent, err := os.ReadFile(filepath.Join(root, validKeys[0]))
	if err != nil {
		t.Fatalf("failed to read source: %v", err)
	}
	dstContent, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("downloaded file missing: %v", err)
	}
	if string(dstContent) != string(srcContent) {
		t.Errorf("downloaded content differs from source")
	}
}

// TestLocalFileService_Download_UnconfiguredSourceRoot verifies the error
// contract when the service has no source root wired, instead of silently
// reading relative to the process working directory.
func TestLocalFileService_Download_UnconfiguredSourceRoot(t *testing.T) {
	// Scenario: a service constructed without SourceRoot.
	img, err := (&LocalFileService{}).ToImageFiles([]string{
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg",
	})
	if err != nil || len(img) != 1 {
		t.Fatalf("fixture parse failed: %v", err)
	}
	svc := &LocalFileService{}
	if _, err := svc.Download(img[0], filepath.Join(t.TempDir(), "out.svg")); err == nil {
		t.Fatal("Download() without SourceRoot: expected error, got nil")
	}
}

// TestLocalFileService_Download_MissingSource verifies the error contract
// when the ObjectKey does not exist under SourceRoot.
func TestLocalFileService_Download_MissingSource(t *testing.T) {
	// Scenario: the listed file was deleted between listing and download.
	img, err := (&LocalFileService{}).ToImageFiles([]string{
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg",
	})
	if err != nil || len(img) != 1 {
		t.Fatalf("fixture parse failed: %v", err)
	}

	svc := &LocalFileService{SourceRoot: t.TempDir()}
	if _, err := svc.Download(img[0], filepath.Join(t.TempDir(), "out.svg")); err == nil {
		t.Fatal("Download() with missing source: expected error, got nil")
	}
}
