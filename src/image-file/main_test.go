package imagefile

import (
	"os"
	"path/filepath"
	"testing"
)

// NOTE: These tests cover the Local Set path pattern, which is the shape
// produced by local-mode runs (run.sh) and S3 object keys. The Web Set,
// Web Family, and Local Family patterns are currently broken in
// populateImageFile (mis-indexed capture groups) and are intentionally not
// tested until that contract is fixed.

// TestNewImageFile_LocalSetPath verifies every parsed field for the canonical
// contributor/type/family/set/file path shape.
func TestNewImageFile_LocalSetPath(t *testing.T) {
	// Scenario: an icon SVG at the standard five-segment source layout.
	img, err := NewImageFile("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg")
	if err != nil {
		t.Fatalf("NewImageFile() error = %v", err)
	}
	if img == nil {
		t.Fatal("NewImageFile() = nil, want parsed ImageFile")
	}

	if !img.IsValid {
		t.Error("IsValid = false, want true")
	}
	if img.Contributor != "iconify" {
		t.Errorf("Contributor = %q, want %q", img.Contributor, "iconify")
	}
	if img.ProductType == nil || *img.ProductType != "icons" {
		t.Errorf("ProductType = %v, want icons", img.ProductType)
	}
	if img.FamilyUniqueID != "2C11DB2D5F79" {
		t.Errorf("FamilyUniqueID = %q, want %q", img.FamilyUniqueID, "2C11DB2D5F79")
	}
	if img.SetUniqueID == nil || *img.SetUniqueID != "B24091F3DF3E" {
		t.Errorf("SetUniqueID = %v, want B24091F3DF3E", img.SetUniqueID)
	}
	if img.Filename != "coffee-cup.svg" {
		t.Errorf("Filename = %q, want %q", img.Filename, "coffee-cup.svg")
	}
	if img.Extension != "svg" {
		t.Errorf("Extension = %q, want %q", img.Extension, "svg")
	}
	if img.Slug != "coffee-cup" {
		t.Errorf("Slug = %q, want %q", img.Slug, "coffee-cup")
	}
	if img.Stem != "coffee-cup" {
		t.Errorf("Stem = %q, want %q", img.Stem, "coffee-cup")
	}
	if img.ObjectKey != "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg" {
		t.Errorf("ObjectKey = %q, want source-relative key", img.ObjectKey)
	}
	if img.OptimizedImageKey != "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.webp" {
		t.Errorf("OptimizedImageKey = %q, want .webp key", img.OptimizedImageKey)
	}
	if img.FileType != "svg" {
		t.Errorf("FileType = %q, want %q", img.FileType, "svg")
	}
}

// TestNewImageFile_AbsolutePath verifies that the pattern matches anywhere in
// the path, producing a source-relative ObjectKey from an absolute input.
func TestNewImageFile_AbsolutePath(t *testing.T) {
	// Scenario: local-mode walk produces absolute paths under the source root.
	input := "/Users/converter/source/iconify/illustrations/58DC40590C5D/FFD4DC6639ED/mountain-sunrise.svg"
	img, err := NewImageFile(input)
	if err != nil {
		t.Fatalf("NewImageFile() error = %v", err)
	}
	if img == nil {
		t.Fatal("NewImageFile() = nil, want parsed ImageFile")
	}
	if img.ObjectKey != "iconify/illustrations/58DC40590C5D/FFD4DC6639ED/mountain-sunrise.svg" {
		t.Errorf("ObjectKey = %q, want source-relative key", img.ObjectKey)
	}
	if img.ProductType == nil || *img.ProductType != "illustrations" {
		t.Errorf("ProductType = %v, want illustrations", img.ProductType)
	}
}

// TestNewImageFile_StemStripsSizeSuffixes verifies the size-suffix stripping
// contract used to group size variants of the same image.
func TestNewImageFile_StemStripsSizeSuffixes(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantSlug string
		wantStem string
	}{
		// Scenario: trailing numeric suffix is treated as a size and stripped.
		{"numeric suffix", "robot-2.svg", "robot-2", "robot"},
		// Scenario: @2x retina suffix is stripped.
		{"retina suffix", "coffee-cup@2x.svg", "coffee-cup@2x", "coffee-cup"},
		// Scenario: multi-word slug with trailing numeric suffix.
		{"multi-word numeric", "man-face-avatar-head-2.svg", "man-face-avatar-head-2", "man-face-avatar-head"},
		// Scenario: hyphenated name without size suffix is unchanged.
		{"no size suffix", "scuba-diver-man.svg", "scuba-diver-man", "scuba-diver-man"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := NewImageFile("iconify/icons/2C11DB2D5F79/B24091F3DF3E/" + tt.filename)
			if err != nil {
				t.Fatalf("NewImageFile() error = %v", err)
			}
			if img == nil {
				t.Fatal("NewImageFile() = nil, want parsed ImageFile")
			}
			if img.Slug != tt.wantSlug {
				t.Errorf("Slug = %q, want %q", img.Slug, tt.wantSlug)
			}
			if img.Stem != tt.wantStem {
				t.Errorf("Stem = %q, want %q", img.Stem, tt.wantStem)
			}
		})
	}
}

// TestNewImageFile_SkipsHiddenAndNonMatching verifies that hidden files and
// paths outside the recognized layout are skipped (nil, nil) rather than
// treated as errors.
func TestNewImageFile_SkipsHiddenAndNonMatching(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Scenario: macOS metadata file must be skipped.
		{"hidden file", ".DS_Store"},
		// Scenario: repo file outside the source layout.
		{"non-matching path", "README.md"},
		// Scenario: unknown product type segment.
		{"unknown product type", "iconify/banners/2C11DB2D5F79/B24091F3DF3E/sale.svg"},
		// Scenario: family unique ID with wrong length (11 chars, not 12).
		{"short family id", "iconify/icons/2C11DB2D5F7/B24091F3DF3E/coffee-cup.svg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := NewImageFile(tt.input)
			if err != nil {
				t.Errorf("NewImageFile(%q) error = %v, want nil", tt.input, err)
			}
			if img != nil {
				t.Errorf("NewImageFile(%q) = %+v, want nil (skipped)", tt.input, img)
			}
		})
	}
}

// TestImageFile_Exists verifies existence checking against the real
// filesystem for both a present and an absent file.
func TestImageFile_Exists(t *testing.T) {
	// Scenario: the parsed source SVG is on disk under a temp source root.
	root := t.TempDir()
	relKey := filepath.Join("iconify", "icons", "2C11DB2D5F79", "B24091F3DF3E", "coffee-cup.svg")
	fullPath := filepath.Join(root, relKey)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create fixture dirs: %v", err)
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`)
	if err := os.WriteFile(fullPath, svg, 0644); err != nil {
		t.Fatalf("failed to seed fixture file: %v", err)
	}

	img, err := NewImageFile(fullPath)
	if err != nil || img == nil {
		t.Fatalf("NewImageFile() = %v, %v; want parsed ImageFile", img, err)
	}

	exists, err := img.Exists()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false for a file on disk, want true")
	}

	// Scenario: the same path after the file is removed.
	if err := os.Remove(fullPath); err != nil {
		t.Fatalf("failed to remove fixture: %v", err)
	}
	exists, err = img.Exists()
	if err != nil {
		t.Fatalf("Exists() after removal error = %v", err)
	}
	if exists {
		t.Error("Exists() = true for a removed file, want false")
	}
}
