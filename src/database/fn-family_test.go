package database

import (
	"reflect"
	"testing"

	"github.com/iconifyit/go-batch-svg-to-webp/src/models"

	"gorm.io/gorm"
)

func TestDatabaseService_GetFamilyById(t *testing.T) {
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
		want    *models.Family
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetFamilyById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetFamilyById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetFamilyById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetFamily(t *testing.T) {
	type fields struct {
		DB *gorm.DB
	}
	type args struct {
		params *QueryParams
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Family
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetFamily(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetFamily() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetFamily() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetFamilies(t *testing.T) {
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
		want    []models.Family
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetFamilies(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetFamilies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetFamilies() = %v, want %v", got, tt.want)
			}
		})
	}
}
