package imageprocessor

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewConfig_ParsesYAMLAndAppliesDefaults verifies that explicit YAML
// values are parsed and that unset values receive documented defaults.
func TestNewConfig_ParsesYAMLAndAppliesDefaults(t *testing.T) {
	// Scenario: a local-mode config with worker pools and sizes set, but
	// work_dir, logfile, and region left to default.
	yaml := `
is_local: true
upload_to_s3: false
local_source: ./test
local_target: ./test/output
ffmpegPath: /opt/homebrew/bin/ffmpeg
watermark_path: ./assets/watermark.svg
role_arn: arn:aws:iam::000000000000:role/svg-webp-app-role
webp_sizes:
  thumbnail: 128
  preview: 512
  watermark: 512
worker_pool_size: 10
download_worker_pool_size: 5
process_worker_pool_size: 10
logging_output: 3
`
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(yaml), 0644); err != nil {
		t.Fatalf("failed to write config fixture: %v", err)
	}

	config, err := NewConfig(path)
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	// Explicit values are preserved.
	if !config.IsLocal {
		t.Error("IsLocal = false, want true")
	}
	if config.LocalSource != "./test" {
		t.Errorf("LocalSource = %q, want %q", config.LocalSource, "./test")
	}
	if config.WorkerPoolSize != 10 {
		t.Errorf("WorkerPoolSize = %d, want 10", config.WorkerPoolSize)
	}
	if config.DownloadWorkerPoolSize != 5 {
		t.Errorf("DownloadWorkerPoolSize = %d, want 5", config.DownloadWorkerPoolSize)
	}
	if got := config.WebpSizes["thumbnail"]; got != 128 {
		t.Errorf("WebpSizes[thumbnail] = %d, want 128", got)
	}
	if got := config.WebpSizes["preview"]; got != 512 {
		t.Errorf("WebpSizes[preview] = %d, want 512", got)
	}

	// Unset values fall back to defaults.
	if config.WorkDir != "./tmp/work" {
		t.Errorf("WorkDir default = %q, want %q", config.WorkDir, "./tmp/work")
	}
	if config.OutputDir != "./tmp/output" {
		t.Errorf("OutputDir default = %q, want %q", config.OutputDir, "./tmp/output")
	}
	if config.Logfile != "./tmp/image-processor.log" {
		t.Errorf("Logfile default = %q, want %q", config.Logfile, "./tmp/image-processor.log")
	}
	if config.Region != "us-east-1" {
		t.Errorf("Region default = %q, want %q", config.Region, "us-east-1")
	}
}

// TestNewConfig_MissingFile verifies the error contract for a nonexistent
// config path.
func TestNewConfig_MissingFile(t *testing.T) {
	// Scenario: the operator passes a path that does not exist.
	if _, err := NewConfig(filepath.Join(t.TempDir(), "no-such-config.yml")); err == nil {
		t.Fatal("NewConfig() with missing file: expected error, got nil")
	}
}

// TestNewConfig_MalformedYAML verifies the error contract for invalid YAML.
func TestNewConfig_MalformedYAML(t *testing.T) {
	// Scenario: the config file contains non-YAML content.
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte("{{ not yaml : ["), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	if _, err := NewConfig(path); err == nil {
		t.Fatal("NewConfig() with malformed YAML: expected error, got nil")
	}
}

// TestSetDefaults_ZeroConfig verifies every default applied to an empty
// config.
func TestSetDefaults_ZeroConfig(t *testing.T) {
	// Scenario: a zero-value Config receives all documented defaults.
	config := &Config{}
	config.SetDefaults()

	if config.WorkDir != "./tmp/work" {
		t.Errorf("WorkDir = %q, want %q", config.WorkDir, "./tmp/work")
	}
	if config.OutputDir != "./tmp/output" {
		t.Errorf("OutputDir = %q, want %q", config.OutputDir, "./tmp/output")
	}
	if config.WorkerPoolSize != 1 {
		t.Errorf("WorkerPoolSize = %d, want 1", config.WorkerPoolSize)
	}
	if config.DownloadWorkerPoolSize != 1 {
		t.Errorf("DownloadWorkerPoolSize = %d, want 1", config.DownloadWorkerPoolSize)
	}
	if config.Logfile != "./tmp/image-processor.log" {
		t.Errorf("Logfile = %q, want %q", config.Logfile, "./tmp/image-processor.log")
	}
	if config.Region != "us-east-1" {
		t.Errorf("Region = %q, want %q", config.Region, "us-east-1")
	}
	if config.IsLocal {
		t.Error("IsLocal = true, want false")
	}
	if config.AutoCleanup {
		t.Error("AutoCleanup = true, want false")
	}
}

// TestSetDefaults_PreservesExplicitValues verifies defaults never overwrite
// explicitly configured values.
func TestSetDefaults_PreservesExplicitValues(t *testing.T) {
	// Scenario: an operator-tuned config keeps its values after SetDefaults.
	config := &Config{
		WorkDir:                "/Volumes/image-processor-ramdisk/work",
		WorkerPoolSize:         10,
		DownloadWorkerPoolSize: 5,
		Region:                 "eu-west-1",
		Logfile:                "./output.log",
	}
	config.SetDefaults()

	if config.WorkDir != "/Volumes/image-processor-ramdisk/work" {
		t.Errorf("WorkDir = %q, want RAM disk path preserved", config.WorkDir)
	}
	if config.WorkerPoolSize != 10 {
		t.Errorf("WorkerPoolSize = %d, want 10", config.WorkerPoolSize)
	}
	if config.DownloadWorkerPoolSize != 5 {
		t.Errorf("DownloadWorkerPoolSize = %d, want 5", config.DownloadWorkerPoolSize)
	}
	if config.Region != "eu-west-1" {
		t.Errorf("Region = %q, want %q", config.Region, "eu-west-1")
	}
	if config.Logfile != "./output.log" {
		t.Errorf("Logfile = %q, want %q", config.Logfile, "./output.log")
	}
}
