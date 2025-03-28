package database

import (
	"reflect"
	"testing"

	"gorm.io/gorm"
)

func TestNewDatabaseService(t *testing.T) {
	tests := []struct {
		name    string
		want    *DatabaseService
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDatabaseService()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatabaseService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatabaseService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatabaseService_Close(t *testing.T) {
	type fields struct {
		DB *gorm.DB
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
			svc := &DatabaseService{
				DB: tt.fields.DB,
			}
			if err := svc.Close(); (err != nil) != tt.wantErr {
				t.Errorf("DatabaseService.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWhereNot(t *testing.T) {
	type args struct {
		column string
		value  interface{}
	}
	tests := []struct {
		name string
		args args
		want func(tx *gorm.DB) *gorm.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WhereNot(tt.args.column, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WhereNot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhereIn(t *testing.T) {
	type args struct {
		column string
		values []interface{}
	}
	tests := []struct {
		name string
		args args
		want func(tx *gorm.DB) *gorm.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WhereIn(tt.args.column, tt.args.values); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WhereIn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhereLike(t *testing.T) {
	type args struct {
		column  string
		pattern string
	}
	tests := []struct {
		name string
		args args
		want func(tx *gorm.DB) *gorm.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WhereLike(tt.args.column, tt.args.pattern); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WhereLike() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhereCustom(t *testing.T) {
	type args struct {
		customCondition string
		args            []interface{}
	}
	tests := []struct {
		name string
		args args
		want func(tx *gorm.DB) *gorm.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WhereCustom(tt.args.customCondition, tt.args.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WhereCustom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWhere(t *testing.T) {
	type args struct {
		column string
		value  interface{}
	}
	tests := []struct {
		name string
		args args
		want func(tx *gorm.DB) *gorm.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Where(tt.args.column, tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Where() = %v, want %v", got, tt.want)
			}
		})
	}
}
