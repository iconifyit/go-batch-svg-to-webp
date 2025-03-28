package database

import (
	"reflect"
	"testing"

	"github.com/iconifyit/go-batch-svg-to-webp/src/models"

	"gorm.io/gorm"
)

func TestDatabaseService_GetSetById(t *testing.T) {
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
		want    *models.Set
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetSetById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetSetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetSetById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetSet(t *testing.T) {
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
		want    *models.Set
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetSet(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetSets(t *testing.T) {
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
		want    []models.Set
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetSets(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetSets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetSets() = %v, want %v", got, tt.want)
			}
		})
	}
}
