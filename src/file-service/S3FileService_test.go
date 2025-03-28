package fileservice

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

func TestNewS3FileService(t *testing.T) {
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
			if got := NewS3FileService(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewS3FileService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3FileService_Transfer(t *testing.T) {
	type fields struct {
		UUID         string
		Session      *session.Session
		Config       FileServiceConfig
		BucketName   string
		ObjectKey    string
		SourceBucket string
		TargetBucket string
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
			svc := &S3FileService{
				UUID:         tt.fields.UUID,
				Session:      tt.fields.Session,
				Config:       tt.fields.Config,
				BucketName:   tt.fields.BucketName,
				ObjectKey:    tt.fields.ObjectKey,
				SourceBucket: tt.fields.SourceBucket,
				TargetBucket: tt.fields.TargetBucket,
			}
			if err := svc.Transfer(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("S3FileService.Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestS3FileService_ListFiles(t *testing.T) {
	type fields struct {
		UUID         string
		Session      *session.Session
		Config       FileServiceConfig
		BucketName   string
		ObjectKey    string
		SourceBucket string
		TargetBucket string
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
			svc := &S3FileService{
				UUID:         tt.fields.UUID,
				Session:      tt.fields.Session,
				Config:       tt.fields.Config,
				BucketName:   tt.fields.BucketName,
				ObjectKey:    tt.fields.ObjectKey,
				SourceBucket: tt.fields.SourceBucket,
				TargetBucket: tt.fields.TargetBucket,
			}
			got, err := svc.ListFiles(tt.args.input, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3FileService.ListFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("S3FileService.ListFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3FileService_ToImageFiles(t *testing.T) {
	type fields struct {
		UUID         string
		Session      *session.Session
		Config       FileServiceConfig
		BucketName   string
		ObjectKey    string
		SourceBucket string
		TargetBucket string
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
			svc := &S3FileService{
				UUID:         tt.fields.UUID,
				Session:      tt.fields.Session,
				Config:       tt.fields.Config,
				BucketName:   tt.fields.BucketName,
				ObjectKey:    tt.fields.ObjectKey,
				SourceBucket: tt.fields.SourceBucket,
				TargetBucket: tt.fields.TargetBucket,
			}
			got, err := svc.ToImageFiles(tt.args.files)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3FileService.ToImageFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("S3FileService.ToImageFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3FileService_Download(t *testing.T) {
	type fields struct {
		UUID         string
		Session      *session.Session
		Config       FileServiceConfig
		BucketName   string
		ObjectKey    string
		SourceBucket string
		TargetBucket string
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
			svc := &S3FileService{
				UUID:         tt.fields.UUID,
				Session:      tt.fields.Session,
				Config:       tt.fields.Config,
				BucketName:   tt.fields.BucketName,
				ObjectKey:    tt.fields.ObjectKey,
				SourceBucket: tt.fields.SourceBucket,
				TargetBucket: tt.fields.TargetBucket,
			}
			got, err := svc.Download(tt.args.file, tt.args.dest)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3FileService.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("S3FileService.Download() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3FileService_Exists(t *testing.T) {
	type fields struct {
		UUID         string
		Session      *session.Session
		Config       FileServiceConfig
		BucketName   string
		ObjectKey    string
		SourceBucket string
		TargetBucket string
	}
	type args struct {
		objectKey string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &S3FileService{
				UUID:         tt.fields.UUID,
				Session:      tt.fields.Session,
				Config:       tt.fields.Config,
				BucketName:   tt.fields.BucketName,
				ObjectKey:    tt.fields.ObjectKey,
				SourceBucket: tt.fields.SourceBucket,
				TargetBucket: tt.fields.TargetBucket,
			}
			got, err := svc.Exists(tt.args.objectKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("S3FileService.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("S3FileService.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestS3FileService_Upload(t *testing.T) {
	type fields struct {
		UUID         string
		Session      *session.Session
		Config       FileServiceConfig
		BucketName   string
		ObjectKey    string
		SourceBucket string
		TargetBucket string
	}
	type args struct {
		localPath string
		objectKey string
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
			svc := &S3FileService{
				UUID:         tt.fields.UUID,
				Session:      tt.fields.Session,
				Config:       tt.fields.Config,
				BucketName:   tt.fields.BucketName,
				ObjectKey:    tt.fields.ObjectKey,
				SourceBucket: tt.fields.SourceBucket,
				TargetBucket: tt.fields.TargetBucket,
			}
			if err := svc.Upload(tt.args.localPath, tt.args.objectKey); (err != nil) != tt.wantErr {
				t.Errorf("S3FileService.Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
