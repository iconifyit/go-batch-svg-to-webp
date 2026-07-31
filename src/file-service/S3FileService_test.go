package fileservice

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"
)

// mockS3Client is a test double for the S3 API. It records the inputs each
// method receives and returns configurable results, so tests can verify the
// service sends the right requests without any network access.
type mockS3Client struct {
	s3iface.S3API

	putInputs []*s3.PutObjectInput
	putBodies []string
	putErr    error

	getInput  *s3.GetObjectInput
	getOutput *s3.GetObjectOutput
	getErr    error

	headInput *s3.HeadObjectInput
	headErr   error

	listInput *s3.ListObjectsV2Input
	listPages []*s3.ListObjectsV2Output
}

func (m *mockS3Client) PutObject(in *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	m.putInputs = append(m.putInputs, in)
	if in.Body != nil {
		content, err := io.ReadAll(in.Body)
		if err != nil {
			return nil, err
		}
		m.putBodies = append(m.putBodies, string(content))
	}
	if m.putErr != nil {
		return nil, m.putErr
	}
	return &s3.PutObjectOutput{}, nil
}

func (m *mockS3Client) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	m.getInput = in
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getOutput, nil
}

func (m *mockS3Client) HeadObject(in *s3.HeadObjectInput) (*s3.HeadObjectOutput, error) {
	m.headInput = in
	if m.headErr != nil {
		return nil, m.headErr
	}
	return &s3.HeadObjectOutput{}, nil
}

func (m *mockS3Client) ListObjectsV2Pages(in *s3.ListObjectsV2Input, fn func(*s3.ListObjectsV2Output, bool) bool) error {
	m.listInput = in
	for i, page := range m.listPages {
		if !fn(page, i == len(m.listPages)-1) {
			break
		}
	}
	return nil
}

// seedLocalFile writes a small WebP-like fixture and returns its path.
func seedLocalFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to seed fixture file: %v", err)
	}
	return path
}

// TestS3Transfer_DefaultsToServiceBucket verifies Transfer falls back to the
// factory-wired target bucket and uploads the file bytes under the target
// key when the caller does not name a bucket.
func TestS3Transfer_DefaultsToServiceBucket(t *testing.T) {
	// Scenario: a finished WebP is transferred without naming a bucket, on a
	// service wired like NewFileService builds it (TargetBucket only).
	mock := &mockS3Client{}
	svc := &S3FileService{TargetBucket: "vectoricons-webp-staging", Client: mock}
	src := seedLocalFile(t, "coffee-cup-preview.webp", "RIFFxxxxWEBPVP8 ")

	err := svc.Transfer(TransferInput{
		SourceFilePath: src,
		TargetFilePath: "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-preview.webp",
	})
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}

	if len(mock.putInputs) != 1 {
		t.Fatalf("PutObject called %d times, want 1", len(mock.putInputs))
	}
	put := mock.putInputs[0]
	if aws.StringValue(put.Bucket) != "vectoricons-webp-staging" {
		t.Errorf("bucket = %q, want the factory-wired target bucket", aws.StringValue(put.Bucket))
	}
	if aws.StringValue(put.Key) != "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-preview.webp" {
		t.Errorf("key = %q, want target file path", aws.StringValue(put.Key))
	}
	if mock.putBodies[0] != "RIFFxxxxWEBPVP8 " {
		t.Errorf("uploaded body = %q, want the file content", mock.putBodies[0])
	}
	if got := aws.StringValue(put.ContentType); got != "image/webp" {
		t.Errorf("content type = %q, want image/webp", got)
	}
}

// TestS3Transfer_ExplicitBucketWins verifies an explicit bucket overrides
// the service default.
func TestS3Transfer_ExplicitBucketWins(t *testing.T) {
	// Scenario: caller targets a specific bucket for one transfer.
	mock := &mockS3Client{}
	svc := &S3FileService{TargetBucket: "vectoricons-webp-staging", Client: mock}
	src := seedLocalFile(t, "robot-thumbnail.webp", "RIFFxxxxWEBP")

	err := svc.Transfer(TransferInput{
		Bucket:         "vectoricons-public",
		SourceFilePath: src,
		TargetFilePath: "iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-thumbnail.webp",
	})
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if got := aws.StringValue(mock.putInputs[0].Bucket); got != "vectoricons-public" {
		t.Errorf("bucket = %q, want explicit bucket", got)
	}
}

// TestS3Transfer_MissingSource verifies the error contract when the local
// file does not exist, and that no upload is attempted.
func TestS3Transfer_MissingSource(t *testing.T) {
	// Scenario: the WebP to upload was never produced.
	mock := &mockS3Client{}
	svc := &S3FileService{BucketName: "vectoricons-webp-staging", Client: mock}

	err := svc.Transfer(TransferInput{
		SourceFilePath: filepath.Join(t.TempDir(), "missing.webp"),
		TargetFilePath: "iconify/icons/2C11DB2D5F79/B24091F3DF3E/missing.webp",
	})
	if err == nil {
		t.Fatal("Transfer() with missing source: expected error, got nil")
	}
	if len(mock.putInputs) != 0 {
		t.Errorf("PutObject was called despite missing source file")
	}
}

