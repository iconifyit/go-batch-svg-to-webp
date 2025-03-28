package database

import (
	"reflect"
	"testing"

	"github.com/iconifyit/go-batch-svg-to-webp/src/models"

	"gorm.io/gorm"
)

func TestDatabaseService_GetIconById(t *testing.T) {
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
		want    *models.Icon
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetIconById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetIconById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetIconById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetIcon(t *testing.T) {
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
		want    *models.Icon
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetIcon(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetIcon() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetIcon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetIcons(t *testing.T) {
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
		want    []models.Icon
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetIcons(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetIcons() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetIcons() = %v, want %v", got, tt.want)
			}
		})
	}
}
