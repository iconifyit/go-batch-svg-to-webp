package database

import (
	"reflect"
	"testing"

	"github.com/iconifyit/go-batch-svg-to-webp/src/models"

	"gorm.io/gorm"
)

func TestDatabaseService_GetIllustrationById(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Illustration
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetIllustrationById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetIllustrationById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetIllustrationById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetIllustration(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		params QueryParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Illustration
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetIllustration(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetIllustration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetIllustration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetIllustrations(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		params QueryParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Illustration
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetIllustrations(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetIllustrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetIllustrations() = %v, want %v", got, tt.want)
			}
		})
	}
}