// TestS3NilSessionFailsFast verifies that a service with neither an
// injected Client nor an AWS Session returns a clear error instead of a nil
// dereference.
func TestS3NilSessionFailsFast(t *testing.T) {
	// Scenario: a service constructed with buckets but no session or client.
	svc := &S3FileService{SourceBucket: "vectoricons-private", TargetBucket: "vectoricons-webp-staging"}
	src := seedLocalFile(t, "coffee-cup-preview.webp", "RIFFxxxxWEBP")

	if err := svc.Upload(src, "coffee-cup-preview.webp"); err == nil {
		t.Error("Upload() without session or client: expected error, got nil")
	}
	if _, err := svc.ListFiles(ListFilesInput{SourceRoot: "vectoricons-private"}, nil); err == nil {
		t.Error("ListFiles() without session or client: expected error, got nil")
	}
	if _, err := svc.Exists("coffee-cup-preview.webp"); err == nil {
		t.Error("Exists() without session or client: expected error, got nil")
	}
}

// TestNewFileService_S3PassesSession verifies the factory wires the AWS
// session into the S3 implementation.
func TestNewFileService_S3PassesSession(t *testing.T) {
	// Scenario: S3-mode construction with a session, as NewImageProcessor
	// wires it.
	sess, err := session.NewSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	svc := NewFileService(ServiceInput{
		IsLocal:    false,
		Session:    sess,
		SourceRoot: "vectoricons-private",
		TargetRoot: "vectoricons-webp-staging",
	})
	s3svc, ok := svc.(*S3FileService)
	if !ok {
		t.Fatalf("NewFileService() = %T, want *S3FileService", svc)
	}
	if s3svc.Session != sess {
		t.Error("factory did not pass the AWS session through to S3FileService")
	}
}

// TestS3UnconfiguredBucketFailsFast verifies Transfer, Upload, and Exists
// return a clear configuration error instead of sending S3 requests with an
// empty bucket name when neither BucketName nor TargetBucket is set.
func TestS3UnconfiguredBucketFailsFast(t *testing.T) {
	// Scenario: a service constructed without any bucket wiring.
	mock := &mockS3Client{}
	svc := &S3FileService{Client: mock}
	src := seedLocalFile(t, "coffee-cup-preview.webp", "RIFFxxxxWEBP")

	if err := svc.Transfer(TransferInput{SourceFilePath: src, TargetFilePath: "coffee-cup-preview.webp"}); err == nil {
		t.Error("Transfer() without a bucket: expected error, got nil")
	}
	if err := svc.Upload(src, "coffee-cup-preview.webp"); err == nil {
		t.Error("Upload() without a bucket: expected error, got nil")
	}
	if _, err := svc.Exists("coffee-cup-preview.webp"); err == nil {
		t.Error("Exists() without a bucket: expected error, got nil")
	}
	if len(mock.putInputs) != 0 || mock.headInput != nil {
		t.Error("S3 API was called despite missing bucket configuration")
	}
}

