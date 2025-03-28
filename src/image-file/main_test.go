package imagefile

import (
	"reflect"
	"testing"
)

func TestNewImageFile(t *testing.T) {
	type args struct {
		inputPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *ImageFile
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewImageFile(tt.args.inputPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewImageFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewImageFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_populateImageFile(t *testing.T) {
	type args struct {
		image      *ImageFile
		matches    []string
		startIndex int
	}
	tests := []struct {
		name    string
		args    args
		want    *ImageFile
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := populateImageFile(tt.args.image, tt.args.matches, tt.args.startIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("populateImageFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("populateImageFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFile_getOptimizedImageKey(t *testing.T) {
	type fields struct {
		URL               string
		ObjectKey         string
		InputPath         string
		Contributor       string
		ProductType       *string
		FamilyUniqueID    string
		SetUniqueID       *string
		Filename          string
		Extension         string
		Slug              string
		Stem              string
		IsValid           bool
		Error             string
		OptimizedImageKey string
		FileType          string
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
			image := &ImageFile{
				URL:               tt.fields.URL,
				ObjectKey:         tt.fields.ObjectKey,
				InputPath:         tt.fields.InputPath,
				Contributor:       tt.fields.Contributor,
				ProductType:       tt.fields.ProductType,
				FamilyUniqueID:    tt.fields.FamilyUniqueID,
				SetUniqueID:       tt.fields.SetUniqueID,
				Filename:          tt.fields.Filename,
				Extension:         tt.fields.Extension,
				Slug:              tt.fields.Slug,
				Stem:              tt.fields.Stem,
				IsValid:           tt.fields.IsValid,
				Error:             tt.fields.Error,
				OptimizedImageKey: tt.fields.OptimizedImageKey,
				FileType:          tt.fields.FileType,
			}
			if got := image.getOptimizedImageKey(); got != tt.want {
				t.Errorf("ImageFile.getOptimizedImageKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFile_buildObjectKey(t *testing.T) {
	type fields struct {
		URL               string
		ObjectKey         string
		InputPath         string
		Contributor       string
		ProductType       *string
		FamilyUniqueID    string
		SetUniqueID       *string
		Filename          string
		Extension         string
		Slug              string
		Stem              string
		IsValid           bool
		Error             string
		OptimizedImageKey string
		FileType          string
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
			image := &ImageFile{
				URL:               tt.fields.URL,
				ObjectKey:         tt.fields.ObjectKey,
				InputPath:         tt.fields.InputPath,
				Contributor:       tt.fields.Contributor,
				ProductType:       tt.fields.ProductType,
				FamilyUniqueID:    tt.fields.FamilyUniqueID,
				SetUniqueID:       tt.fields.SetUniqueID,
				Filename:          tt.fields.Filename,
				Extension:         tt.fields.Extension,
				Slug:              tt.fields.Slug,
				Stem:              tt.fields.Stem,
				IsValid:           tt.fields.IsValid,
				Error:             tt.fields.Error,
				OptimizedImageKey: tt.fields.OptimizedImageKey,
				FileType:          tt.fields.FileType,
			}
			if got := image.buildObjectKey(); got != tt.want {
				t.Errorf("ImageFile.buildObjectKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFile_stripSizeFromSlug(t *testing.T) {
	type fields struct {
		URL               string
		ObjectKey         string
		InputPath         string
		Contributor       string
		ProductType       *string
		FamilyUniqueID    string
		SetUniqueID       *string
		Filename          string
		Extension         string
		Slug              string
		Stem              string
		IsValid           bool
		Error             string
		OptimizedImageKey string
		FileType          string
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
			image := &ImageFile{
				URL:               tt.fields.URL,
				ObjectKey:         tt.fields.ObjectKey,
				InputPath:         tt.fields.InputPath,
				Contributor:       tt.fields.Contributor,
				ProductType:       tt.fields.ProductType,
				FamilyUniqueID:    tt.fields.FamilyUniqueID,
				SetUniqueID:       tt.fields.SetUniqueID,
				Filename:          tt.fields.Filename,
				Extension:         tt.fields.Extension,
				Slug:              tt.fields.Slug,
				Stem:              tt.fields.Stem,
				IsValid:           tt.fields.IsValid,
				Error:             tt.fields.Error,
				OptimizedImageKey: tt.fields.OptimizedImageKey,
				FileType:          tt.fields.FileType,
			}
			if got := image.stripSizeFromSlug(); got != tt.want {
				t.Errorf("ImageFile.stripSizeFromSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFile_ToJSON(t *testing.T) {
	type fields struct {
		URL               string
		ObjectKey         string
		InputPath         string
		Contributor       string
		ProductType       *string
		FamilyUniqueID    string
		SetUniqueID       *string
		Filename          string
		Extension         string
		Slug              string
		Stem              string
		IsValid           bool
		Error             string
		OptimizedImageKey string
		FileType          string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image := &ImageFile{
				URL:               tt.fields.URL,
				ObjectKey:         tt.fields.ObjectKey,
				InputPath:         tt.fields.InputPath,
				Contributor:       tt.fields.Contributor,
				ProductType:       tt.fields.ProductType,
				FamilyUniqueID:    tt.fields.FamilyUniqueID,
				SetUniqueID:       tt.fields.SetUniqueID,
				Filename:          tt.fields.Filename,
				Extension:         tt.fields.Extension,
				Slug:              tt.fields.Slug,
				Stem:              tt.fields.Stem,
				IsValid:           tt.fields.IsValid,
				Error:             tt.fields.Error,
				OptimizedImageKey: tt.fields.OptimizedImageKey,
				FileType:          tt.fields.FileType,
			}
			if got := image.ToJSON(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageFile.ToJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFile_ToSnakeCase(t *testing.T) {
	type fields struct {
		URL               string
		ObjectKey         string
		InputPath         string
		Contributor       string
		ProductType       *string
		FamilyUniqueID    string
		SetUniqueID       *string
		Filename          string
		Extension         string
		Slug              string
		Stem              string
		IsValid           bool
		Error             string
		OptimizedImageKey string
		FileType          string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]interface{}
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image := &ImageFile{
				URL:               tt.fields.URL,
				ObjectKey:         tt.fields.ObjectKey,
				InputPath:         tt.fields.InputPath,
				Contributor:       tt.fields.Contributor,
				ProductType:       tt.fields.ProductType,
				FamilyUniqueID:    tt.fields.FamilyUniqueID,
				SetUniqueID:       tt.fields.SetUniqueID,
				Filename:          tt.fields.Filename,
				Extension:         tt.fields.Extension,
				Slug:              tt.fields.Slug,
				Stem:              tt.fields.Stem,
				IsValid:           tt.fields.IsValid,
				Error:             tt.fields.Error,
				OptimizedImageKey: tt.fields.OptimizedImageKey,
				FileType:          tt.fields.FileType,
			}
			if got := image.ToSnakeCase(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ImageFile.ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageFile_Exists(t *testing.T) {
	type fields struct {
		URL               string
		ObjectKey         string
		InputPath         string
		Contributor       string
		ProductType       *string
		FamilyUniqueID    string
		SetUniqueID       *string
		Filename          string
		Extension         string
		Slug              string
		Stem              string
		IsValid           bool
		Error             string
		OptimizedImageKey string
		FileType          string
	}
	tests := []struct {
		name    string
		fields  fields
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			image := &ImageFile{
				URL:               tt.fields.URL,
				ObjectKey:         tt.fields.ObjectKey,
				InputPath:         tt.fields.InputPath,
				Contributor:       tt.fields.Contributor,
				ProductType:       tt.fields.ProductType,
				FamilyUniqueID:    tt.fields.FamilyUniqueID,
				SetUniqueID:       tt.fields.SetUniqueID,
				Filename:          tt.fields.Filename,
				Extension:         tt.fields.Extension,
				Slug:              tt.fields.Slug,
				Stem:              tt.fields.Stem,
				IsValid:           tt.fields.IsValid,
				Error:             tt.fields.Error,
				OptimizedImageKey: tt.fields.OptimizedImageKey,
				FileType:          tt.fields.FileType,
			}
			got, err := image.Exists()
			if (err != nil) != tt.wantErr {
				t.Errorf("ImageFile.Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ImageFile.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}
