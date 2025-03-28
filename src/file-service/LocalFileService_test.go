package fileservice

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

func TestNewLocalFileService(t *testing.T) {
	type args struct {
		config *ServiceInput
	}
	tests := []struct {
		name string
		args args
		want IFileService
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLocalFileService(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLocalFileService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalFileService_Transfer(t *testing.T) {
	type fields struct {
		UUID       string
		Session    *session.Session
		Config     FileServiceConfig
		SourceRoot string
		TargetRoot string
	}
	type args struct {
		input TransferInput
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
			svc := &LocalFileService{
				UUID:       tt.fields.UUID,
				Session:    tt.fields.Session,
				Config:     tt.fields.Config,
				SourceRoot: tt.fields.SourceRoot,
				TargetRoot: tt.fields.TargetRoot,
			}
			if err := svc.Transfer(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("LocalFileService.Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalFileService_ListFiles(t *testing.T) {
	type fields struct {
		UUID       string
		Session    *session.Session
		Config     FileServiceConfig
		SourceRoot string
		TargetRoot string
	}
	type args struct {
		input  ListFilesInput
		filter func(*string) bool
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
			svc := &LocalFileService{
				UUID:       tt.fields.UUID,
				Session:    tt.fields.Session,
				Config:     tt.fields.Config,
				SourceRoot: tt.fields.SourceRoot,
				TargetRoot: tt.fields.TargetRoot,
			}
			got, err := svc.ListFiles(tt.args.input, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("LocalFileService.ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LocalFileService.ListFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalFileService_ToImageFiles(t *testing.T) {
	type fields struct {
		UUID       string
		Session    *session.Session
		Config     FileServiceConfig
		SourceRoot string
		TargetRoot string
	}
	type args struct {
		files []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*imagefile.ImageFile
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &LocalFileService{
				UUID:       tt.fields.UUID,
				Session:    tt.fields.Session,
				Config:     tt.fields.Config,
				SourceRoot: tt.fields.SourceRoot,
				TargetRoot: tt.fields.TargetRoot,
			}
			got, err := svc.ToImageFiles(tt.args.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("LocalFileService.ToImageFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LocalFileService.ToImageFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalFileService_Download(t *testing.T) {
	type fields struct {
		UUID       string
		Session    *session.Session
		Config     FileServiceConfig
		SourceRoot string
		TargetRoot string
	}
	type args struct {
		file *imagefile.ImageFile
		dest string
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
			svc := &LocalFileService{
				UUID:       tt.fields.UUID,
				Session:    tt.fields.Session,
				Config:     tt.fields.Config,
				SourceRoot: tt.fields.SourceRoot,
				TargetRoot: tt.fields.TargetRoot,
			}
			got, err := svc.Download(tt.args.file, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("LocalFileService.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LocalFileService.Download() = %v, want %v", got, tt.want)
			}
		})
	}
}
