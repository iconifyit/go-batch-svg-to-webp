package imagefile

// import (
// 	"testing"
// )

// func TestGetRelativePath(t *testing.T) {

// 	tests := []struct {
// 		name     string
// 		fullpath string
// 		want     string
// 	}{
// 		{
// 			name:     "local source",
// 			fullpath: "/local/source/contributor1/family/set/file.svg",
// 			want:     "contributor1/family/set/file.svg",
// 		},
// 		{
// 			name:     "remote source",
// 			fullpath: "/remote/source/contributor1/family/set/file.svg",
// 			want:     "contributor1/family/set/file.svg",
// 		},
// 		{
// 			name:     "starts with contributor",
// 			fullpath: "contributor1/family/set/file.svg",
// 			want:     "contributor1/family/set/file.svg",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			imageFile, err := NewImageFile(
// 				map[string]string{
// 					"work_dir":   "/ramdisk",
// 					"target_dir": "/output",
// 				},
// 				"contributor1",
// 				tt.fullpath,
// 			)

// 			if err != nil {
// 				t.Fatalf("NewImageFile() returned an error: %v", err)
// 			}

// 			_, err = imageFile.GetRelativePath()
// 			if err != nil {
// 				t.Fatalf("GetRelativePath() returned an error: %v", err)
// 			}

// 			if imageFile.RelativePath != tt.want {
// 				t.Errorf("GetRelativePath() = %s; want %s", imageFile.RelativePath, tt.want)
// 			}
// 		})
// 	}
// }

// func TestGetWorkSourcePath(t *testing.T) {
// 	imageFile, err := NewImageFile(
// 		map[string]string{
// 			"work_dir":   "/ramdisk",
// 			"target_dir": "/output",
// 			"uuid":       "12345",
// 			"is_local":   "true",
// 		},
// 		"contributor1",
// 		"/local/source/contributor1/path/to/file.svg",
// 	)

// 	if err != nil {
// 		t.Fatalf("NewImageFile() returned an error: %v", err)
// 	}

// 	expected := "/ramdisk/12345/source/contributor1/path/to/file.svg"
// 	result := imageFile.GetWorkSourcePath()

// 	if result != expected {
// 		t.Errorf("GetWorkSourcePath() = %s; want %s", result, expected)
// 	}
// }

// func TestGetIntermediatePath(t *testing.T) {
// 	imageFile, err := NewImageFile(
// 		map[string]string{
// 			"work_dir":   "/ramdisk",
// 			"target_dir": "/output",
// 			"uuid":       "12345",
// 			"is_local":   "true",
// 		},
// 		"contributor1",
// 		"/local/source/contributor1/path/to/file.svg",
// 	)

// 	if err != nil {
// 		t.Fatalf("NewImageFile() returned an error: %v", err)
// 	}

// 	expected := "/ramdisk/12345/intermediate/contributor1/path/to/file.png"
// 	result := imageFile.GetIntermediatePath()

// 	if result != expected {
// 		t.Errorf("GetIntermediatePath() = %s; want %s", result, expected)
// 	}
// }

// func TestGetWorkOutputPath(t *testing.T) {
// 	imageFile, err := NewImageFile(
// 		map[string]string{
// 			"work_dir":   "/ramdisk",
// 			"target_dir": "/output",
// 			"uuid":       "12345",
// 			"is_local":   "true",
// 		},
// 		"contributor1",
// 		"/local/source/contributor1/path/to/file.svg",
// 	)

// 	if err != nil {
// 		t.Fatalf("NewImageFile() returned an error: %v", err)
// 	}

// 	expected := "/ramdisk/12345/output/contributor1/path/to/file.webp"
// 	result := imageFile.GetWorkOutputPath()

// 	if result != expected {
// 		t.Errorf("GetWorkOutputPath() = %s; want %s", result, expected)
// 	}
// }

// func TestGetTargetPath(t *testing.T) {
// 	imageFile, err := NewImageFile(
// 		map[string]string{
// 			"work_dir":   "/ramdisk",
// 			"target_dir": "test/output/",
// 			"uuid":       "12345",
// 			"is_local":   "true",
// 		},
// 		"contributor1",
// 		"/ramdisk/12345/source/contributor1/path/to/file.svg",
// 	)

// 	if err != nil {
// 		t.Fatalf("NewImageFile() returned an error: %v", err)
// 	}

// 	expected := "test/output/12345/contributor1/path/to/file.webp"
// 	result := imageFile.GetTargetPath()

// 	if result != expected {
// 		t.Errorf("GetTargetPath() = %s; want %s", result, expected)
// 	}
// }

// func TestGetWebPPath(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		fpath    string
// 		expected string
// 	}{
// 		{
// 			name:     "webp with RelativePath",
// 			fpath:    "contributor1/family/set/file.svg",
// 			expected: "contributor1/family/set/file.webp",
// 		},
// 		{
// 			name:     "webp with WorkingOutputPath",
// 			fpath:    "/ramdisk/12345/output/contributor1/family/set/file.svg",
// 			expected: "/ramdisk/12345/output/contributor1/family/set/file.webp",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			imageFile, err := NewImageFile(
// 				map[string]string{
// 					"work_dir":   "/ramdisk",
// 					"target_dir": "/output",
// 					"uuid":       "12345",
// 					"is_local":   "true",
// 				},
// 				"contributor1",
// 				"/local/source/contributor1/path/to/file.svg",
// 			)

// 			if err != nil {
// 				t.Fatalf("NewImageFile() returned an error: %v", err)
// 			}

// 			result := imageFile.GetWebPPath(tt.fpath)
// 			if result != tt.expected {
// 				t.Errorf("GetWebPPath() = %s; want %s", result, tt.expected)
// 			}
// 		})
// 	}
// }

// func TestGetRelativePath_ContributorNotFound(t *testing.T) {
// 	_, err := NewImageFile(
// 		map[string]string{
// 			"work_dir":   "/ramdisk",
// 			"target_dir": "/output",
// 		},
// 		"contributor1",
// 		"/base/path/no-contributor/path/to/file.svg",
// 	)

// 	expectedErr := "contributor contributor1 not found in path /base/path/no-contributor/path/to/file.svg"
// 	if err.Error() != expectedErr {
// 		t.Errorf("GetRelativePath() error = %v; want %v", err, expectedErr)
// 	}
// }

// func TestNewImageFile(t *testing.T) {
// 	imageFile, err := NewImageFile(
// 		map[string]string{
// 			"work_dir":   "/ramdisk",
// 			"target_dir": "/output",
// 		},
// 		"contributor1",
// 		"/local/source/contributor1/path/to/file.svg",
// 	)

// 	if err != nil {
// 		t.Fatalf("NewImageFile() returned an error: %v", err)
// 	}

// 	_, err = imageFile.GetRelativePath()
// 	if err != nil {
// 		t.Fatalf("GetRelativePath() returned an error: %v", err)
// 	}

// 	expectedRelativePath := "contributor1/path/to/file.svg"
// 	if imageFile.RelativePath != expectedRelativePath {
// 		t.Errorf("NewImageFile() RelativePath = %s; want %s", imageFile.RelativePath, expectedRelativePath)
// 	}
// }
