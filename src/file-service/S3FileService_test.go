package fileservice

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
// service bucket and uploads the file bytes under the target key.
func TestS3Transfer_DefaultsToServiceBucket(t *testing.T) {
	// Scenario: a finished WebP is transferred without naming a bucket.
	mock := &mockS3Client{}
	svc := &S3FileService{BucketName: "vectoricons-webp-staging", Client: mock}
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
		t.Errorf("bucket = %q, want service bucket", aws.StringValue(put.Bucket))
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
	svc := &S3FileService{BucketName: "vectoricons-webp-staging", Client: mock}
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

// TestS3Upload verifies Upload sends the file bytes to the service bucket
// under the given key.
func TestS3Upload(t *testing.T) {
	// Scenario: upload a watermarked WebP to the staging bucket.
	mock := &mockS3Client{}
	svc := &S3FileService{BucketName: "vectoricons-webp-staging", Client: mock}
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
	svc := &S3FileService{BucketName: "vectoricons-webp-staging", Client: mock}
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
// body under the configured local source root.
func TestS3Download(t *testing.T) {
	// Scenario: download a source SVG from the private bucket to local disk.
	objectKey := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg"
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`

	localRoot := t.TempDir()
	// Download writes to LocalSource/ObjectKey and does not create parent
	// directories, so the fixture pre-creates them.
	if err := os.MkdirAll(filepath.Join(localRoot, filepath.Dir(objectKey)), 0755); err != nil {
		t.Fatalf("failed to create local dirs: %v", err)
	}

	mock := &mockS3Client{
		getOutput: &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(svgContent))},
	}
	svc := &S3FileService{
		SourceBucket: "vectoricons-private",
		Config:       FileServiceConfig{LocalSource: localRoot},
		Client:       mock,
	}

	img, err := imagefile.NewImageFile(objectKey)
	if err != nil || img == nil {
		t.Fatalf("fixture parse failed: %v", err)
	}

	localPath, err := svc.Download(img, "")
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	if aws.StringValue(mock.getInput.Bucket) != "vectoricons-private" {
		t.Errorf("GetObject bucket = %q, want source bucket", aws.StringValue(mock.getInput.Bucket))
	}
	if aws.StringValue(mock.getInput.Key) != objectKey {
		t.Errorf("GetObject key = %q, want object key", aws.StringValue(mock.getInput.Key))
	}
	written, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("downloaded file missing: %v", err)
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

// TestS3Exists verifies the boolean contract: true when HeadObject succeeds,
// false (without error) when it fails.
func TestS3Exists(t *testing.T) {
	// Scenario: check for an already-converted WebP before reprocessing.
	objectKey := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-preview.webp"

	svc := &S3FileService{BucketName: "vectoricons-public", Client: &mockS3Client{}}
	exists, err := svc.Exists(objectKey)
	if err != nil || !exists {
		t.Errorf("Exists() = %v, %v for a present object; want true, nil", exists, err)
	}

	svc = &S3FileService{BucketName: "vectoricons-public", Client: &mockS3Client{headErr: errors.New("NotFound")}}
	exists, err = svc.Exists(objectKey)
	if err != nil || exists {
		t.Errorf("Exists() = %v, %v for a missing object; want false, nil", exists, err)
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
	svc := &S3FileService{BucketName: "vectoricons-private", Client: mock}

	svgOnly := func(key *string) bool { return strings.HasSuffix(*key, ".svg") }
	files, err := svc.ListFiles(ListFilesInput{}, svgOnly)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if aws.StringValue(mock.listInput.Bucket) != "vectoricons-private" {
		t.Errorf("list bucket = %q, want service bucket", aws.StringValue(mock.listInput.Bucket))
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
// matching the LocalFileService contract instead of panicking.
func TestS3ListFiles_NilFilter(t *testing.T) {
	// Scenario: a caller lists the bucket without any filtering.
	mock := &mockS3Client{
		listPages: []*s3.ListObjectsV2Output{
			{Contents: []*s3.Object{
				{Key: aws.String("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg")},
				{Key: aws.String("iconify/manifest.json")},
			}},
		},
	}
	svc := &S3FileService{BucketName: "vectoricons-private", Client: mock}

	files, err := svc.ListFiles(ListFilesInput{}, nil)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("ListFiles() with nil filter = %v, want all 2 objects", files)
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
