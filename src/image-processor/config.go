package imageprocessor

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Contributor             string         `yaml:"contributor"`
	SourceBucket            string         `yaml:"source_bucket"`
	TargetBucket            string         `yaml:"target_bucket"`
	Include                 []string       `yaml:"include_prefixes"`
	Exclude                 []string       `yaml:"omit_prefixes"`
	Region                  string         `yaml:"aws_region"`
	DryRun                  bool           `yaml:"dry_run"`
	FFmpegPath              string         `yaml:"ffmpegPath"`
	IsLocal                 bool           `yaml:"is_local"`
	UploadToS3              bool           `yaml:"upload_to_s3"`
	LocalSource             string         `yaml:"local_source"`
	LocalTarget             string         `yaml:"local_target"`
	AutoCleanup             bool           `yaml:"auto_cleanup"`
	WebpSizes               map[string]int `yaml:"webp_sizes"`
	WatermarkPath           string         `yaml:"watermark_path"`
	RoleArn                 string         `yaml:"role_arn"`
	LoggingOutput           int            `yaml:"logging_output"`
	Logfile                 string         `yaml:"logfile"`
	WorkDir                 string         `yaml:"work_dir"`
	OutputDir               string         `yaml:"output_dir"`
	UseHardwareAcceleration bool           `yaml:"use_hardware_acceleration"`
	WorkerPoolSize          int            `yaml:"worker_pool_size"`
	DownloadWorkerPoolSize  int            `yaml:"download_worker_pool_size"`
	ProcessWorkerPoolSize   int            `yaml:"process_worker_pool_size"`
}

func (config *Config) GetSourceDir() string {
	return filepath.Join(config.WorkDir, "source")
}

func (config *Config) GetIntermediateDir() string {
	return filepath.Join(config.WorkDir, "intermediate")
}

func (config *Config) GetTargetDir() string {
	return filepath.Join(config.WorkDir, "output")
}

func (config *Config) GetWorkDir() string {
	return config.WorkDir
}

func (config *Config) SetDefaults() {

	// Test if *config.LocalRun is set. If not set, set to false
	config.IsLocal = config.IsLocal || false

	config.UploadToS3 = config.UploadToS3 || false

	if config.WorkDir == "" {
		config.WorkDir = "./tmp/work"
	}

	if config.OutputDir == "" {
		config.OutputDir = "./tmp/output"
	}

	// UseHardwareAcceleration defaults to false
	config.UseHardwareAcceleration = config.UseHardwareAcceleration || false

	// WorkerPoolSize defaults to 1
	if config.WorkerPoolSize == 0 {
		config.WorkerPoolSize = 1
	}

	// DownloadWorkerPoolSize defaults to 1
	if config.DownloadWorkerPoolSize == 0 {
		config.DownloadWorkerPoolSize = 1
	}

	// Logfile defaults to './tmp/image-processor.log'
	if config.Logfile == "" {
		config.Logfile = "./tmp/image-processor.log"
	}

	// auto_cleanup defaults to false
	config.AutoCleanup = config.AutoCleanup || false

	// Region defaults to 'us-east-1'
	if config.Region == "" {
		config.Region = "us-east-1"
	}
}

func NewConfig(configpath string) (*Config, error) {
	file, err := os.Open(configpath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %s", err)
	}
	defer file.Close()

	config := &Config{}

	if err := yaml.NewDecoder(file).Decode(config); err != nil {
		return nil, fmt.Errorf("error decoding config file: %s", err)
	}

	config.SetDefaults()

	return config, nil
}
