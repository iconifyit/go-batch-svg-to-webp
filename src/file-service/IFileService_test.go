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

// TestNewS3FileService_PopulatesBuckets verifies the exported constructor
// wires the roots into bucket names, matching the factory contract.
func TestNewS3FileService_PopulatesBuckets(t *testing.T) {
	// Scenario: direct construction via the exported constructor.
	svc := NewS3FileService(&ServiceInput{
		UUID:       "02b5e8da-a37b-4666-9892-44706466438e",
		SourceRoot: "vectoricons-private",
		TargetRoot: "vectoricons-webp-staging",
	})
	s3svc, ok := svc.(*S3FileService)
	if !ok {
		t.Fatalf("NewS3FileService() = %T, want *S3FileService", svc)
	}
	if s3svc.SourceBucket != "vectoricons-private" {
		t.Errorf("SourceBucket = %q, want %q", s3svc.SourceBucket, "vectoricons-private")
	}
	if s3svc.TargetBucket != "vectoricons-webp-staging" {
		t.Errorf("TargetBucket = %q, want %q", s3svc.TargetBucket, "vectoricons-webp-staging")
	}
}

// TestNewLocalFileService_PopulatesRoots verifies the exported constructor
// wires the local roots, matching the factory contract.
func TestNewLocalFileService_PopulatesRoots(t *testing.T) {
	// Scenario: direct construction via the exported constructor.
	svc := NewLocalFileService(&ServiceInput{
		UUID:       "02b5e8da-a37b-4666-9892-44706466438e",
		SourceRoot: "./test/input",
		TargetRoot: "./test/output",
	})
	local, ok := svc.(*LocalFileService)
	if !ok {
		t.Fatalf("NewLocalFileService() = %T, want *LocalFileService", svc)
	}
	if local.SourceRoot != "./test/input" {
		t.Errorf("SourceRoot = %q, want %q", local.SourceRoot, "./test/input")
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
