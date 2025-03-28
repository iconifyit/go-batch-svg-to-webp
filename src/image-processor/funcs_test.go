package imageprocessor

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	fileservice "github.com/iconifyit/go-batch-svg-to-webp/src/file-service"
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

func TestImageProcessor_ConvertSVGToPNG(t *testing.T) {
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
		svgPath string
		pngPath string
		size    int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := ip.ConvertSVGToPNG(tt.args.svgPath, tt.args.pngPath, tt.args.size); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.ConvertSVGToPNG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_RunFFmpeg(t *testing.T) {
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
		input  string
		output string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := ip.RunFFmpeg(tt.args.input, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.RunFFmpeg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_Watermark(t *testing.T) {
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
		inputFilePath  string
		outputFilePath string
		size           int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := ip.Watermark(tt.args.inputFilePath, tt.args.outputFilePath, tt.args.size); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.Watermark() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_LoadWatermark(t *testing.T) {
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
		want    string
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
			got, err := ip.LoadWatermark()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.LoadWatermark() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ImageProcessor.LoadWatermark() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_ProcessFile(t *testing.T) {
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
		imgFile imagefile.ImageFile
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := ip.ProcessFile(tt.args.imgFile); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.ProcessFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_CopyFile(t *testing.T) {
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
		src string
		dst string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := ip.CopyFile(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.CopyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_RenameFolderWithTimestamp(t *testing.T) {
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
		folderPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
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
			if err := ip.RenameFolderWithTimestamp(tt.args.folderPath); (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.RenameFolderWithTimestamp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestImageProcessor_downloadWorker(t *testing.T) {
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
			ip.downloadWorker(tt.args.id, tt.args.errorChan)
		})
	}
}

func TestImageProcessor_downloadFile(t *testing.T) {
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
		file *imagefile.ImageFile
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
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
			got, err := ip.downloadFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageProcessor.downloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ImageProcessor.downloadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageProcessor_processWorker(t *testing.T) {
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
			ip.processWorker(tt.args.id, tt.args.errorChan)
		})
	}
}

func TestImageProcessor_Cleanup(t *testing.T) {
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
			ip.Cleanup()
		})
	}
}
