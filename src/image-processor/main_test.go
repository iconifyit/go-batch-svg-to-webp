package imageprocessor

import (
	"reflect"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	fileservice "github.com/iconifyit/go-batch-svg-to-webp/src/file-service"
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

func TestIsDir(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsDir(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsFile(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewImageProcessor(t *testing.T) {
	type args struct {
		contributor string
		configFile  string
	}
	tests := []struct {
		name string
		args args
		want *ImageProcessor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewImageProcessor(tt.args.contributor, tt.args.configFile); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewImageProcessor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSessionWithRole(t *testing.T) {
	type args struct {
		roleArn string
		region  string
	}
	tests := []struct {
		name    string
		args    args
		want    *session.Session
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SessionWithRole(tt.args.roleArn, tt.args.region)
			if (err != nil) != tt.wantErr {
				t.Errorf("SessionWithRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SessionWithRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	type args struct {
		path string
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
			got, err := LoadConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_IsLocalRun(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if got := ip.IsLocalRun(); got != tt.want {
				t.Errorf("ImageProcessor.IsLocalRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_SetupLogging(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if err := ip.SetupLogging(); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.SetupLogging() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_ShouldInclude(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	type args struct {
		filePath *string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if got := ip.ShouldInclude(tt.args.filePath); got != tt.want {
				t.Errorf("ImageProcessor.ShouldInclude() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_ListFiles(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			got, err := ip.ListFiles()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageProcessor.ListFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_ListDirs(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	type args struct {
		rootdir string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			got, err := ip.ListDirs(tt.args.rootdir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.ListDirs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageProcessor.ListDirs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_IsValidContributor(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	type args struct {
		username string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if got := ip.IsValidContributor(tt.args.username); got != tt.want {
				t.Errorf("ImageProcessor.IsValidContributor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_ListFilesForContributor(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	type args struct {
		contributor string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			got, err := ip.ListFilesForContributor(tt.args.contributor)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.ListFilesForContributor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageProcessor.ListFilesForContributor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_worker(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	type args struct {
		id        int
		jobs      <-chan imagefile.ImageFile
		wg        *sync.WaitGroup
		errorChan chan<- error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			ip.worker(tt.args.id, tt.args.jobs, tt.args.wg, tt.args.errorChan)
		})
	}
}

func TestImageProcessor_ImageFiles(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	type args struct {
		files []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []imagefile.ImageFile
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if got := ip.ImageFiles(tt.args.files); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageProcessor.ImageFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_ProcessFiles(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if err := ip.ProcessFiles(); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.ProcessFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_Run(t *testing.T) {
	type fields struct {
		UUID          string
		Contributor   string
		Config        *Config
		Session       *session.Session
		DownloadQueue chan imagefile.ImageFile
		ProcessQueue  chan imagefile.ImageFile
		FileService   fileservice.IFileService
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageProcessor{
				UUID:          tt.fields.UUID,
				Contributor:   tt.fields.Contributor,
				Config:        tt.fields.Config,
				Session:       tt.fields.Session,
				DownloadQueue: tt.fields.DownloadQueue,
				ProcessQueue:  tt.fields.ProcessQueue,
				FileService:   tt.fields.FileService,
			}
			if err := ip.Run(); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
