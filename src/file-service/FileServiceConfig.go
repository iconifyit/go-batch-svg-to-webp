package fileservice

type FileServiceConfig struct {
	LoggingOutput int    `yaml:"logging_output"`
	Logfile       string `yaml:"logfile"`
	WorkDir       string `yaml:"work_dir"`
	OutputDir     string `yaml:"output_dir"`
	LocalSource   string `yaml:"local_source"`
	LocalTarget   string `yaml:"local_target"`
	SourceBucket  string `yaml:"source_bucket"`
	TargetBucket  string `yaml:"target_bucket"`
	IsLocal       bool   `json:"is_local"`
}
