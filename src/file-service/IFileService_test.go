package fileservice

import (
	"reflect"
	"testing"
)

func TestNewFileService(t *testing.T) {
	type args struct {
		input ServiceInput
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
			if got := NewFileService(tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileService() = %v, want %v", got, tt.want)
			}
		})
	}
}
