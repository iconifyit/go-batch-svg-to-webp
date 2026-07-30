package imageprocessor

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

// stringPtr is a test helper for building *string file paths.
func stringPtr(s string) *string {
	return &s
}

// TestShouldInclude verifies the include/exclude prefix filtering contract.
func TestShouldInclude(t *testing.T) {
	tests := []struct {
		name     string
		include  []string
		exclude  []string
		filePath *string
		want     bool
	}{
		// Scenario: nil path can never be included.
		{"nil path", nil, nil, nil, false},
		// Scenario: hidden files are always excluded.
		{"hidden file", nil, nil, stringPtr("iconify/icons/2C11DB2D5F79/B24091F3DF3E/.DS_Store"), false},
		// Scenario: no include/exclude rules includes every visible file.
		{"no rules includes all", nil, nil, stringPtr("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg"), true},
		// Scenario: file under an excluded prefix is dropped.
		{"excluded prefix", nil, []string{"iconify/illustrations"}, stringPtr("iconify/illustrations/58DC40590C5D/FFD4DC6639ED/mountain.svg"), false},
		// Scenario: file matching an include prefix is kept.
		{"include match", []string{"iconify/icons"}, nil, stringPtr("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg"), true},
		// Scenario: file outside all include prefixes is dropped.
		{"include miss", []string{"iconify/icons"}, nil, stringPtr("vectopus/icons/AA11BB22CC33/DD44EE55FF66/rocket.svg"), false},
		// Scenario: exclusion wins over inclusion for the same file.
		{"exclude beats include", []string{"iconify"}, []string{"iconify/icons"}, stringPtr("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{Config: &Config{Include: tt.include, Exclude: tt.exclude}}
			if got := ip.ShouldInclude(tt.filePath); got != tt.want {
				t.Errorf("ShouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsDir verifies directory detection for a dir, a file, and a missing
// path.
func TestIsDir(t *testing.T) {
	// Scenario: a real directory, a real file, and a nonexistent path.
	dir := t.TempDir()
	file := filepath.Join(dir, "coffee-cup.svg")
	if err := os.WriteFile(file, []byte("<svg/>"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	if got, err := IsDir(dir); err != nil || !got {
		t.Errorf("IsDir(dir) = %v, %v; want true, nil", got, err)
	}
	if got, err := IsDir(file); err != nil || got {
		t.Errorf("IsDir(file) = %v, %v; want false, nil", got, err)
	}
	if got, err := IsDir(filepath.Join(dir, "missing")); err != nil || got {
		t.Errorf("IsDir(missing) = %v, %v; want false, nil", got, err)
	}
}

// TestIsFile verifies file detection for a file, a dir, and a missing path.
func TestIsFile(t *testing.T) {
	// Scenario: a real file, a real directory, and a nonexistent path.
	dir := t.TempDir()
	file := filepath.Join(dir, "coffee-cup.svg")
	if err := os.WriteFile(file, []byte("<svg/>"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	if got, err := IsFile(file); err != nil || !got {
		t.Errorf("IsFile(file) = %v, %v; want true, nil", got, err)
	}
	if got, err := IsFile(dir); err != nil || got {
		t.Errorf("IsFile(dir) = %v, %v; want false, nil", got, err)
	}
	if got, err := IsFile(filepath.Join(dir, "missing")); err != nil || got {
		t.Errorf("IsFile(missing) = %v, %v; want false, nil", got, err)
	}
}

// TestListDirs verifies that only first-level directories are returned.
func TestListDirs(t *testing.T) {
	// Scenario: a source root containing two contributor dirs and one file.
	root := t.TempDir()
	for _, name := range []string{"iconify", "vectopus"} {
		if err := os.Mkdir(filepath.Join(root, name), 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	ip := &ImageProcessor{Config: &Config{}}
	dirs, err := ip.ListDirs(root)
	if err != nil {
		t.Fatalf("ListDirs() error = %v", err)
	}
	if len(dirs) != 2 {
		t.Fatalf("ListDirs() = %v, want 2 directories", dirs)
	}
	want := map[string]bool{"iconify": true, "vectopus": true}
	for _, d := range dirs {
		if !want[d] {
			t.Errorf("ListDirs() returned unexpected entry %q", d)
		}
	}
}

// TestListDirs_MissingRoot verifies the error contract for a nonexistent
// root directory.
func TestListDirs_MissingRoot(t *testing.T) {
	// Scenario: the configured source root does not exist.
	ip := &ImageProcessor{Config: &Config{}}
	if _, err := ip.ListDirs(filepath.Join(t.TempDir(), "missing-root")); err == nil {
		t.Fatal("ListDirs() with missing root: expected error, got nil")
	}
}

// TestSetupLogging_FileOutput verifies logging_output=2 creates and writes
// the configured log file.
func TestSetupLogging_FileOutput(t *testing.T) {
	// Scenario: file-only logging writes messages to the configured logfile.
	defer resetLogging()

	logfile := filepath.Join(t.TempDir(), "image-processor.log")
	ip := &ImageProcessor{Config: &Config{LoggingOutput: 2, Logfile: logfile}}

	if err := ip.SetupLogging(); err != nil {
		t.Fatalf("SetupLogging() error = %v", err)
	}
	log.Printf("processed iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg")

	content, err := os.ReadFile(logfile)
	if err != nil {
		t.Fatalf("log file not created: %v", err)
	}
	if len(content) == 0 {
		t.Error("log file is empty, want logged message")
	}
}

// TestSetupLogging_InvalidSetting verifies the error contract for an
// out-of-range logging_output value.
func TestSetupLogging_InvalidSetting(t *testing.T) {
	// Scenario: logging_output=7 is not a documented mode.
	defer resetLogging()

	ip := &ImageProcessor{Config: &Config{LoggingOutput: 7, Logfile: filepath.Join(t.TempDir(), "x.log")}}
	if err := ip.SetupLogging(); err == nil {
		t.Fatal("SetupLogging() with invalid setting: expected error, got nil")
	}
}

// resetLogging restores the global logger state mutated by SetupLogging.
func resetLogging() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags)
}

// TestIsLocalRun verifies the local-mode flag passthrough.
func TestIsLocalRun(t *testing.T) {
	// Scenario: local-mode and S3-mode configs.
	if !(&ImageProcessor{Config: &Config{IsLocal: true}}).IsLocalRun() {
		t.Error("IsLocalRun() = false for is_local: true config")
	}
	if (&ImageProcessor{Config: &Config{IsLocal: false}}).IsLocalRun() {
		t.Error("IsLocalRun() = true for is_local: false config")
	}
}
