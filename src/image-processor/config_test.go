package imageprocessor

import (
	"reflect"
	"testing"
)

func TestConfig_GetSourceDir(t *testing.T) {
	type fields struct {
		Contributor             string
		SourceBucket            string
		TargetBucket            string
		Include                 []string
		Exclude                 []string
		Region                  string
		DryRun                  bool
		FFmpegPath              string
		IsLocal                 bool
		UploadToS3              bool
		LocalSource             string
		LocalTarget             string
		AutoCleanup             bool
		WebpSizes               map[string]int
		WatermarkPath           string
		RoleArn                 string
		LoggingOutput           int
		Logfile                 string
		WorkDir                 string
		OutputDir               string
		UseHardwareAcceleration bool
		WorkerPoolSize          int
		DownloadWorkerPoolSize  int
		ProcessWorkerPoolSize   int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Contributor:             tt.fields.Contributor,
				SourceBucket:            tt.fields.SourceBucket,
				TargetBucket:            tt.fields.TargetBucket,
				Include:                 tt.fields.Include,
				Exclude:                 tt.fields.Exclude,
				Region:                  tt.fields.Region,
				DryRun:                  tt.fields.DryRun,
				FFmpegPath:              tt.fields.FFmpegPath,
				IsLocal:                 tt.fields.IsLocal,
				UploadToS3:              tt.fields.UploadToS3,
				LocalSource:             tt.fields.LocalSource,
				LocalTarget:             tt.fields.LocalTarget,
				AutoCleanup:             tt.fields.AutoCleanup,
				WebpSizes:               tt.fields.WebpSizes,
				WatermarkPath:           tt.fields.WatermarkPath,
				RoleArn:                 tt.fields.RoleArn,
				LoggingOutput:           tt.fields.LoggingOutput,
				Logfile:                 tt.fields.Logfile,
				WorkDir:                 tt.fields.WorkDir,
				OutputDir:               tt.fields.OutputDir,
				UseHardwareAcceleration: tt.fields.UseHardwareAcceleration,
				WorkerPoolSize:          tt.fields.WorkerPoolSize,
				DownloadWorkerPoolSize:  tt.fields.DownloadWorkerPoolSize,
				ProcessWorkerPoolSize:   tt.fields.ProcessWorkerPoolSize,
			}
			if got := config.GetSourceDir(); got != tt.want {
				t.Errorf("Config.GetSourceDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetIntermediateDir(t *testing.T) {
	type fields struct {
		Contributor             string
		SourceBucket            string
		TargetBucket            string
		Include                 []string
		Exclude                 []string
		Region                  string
		DryRun                  bool
		FFmpegPath              string
		IsLocal                 bool
		UploadToS3              bool
		LocalSource             string
		LocalTarget             string
		AutoCleanup             bool
		WebpSizes               map[string]int
		WatermarkPath           string
		RoleArn                 string
		LoggingOutput           int
		Logfile                 string
		WorkDir                 string
		OutputDir               string
		UseHardwareAcceleration bool
		WorkerPoolSize          int
		DownloadWorkerPoolSize  int
		ProcessWorkerPoolSize   int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Contributor:             tt.fields.Contributor,
				SourceBucket:            tt.fields.SourceBucket,
				TargetBucket:            tt.fields.TargetBucket,
				Include:                 tt.fields.Include,
				Exclude:                 tt.fields.Exclude,
				Region:                  tt.fields.Region,
				DryRun:                  tt.fields.DryRun,
				FFmpegPath:              tt.fields.FFmpegPath,
				IsLocal:                 tt.fields.IsLocal,
				UploadToS3:              tt.fields.UploadToS3,
				LocalSource:             tt.fields.LocalSource,
				LocalTarget:             tt.fields.LocalTarget,
				AutoCleanup:             tt.fields.AutoCleanup,
				WebpSizes:               tt.fields.WebpSizes,
				WatermarkPath:           tt.fields.WatermarkPath,
				RoleArn:                 tt.fields.RoleArn,
				LoggingOutput:           tt.fields.LoggingOutput,
				Logfile:                 tt.fields.Logfile,
				WorkDir:                 tt.fields.WorkDir,
				OutputDir:               tt.fields.OutputDir,
				UseHardwareAcceleration: tt.fields.UseHardwareAcceleration,
				WorkerPoolSize:          tt.fields.WorkerPoolSize,
				DownloadWorkerPoolSize:  tt.fields.DownloadWorkerPoolSize,
				ProcessWorkerPoolSize:   tt.fields.ProcessWorkerPoolSize,
			}
			if got := config.GetIntermediateDir(); got != tt.want {
				t.Errorf("Config.GetIntermediateDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetTargetDir(t *testing.T) {
	type fields struct {
		Contributor             string
		SourceBucket            string
		TargetBucket            string
		Include                 []string
		Exclude                 []string
		Region                  string
		DryRun                  bool
		FFmpegPath              string
		IsLocal                 bool
		UploadToS3              bool
		LocalSource             string
		LocalTarget             string
		AutoCleanup             bool
		WebpSizes               map[string]int
		WatermarkPath           string
		RoleArn                 string
		LoggingOutput           int
		Logfile                 string
		WorkDir                 string
		OutputDir               string
		UseHardwareAcceleration bool
		WorkerPoolSize          int
		DownloadWorkerPoolSize  int
		ProcessWorkerPoolSize   int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Contributor:             tt.fields.Contributor,
				SourceBucket:            tt.fields.SourceBucket,
				TargetBucket:            tt.fields.TargetBucket,
				Include:                 tt.fields.Include,
				Exclude:                 tt.fields.Exclude,
				Region:                  tt.fields.Region,
				DryRun:                  tt.fields.DryRun,
				FFmpegPath:              tt.fields.FFmpegPath,
				IsLocal:                 tt.fields.IsLocal,
				UploadToS3:              tt.fields.UploadToS3,
				LocalSource:             tt.fields.LocalSource,
				LocalTarget:             tt.fields.LocalTarget,
				AutoCleanup:             tt.fields.AutoCleanup,
				WebpSizes:               tt.fields.WebpSizes,
				WatermarkPath:           tt.fields.WatermarkPath,
				RoleArn:                 tt.fields.RoleArn,
				LoggingOutput:           tt.fields.LoggingOutput,
				Logfile:                 tt.fields.Logfile,
				WorkDir:                 tt.fields.WorkDir,
				OutputDir:               tt.fields.OutputDir,
				UseHardwareAcceleration: tt.fields.UseHardwareAcceleration,
				WorkerPoolSize:          tt.fields.WorkerPoolSize,
				DownloadWorkerPoolSize:  tt.fields.DownloadWorkerPoolSize,
				ProcessWorkerPoolSize:   tt.fields.ProcessWorkerPoolSize,
			}
			if got := config.GetTargetDir(); got != tt.want {
				t.Errorf("Config.GetTargetDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GetWorkDir(t *testing.T) {
	type fields struct {
		Contributor             string
		SourceBucket            string
		TargetBucket            string
		Include                 []string
		Exclude                 []string
		Region                  string
		DryRun                  bool
		FFmpegPath              string
		IsLocal                 bool
		UploadToS3              bool
		LocalSource             string
		LocalTarget             string
		AutoCleanup             bool
		WebpSizes               map[string]int
		WatermarkPath           string
		RoleArn                 string
		LoggingOutput           int
		Logfile                 string
		WorkDir                 string
		OutputDir               string
		UseHardwareAcceleration bool
		WorkerPoolSize          int
		DownloadWorkerPoolSize  int
		ProcessWorkerPoolSize   int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Contributor:             tt.fields.Contributor,
				SourceBucket:            tt.fields.SourceBucket,
				TargetBucket:            tt.fields.TargetBucket,
				Include:                 tt.fields.Include,
				Exclude:                 tt.fields.Exclude,
				Region:                  tt.fields.Region,
				DryRun:                  tt.fields.DryRun,
				FFmpegPath:              tt.fields.FFmpegPath,
				IsLocal:                 tt.fields.IsLocal,
				UploadToS3:              tt.fields.UploadToS3,
				LocalSource:             tt.fields.LocalSource,
				LocalTarget:             tt.fields.LocalTarget,
				AutoCleanup:             tt.fields.AutoCleanup,
				WebpSizes:               tt.fields.WebpSizes,
				WatermarkPath:           tt.fields.WatermarkPath,
				RoleArn:                 tt.fields.RoleArn,
				LoggingOutput:           tt.fields.LoggingOutput,
				Logfile:                 tt.fields.Logfile,
				WorkDir:                 tt.fields.WorkDir,
				OutputDir:               tt.fields.OutputDir,
				UseHardwareAcceleration: tt.fields.UseHardwareAcceleration,
				WorkerPoolSize:          tt.fields.WorkerPoolSize,
				DownloadWorkerPoolSize:  tt.fields.DownloadWorkerPoolSize,
				ProcessWorkerPoolSize:   tt.fields.ProcessWorkerPoolSize,
			}
			if got := config.GetWorkDir(); got != tt.want {
				t.Errorf("Config.GetWorkDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_SetDefaults(t *testing.T) {
	type fields struct {
		Contributor             string
		SourceBucket            string
		TargetBucket            string
		Include                 []string
		Exclude                 []string
		Region                  string
		DryRun                  bool
		FFmpegPath              string
		IsLocal                 bool
		UploadToS3              bool
		LocalSource             string
		LocalTarget             string
		AutoCleanup             bool
		WebpSizes               map[string]int
		WatermarkPath           string
		RoleArn                 string
		LoggingOutput           int
		Logfile                 string
		WorkDir                 string
		OutputDir               string
		UseHardwareAcceleration bool
		WorkerPoolSize          int
		DownloadWorkerPoolSize  int
		ProcessWorkerPoolSize   int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Contributor:             tt.fields.Contributor,
				SourceBucket:            tt.fields.SourceBucket,
				TargetBucket:            tt.fields.TargetBucket,
				Include:                 tt.fields.Include,
				Exclude:                 tt.fields.Exclude,
				Region:                  tt.fields.Region,
				DryRun:                  tt.fields.DryRun,
				FFmpegPath:              tt.fields.FFmpegPath,
				IsLocal:                 tt.fields.IsLocal,
				UploadToS3:              tt.fields.UploadToS3,
				LocalSource:             tt.fields.LocalSource,
				LocalTarget:             tt.fields.LocalTarget,
				AutoCleanup:             tt.fields.AutoCleanup,
				WebpSizes:               tt.fields.WebpSizes,
				WatermarkPath:           tt.fields.WatermarkPath,
				RoleArn:                 tt.fields.RoleArn,
				LoggingOutput:           tt.fields.LoggingOutput,
				Logfile:                 tt.fields.Logfile,
				WorkDir:                 tt.fields.WorkDir,
				OutputDir:               tt.fields.OutputDir,
				UseHardwareAcceleration: tt.fields.UseHardwareAcceleration,
				WorkerPoolSize:          tt.fields.WorkerPoolSize,
				DownloadWorkerPoolSize:  tt.fields.DownloadWorkerPoolSize,
				ProcessWorkerPoolSize:   tt.fields.ProcessWorkerPoolSize,
			}
			config.SetDefaults()
		})
	}
}

func TestNewConfig(t *testing.T) {
	type args struct {
		configpath string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.configpath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
