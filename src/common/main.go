package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

// ToJSON converts any interface{} to a pretty-printed JSON string.
// Returns the JSON string or an empty string if the conversion fails.
func ToJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ") // Pretty print with indentation
	if err != nil {
		log.Printf("Error converting to JSON: %s", err)
		return ""
	}
	return string(jsonBytes)
}

// func ToJSON(input interface{}) string {
// 	json, _ := json.Marshal(input)
// 	return string(json)
// }

// StringInSlice checks if a string exists in a slice.
func StringInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// ObjectKeyAsType returns the object key with a new file type
func AsType(filepath, newType string) (string, error) {
	if filepath == "" || newType == "" {
		return "", fmt.Errorf("ObjectKey or newKey is empty")
	}
	return strings.TrimSuffix(filepath, path.Ext(filepath)) + "." + newType, nil
}

// ObjectKeyWithSuffix returns the object key with a suffix
func AddSuffix(filepath, suffix string) (string, error) {
	if filepath == "" || suffix == "" {
		return "", fmt.Errorf("ObjectKey or suffix is empty")
	}
	ext := path.Ext(filepath)
	return strings.TrimSuffix(filepath, ext) + "-" + suffix + ext, nil
}