// TestS3TargetBucket_ExplicitNameOverridesTarget verifies a manually set
// BucketName takes precedence over the factory-wired TargetBucket.
func TestS3TargetBucket_ExplicitNameOverridesTarget(t *testing.T) {
	// Scenario: an operator overrides the bucket for a one-off upload.
	mock := &mockS3Client{}
	svc := &S3FileService{
		BucketName:   "vectoricons-public",
		TargetBucket: "vectoricons-webp-staging",
		Client:       mock,
	}
	src := seedLocalFile(t, "coffee-cup-preview.webp", "RIFFxxxxWEBP")

	if err := svc.Upload(src, "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-preview.webp"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if got := aws.StringValue(mock.putInputs[0].Bucket); got != "vectoricons-public" {
		t.Errorf("bucket = %q, want explicit BucketName to win", got)
	}
}

// TestS3Upload verifies Upload sends the file bytes to the factory-wired
// target bucket under the given key.
func TestS3Upload(t *testing.T) {
	// Scenario: upload a watermarked WebP to the staging bucket.
	mock := &mockS3Client{}
	svc := &S3FileService{TargetBucket: "vectoricons-webp-staging", Client: mock}
	src := seedLocalFile(t, "coffee-cup-watermark.webp", "RIFFyyyyWEBPVP8 ")

	if err := svc.Upload(src, "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-watermark.webp"); err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	if len(mock.putInputs) != 1 {
		t.Fatalf("PutObject called %d times, want 1", len(mock.putInputs))
	}
	if got := aws.StringValue(mock.putInputs[0].Key); got != "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-watermark.webp" {
		t.Errorf("key = %q, want object key", got)
	}
	if mock.putBodies[0] != "RIFFyyyyWEBPVP8 " {
		t.Errorf("uploaded body = %q, want the file content", mock.putBodies[0])
	}
	if got := aws.StringValue(mock.putInputs[0].ContentType); got != "image/webp" {
		t.Errorf("content type = %q, want image/webp", got)
	}
}

// TestContentTypeForFile verifies the extension-to-MIME mapping, including
// the unknown-extension fallback and case insensitivity.
func TestContentTypeForFile(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		// Scenario: the formats this pipeline produces and consumes.
		{"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-preview.webp", "image/webp"},
		{"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg", "image/svg+xml"},
		{"work/intermediate/coffee-cup-thumbnail.png", "image/png"},
		{"uploads/photo.jpg", "image/jpeg"},
		{"uploads/photo.jpeg", "image/jpeg"},
		// Scenario: uppercase extension still maps.
		{"uploads/COFFEE-CUP.WEBP", "image/webp"},
		// Scenario: unknown extension falls back to a generic binary type.
		{"iconify/manifest.json", "application/octet-stream"},
		{"no-extension", "application/octet-stream"},
	}
	for _, tt := range tests {
		if got := contentTypeForFile(tt.path); got != tt.want {
			t.Errorf("contentTypeForFile(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// TestS3Upload_PutError verifies an S3 failure surfaces as a wrapped error.
func TestS3Upload_PutError(t *testing.T) {
	// Scenario: S3 rejects the upload (e.g. access denied).
	mock := &mockS3Client{putErr: errors.New("AccessDenied")}
	svc := &S3FileService{TargetBucket: "vectoricons-webp-staging", Client: mock}
	src := seedLocalFile(t, "robot-preview.webp", "RIFFzzzzWEBP")

	err := svc.Upload(src, "iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-preview.webp")
	if err == nil {
		t.Fatal("Upload() with failing PutObject: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "AccessDenied") {
		t.Errorf("error %q does not surface the S3 failure", err)
	}
}

// TestS3Download verifies Download requests the right object and writes its
// body to the dest path, creating parent directories - the same contract as
// LocalFileService.Download.
func TestS3Download(t *testing.T) {
	// Scenario: download a source SVG from the private bucket into a nested
	// work-dir path that does not exist yet.
	objectKey := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg"
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`

	mock := &mockS3Client{
		getOutput: &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(svgContent))},
	}
	svc := &S3FileService{SourceBucket: "vectoricons-private", Client: mock}

	img, err := imagefile.NewImageFile(objectKey)
	if err != nil || img == nil {
		t.Fatalf("fixture parse failed: %v", err)
	}

	dest := filepath.Join(t.TempDir(), "work", "source", objectKey)
	got, err := svc.Download(img, dest)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	if got != dest {
		t.Errorf("Download() returned %q, want dest %q", got, dest)
	}

	if aws.StringValue(mock.getInput.Bucket) != "vectoricons-private" {
		t.Errorf("GetObject bucket = %q, want source bucket", aws.StringValue(mock.getInput.Bucket))
	}
	if aws.StringValue(mock.getInput.Key) != objectKey {
		t.Errorf("GetObject key = %q, want object key", aws.StringValue(mock.getInput.Key))
	}
	written, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("downloaded file missing at dest: %v", err)
	}
	if string(written) != svgContent {
		t.Errorf("downloaded content differs from the S3 object body")
	}
}

// TestS3Download_GetError verifies an S3 failure surfaces as a wrapped
// error.
func TestS3Download_GetError(t *testing.T) {
	// Scenario: the object was deleted between listing and download.
	mock := &mockS3Client{getErr: errors.New("NoSuchKey")}
	svc := &S3FileService{SourceBucket: "vectoricons-private", Client: mock}

	img, err := imagefile.NewImageFile("iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg")
	if err != nil || img == nil {
		t.Fatalf("fixture parse failed: %v", err)
	}

	if _, err := svc.Download(img, ""); err == nil {
		t.Fatal("Download() with failing GetObject: expected error, got nil")
	}
}

// TestS3Exists verifies the contract: true for a present object, (false,
// nil) only for a genuine not-found response, and an error for operational
// failures so they are not mistaken for absence.
func TestS3Exists(t *testing.T) {
	// Scenario: check for an already-converted WebP before reprocessing.
	objectKey := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-preview.webp"

	mock := &mockS3Client{}
	svc := &S3FileService{TargetBucket: "vectoricons-public", Client: mock}
	exists, err := svc.Exists(objectKey)
	if err != nil || !exists {
		t.Errorf("Exists() = %v, %v for a present object; want true, nil", exists, err)
	}
	if got := aws.StringValue(mock.headInput.Bucket); got != "vectoricons-public" {
		t.Errorf("HeadObject bucket = %q, want the factory-wired target bucket", got)
	}

	// Scenario: the object genuinely does not exist (S3 NotFound).
	notFound := awserr.New("NotFound", "Not Found", nil)
	svc = &S3FileService{TargetBucket: "vectoricons-public", Client: &mockS3Client{headErr: notFound}}
	exists, err = svc.Exists(objectKey)
	if err != nil || exists {
		t.Errorf("Exists() = %v, %v for a missing object; want false, nil", exists, err)
	}

	// Scenario: an operational failure (access denied) must surface as an
	// error, not report the object as absent.
	denied := awserr.New("AccessDenied", "Access Denied", nil)
	svc = &S3FileService{TargetBucket: "vectoricons-public", Client: &mockS3Client{headErr: denied}}
	if _, err = svc.Exists(objectKey); err == nil {
		t.Error("Exists() with AccessDenied: expected error, got nil")
	}
}

// TestS3ListFiles_UnconfiguredBucketFailsFast verifies listing without any
// bucket configuration returns a clear error instead of an invalid request.
func TestS3ListFiles_UnconfiguredBucketFailsFast(t *testing.T) {
	// Scenario: a service constructed without any bucket wiring.
	svc := &S3FileService{Client: &mockS3Client{}}
	if _, err := svc.ListFiles(ListFilesInput{}, nil); err == nil {
		t.Error("ListFiles() without a bucket: expected error, got nil")
	}
}

// TestS3ListFiles verifies pagination is walked, the filter is applied, and
// the listing targets the service bucket.
func TestS3ListFiles(t *testing.T) {
	// Scenario: two pages of objects; only SVG source files pass the filter.
	mock := &mockS3Client{
		listPages: []*s3.ListObjectsV2Output{
			{Contents: []*s3.Object{
				{Key: aws.String("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg")},
				{Key: aws.String("iconify/manifest.json")},
			}},
			{Contents: []*s3.Object{
				{Key: aws.String("iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg")},
			}},
		},
	}
	svc := &S3FileService{SourceBucket: "vectoricons-private", Client: mock}

	svgOnly := func(key *string) bool { return strings.HasSuffix(*key, ".svg") }
	files, err := svc.ListFiles(ListFilesInput{SourceRoot: "vectoricons-private"}, svgOnly)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if aws.StringValue(mock.listInput.Bucket) != "vectoricons-private" {
		t.Errorf("list bucket = %q, want the requested source bucket", aws.StringValue(mock.listInput.Bucket))
	}
	want := []string{
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg",
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg",
	}
	if len(files) != len(want) {
		t.Fatalf("ListFiles() = %v, want %v", files, want)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("ListFiles()[%d] = %q, want %q", i, files[i], want[i])
		}
	}
}

// TestS3ListFiles_NilFilter verifies a nil filter accepts every object,
// matching the LocalFileService contract instead of panicking, and that an
// empty input falls back to the service's configured source bucket.
func TestS3ListFiles_NilFilter(t *testing.T) {
	// Scenario: a caller lists the bucket without any filtering and without
	// naming a bucket in the input.
	mock := &mockS3Client{
		listPages: []*s3.ListObjectsV2Output{
			{Contents: []*s3.Object{
				{Key: aws.String("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg")},
				{Key: aws.String("iconify/manifest.json")},
			}},
		},
	}
	svc := &S3FileService{SourceBucket: "vectoricons-private", Client: mock}

	files, err := svc.ListFiles(ListFilesInput{}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("ListFiles() with nil filter = %v, want all 2 objects", files)
	}
	if aws.StringValue(mock.listInput.Bucket) != "vectoricons-private" {
		t.Errorf("list bucket = %q, want fallback to service source bucket", aws.StringValue(mock.listInput.Bucket))
	}
}

// TestS3ToImageFiles verifies conversion keeps valid image paths and drops
// non-image paths, mirroring the local implementation.
func TestS3ToImageFiles(t *testing.T) {
	// Scenario: a listing mixing two icons and a manifest file.
	svc := &S3FileService{}
	imageFiles, err := svc.ToImageFiles([]string{
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg",
		"iconify/manifest.json",
		"iconify/icons/2C11DB2D5F79/B24091F3DF3E/robot-2.svg",
	})
	if err != nil {
		t.Fatalf("ToImageFiles() error = %v", err)
	}
	if len(imageFiles) != 2 {
		t.Fatalf("ToImageFiles() returned %d files, want 2", len(imageFiles))
	}
	if imageFiles[0].ObjectKey != "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg" {
		t.Errorf("first ObjectKey = %q", imageFiles[0].ObjectKey)
	}
}
