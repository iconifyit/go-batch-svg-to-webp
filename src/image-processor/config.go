// package imageprocessor

// import (
// 	"os"

// 	"gopkg.in/yaml.v2"
// )

// type Config struct {
// 	SourceBucket            string         `yaml:"source_bucket"`
// 	TargetBucket            string         `yaml:"target_bucket"`
// 	Include                 []string       `yaml:"include_prefixes"`
// 	Exclude                 []string       `yaml:"omit_prefixes"`
// 	Region                  string         `yaml:"aws_region"`
// 	DryRun                  bool           `yaml:"dry_run"`
// 	FFmpegPath              string         `yaml:"ffmpegPath"`
// 	IsLocal                 bool           `yaml:"is_local"`
// 	LocalSource             string         `yaml:"local_source"`
// 	LocalTarget             string         `yaml:"local_target"`
// 	AutoCleanup             bool           `yaml:"auto_cleanup"`
// 	WebpSizes               map[string]int `yaml:"webp_sizes"`
// 	WatermarkPath           string         `yaml:"watermark_path"`
// 	RoleArn                 string         `yaml:"role_arn"`
// 	LoggingOutput           int            `yaml:"logging_output"`
// 	Logfile                 string         `yaml:"logfile"`
// 	WorkDir                 string         `yaml:"work_dir"`
// 	OutputDir               string         `yaml:"output_dir"`
// 	UseHardwareAcceleration bool           `yaml:"use_hardware_acceleration"`
// 	WorkerPoolSize          int            `yaml:"worker_pool_size"`
// 	DownloadWorkerPoolSize  int            `yaml:"download_worker_pool_size"`
// 	ProcessWorkerPoolSize   int            `yaml:"process_worker_pool_size"`
// }

// // LoadConfig reads the configuration from the provided path.
// func LoadConfig(path string) (*Config, error) {
// 	file, err := os.Open(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	config := &Config{}
// 	if err := yaml.NewDecoder(file).Decode(config); err != nil {
// 		return nil, err
// 	}

// 	return config, nil
// }
