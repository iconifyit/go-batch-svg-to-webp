package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestToJSON verifies that a realistic struct is rendered as pretty-printed
// JSON that round-trips back to the same values.
func TestToJSON(t *testing.T) {
	// Scenario: a contributor record with two fields is marshaled to JSON.
	input := struct {
		Username string `json:"username"`
		IconSets int    `json:"icon_sets"`
	}{Username: "iconify", IconSets: 42}

	got := ToJSON(input)

	var decoded struct {
		Username string `json:"username"`
		IconSets int    `json:"icon_sets"`
	}
	if err := json.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("ToJSON produced invalid JSON: %v\noutput: %s", err, got)
	}
	if decoded.Username != "iconify" || decoded.IconSets != 42 {
		t.Errorf("ToJSON round-trip = %+v, want {iconify 42}", decoded)
	}
}

// TestToJSON_Unmarshalable verifies the error contract: values that cannot be
// marshaled (a channel) yield an empty string rather than a panic.
func TestToJSON_Unmarshalable(t *testing.T) {
	// Scenario: a channel cannot be represented in JSON.
	if got := ToJSON(make(chan int)); got != "" {
		t.Errorf("ToJSON(chan) = %q, want empty string", got)
	}
}

// TestStringInSlice covers the membership contract: present, absent, and
// empty-slice inputs.
func TestStringInSlice(t *testing.T) {
	// Scenario: contributor exclude-list lookups.
	contributors := []string{"iconify", "vectopus", "diversity-avatars"}

	if !StringInSlice("vectopus", contributors) {
		t.Errorf("StringInSlice(vectopus) = false, want true")
	}
	if StringInSlice("unknown-vendor", contributors) {
		t.Errorf("StringInSlice(unknown-vendor) = true, want false")
	}
	if StringInSlice("iconify", []string{}) {
		t.Errorf("StringInSlice on empty slice = true, want false")
	}
}

// TestCopyFile verifies a real file's bytes are copied to the destination.
func TestCopyFile(t *testing.T) {
	// Scenario: copy an SVG source file into a work directory.
	dir := t.TempDir()
	src := filepath.Join(dir, "coffee-cup.svg")
	dst := filepath.Join(dir, "work", "coffee-cup.svg")
	content := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><circle cx="12" cy="12" r="10"/></svg>`)

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("failed to seed source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatalf("failed to create destination dir: %v", err)
	}

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("copied content = %q, want %q", got, content)
	}
}

// TestCopyFile_MissingSource verifies the error contract for a nonexistent
// source path.
func TestCopyFile_MissingSource(t *testing.T) {
	// Scenario: the source SVG was never downloaded.
	dir := t.TempDir()
	err := CopyFile(filepath.Join(dir, "does-not-exist.svg"), filepath.Join(dir, "out.svg"))
	if err == nil {
		t.Fatal("CopyFile() with missing source: expected error, got nil")
	}
}

// TestAsType verifies extension replacement on object keys.
func TestAsType(t *testing.T) {
	// Scenario: derive the WebP object key from an SVG object key.
	got, err := AsType("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.svg", "webp")
	if err != nil {
		t.Fatalf("AsType() error = %v", err)
	}
	want := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.webp"
	if got != want {
		t.Errorf("AsType() = %q, want %q", got, want)
	}
}

// TestAsType_EmptyArgs verifies both empty-argument error paths.
func TestAsType_EmptyArgs(t *testing.T) {
	// Scenario: guard against empty object keys or types.
	if _, err := AsType("", "webp"); err == nil {
		t.Error("AsType with empty path: expected error, got nil")
	}
	if _, err := AsType("coffee-cup.svg", ""); err == nil {
		t.Error("AsType with empty type: expected error, got nil")
	}
}

// TestAddSuffix verifies the size-variant suffix is inserted before the
// extension.
func TestAddSuffix(t *testing.T) {
	// Scenario: derive the thumbnail variant key for a WebP file.
	got, err := AddSuffix("iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup.webp", "thumbnail")
	if err != nil {
		t.Fatalf("AddSuffix() error = %v", err)
	}
	want := "iconify/icons/2C11DB2D5F79/B24091F3DF3E/coffee-cup-thumbnail.webp"
	if got != want {
		t.Errorf("AddSuffix() = %q, want %q", got, want)
	}
}

// TestAddSuffix_EmptyArgs verifies both empty-argument error paths.
func TestAddSuffix_EmptyArgs(t *testing.T) {
	// Scenario: guard against empty object keys or suffixes.
	if _, err := AddSuffix("", "thumbnail"); err == nil {
		t.Error("AddSuffix with empty path: expected error, got nil")
	}
	if _, err := AddSuffix("coffee-cup.webp", ""); err == nil {
		t.Error("AddSuffix with empty suffix: expected error, got nil")
	}
}
