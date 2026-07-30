package fileservice

import (
	"testing"
)

// TestNewFileService_Local verifies the factory returns a LocalFileService
// wired with the local roots when IsLocal is set.
func TestNewFileService_Local(t *testing.T) {
	// Scenario: local-mode config selects the filesystem implementation.
	svc := NewFileService(ServiceInput{
		IsLocal:    true,
		SourceRoot: "./test",
		TargetRoot: "./test/output",
	})

	local, ok := svc.(*LocalFileService)
	if !ok {
		t.Fatalf("NewFileService(IsLocal: true) = %T, want *LocalFileService", svc)
	}
	if local.SourceRoot != "./test" {
		t.Errorf("SourceRoot = %q, want %q", local.SourceRoot, "./test")
	}
	if local.TargetRoot != "./test/output" {
		t.Errorf("TargetRoot = %q, want %q", local.TargetRoot, "./test/output")
	}
}

// TestNewFileService_S3 verifies the factory returns an S3FileService wired
// with the bucket names when IsLocal is not set.
func TestNewFileService_S3(t *testing.T) {
	// Scenario: S3-mode config selects the S3 implementation, mapping
	// SourceRoot/TargetRoot to bucket names.
	svc := NewFileService(ServiceInput{
		IsLocal:    false,
		SourceRoot: "vectoricons-svg-source",
		TargetRoot: "vectoricons-webp-staging",
	})

	s3svc, ok := svc.(*S3FileService)
	if !ok {
		t.Fatalf("NewFileService(IsLocal: false) = %T, want *S3FileService", svc)
	}
	if s3svc.SourceBucket != "vectoricons-svg-source" {
		t.Errorf("SourceBucket = %q, want %q", s3svc.SourceBucket, "vectoricons-svg-source")
	}
	if s3svc.TargetBucket != "vectoricons-webp-staging" {
		t.Errorf("TargetBucket = %q, want %q", s3svc.TargetBucket, "vectoricons-webp-staging")
	}
}
