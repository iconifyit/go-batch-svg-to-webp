package database

import (
	"reflect"
	"testing"

	"github.com/iconifyit/go-batch-svg-to-webp/src/models"

	"gorm.io/gorm"
)

func TestDatabaseService_GetUserById(t *testing.T) {
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
		want    *models.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetUserById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetUserById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetUserById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetUser(t *testing.T) {
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
		want    *models.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetUser(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_GetUsers(t *testing.T) {
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
		want    []models.User
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			got, err := svc.GetUsers(tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.GetUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DatabaseService.GetUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}
