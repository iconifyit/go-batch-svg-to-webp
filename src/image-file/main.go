package imagefile

import (
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

type ImageFile struct {
	URL               string
	ObjectKey         string
	InputPath         string
	Contributor       string
	ProductType       *string
	FamilyUniqueID    string
	SetUniqueID       *string
	Filename          string
	Extension         string
	Slug              string
	Stem              string
	IsValid           bool
	Error             string
	OptimizedImageKey string
	FileType          string
}

// // NewImageFile initializes and validates an ImageFile instance
// func xNewImageFile(inputURL string) (*ImageFile, error) {
// 	image := &ImageFile{
// 		URL:       inputURL,
// 		InputPath: inputURL,
// 	}

// 	log.Printf("NewImageFile : inputURL %v", inputURL)
// 	log.Printf("NewImageFile : image %v", fn.ToJSON(image))

// 	// Parse the URL
// 	parsedURL, err := url.Parse(inputURL)
// 	if err != nil {
// 		image.IsValid = false
// 		image.Error = "Invalid URL"
// 		return image, err
// 	}

// 	// Validate hostname
// 	if !strings.Contains(parsedURL.Hostname(), "vectopus") {
// 		image.IsValid = false
// 		image.Error = "Invalid domain"
// 		return image, errors.New("invalid domain")
// 	}

// 	// Define the regex pattern
// 	pattern := `^(?:https:\/\/[^/]*vectopus[^/]*\/)?([^/]+)\/(icons|illustrations|previews)\/([A-Z0-9]{12})(?:\/([A-Z0-9]{12}))?\/(.+)\.(png|jpg|jpeg|svg|webp)$`
// 	regex := regexp.MustCompile(pattern)

// 	// Match the URL against the regex
// 	matches := regex.FindStringSubmatch(inputURL)
// 	if len(matches) == 0 {
// 		image.IsValid = false
// 		image.Error = "Invalid URL format"
// 		return image, errors.New("invalid URL format")
// 	}

// 	// Extract fields
// 	image.IsValid = true
// 	image.Contributor = matches[1]
// 	if matches[2] != "previews" {
// 		image.ProductType = &matches[2]
// 	}
// 	image.FamilyUniqueID = matches[3]
// 	if matches[4] != "" {
// 		image.SetUniqueID = &matches[4]
// 	}
// 	image.Filename = matches[5] + "." + matches[6]
// 	image.Extension = matches[6]
// 	image.Slug = strings.TrimSuffix(path.Base(image.Filename), "."+matches[6])
// 	image.Stem = image.stripSizeFromSlug()
// 	image.ObjectKey = image.buildObjectKey()
// 	image.OptimizedImageKey = image.getOptimizedImageKey()
// 	image.FileType = image.Extension

//		return image, nil
//	}
func NewImageFile(inputPath string) (*ImageFile, error) {
	// Skip files starting with "."
	if len(inputPath) > 0 && inputPath[0] == '.' {
		log.Printf("Skipping hidden file: %s", inputPath)
		return nil, nil
	}

	image := &ImageFile{
		URL:       inputPath,
		InputPath: inputPath,
	}

	// Define regex patterns for different path types
	webSetPattern := `^https:\/\/[^/]*vectopus[^/]*\/([^/]+)\/(icons|illustrations)\/([A-Z0-9]{12})\/([A-Z0-9]{12})\/([^/]+\.(?:png|jpg|jpeg|svg|webp))$`
	webFamilyPattern := `^https:\/\/[^/]*vectopus[^/]*\/([^/]+)\/previews\/([A-Z0-9]{12})\/([^/]+\.(?:png|jpg|jpeg|svg|webp))$`
	localSetPattern := `([^/]+)\/(icons|illustrations)\/([A-Z0-9]{12})\/([A-Z0-9]{12})\/([^/]+\.(?:png|jpg|jpeg|svg|webp))$`
	localFamilyPattern := `([^/]+)\/previews\/([A-Z0-9]{12})\/([^/]+\.(?:png|jpg|jpeg|svg|webp))$`

	patterns := []struct {
		name    string
		regex   *regexp.Regexp
		handler func([]string) (*ImageFile, error)
	}{
		{
			name:  "Web Set",
			regex: regexp.MustCompile(webSetPattern),
			handler: func(matches []string) (*ImageFile, error) {
				return populateImageFile(image, matches, 1)
			},
		},
		{
			name:  "Web Family",
			regex: regexp.MustCompile(webFamilyPattern),
			handler: func(matches []string) (*ImageFile, error) {
				return populateImageFile(image, matches, 1)
			},
		},
		{
			name:  "Local Set",
			regex: regexp.MustCompile(localSetPattern),
			handler: func(matches []string) (*ImageFile, error) {
				return populateImageFile(image, matches, 0)
			},
		},
		{
			name:  "Local Family",
			regex: regexp.MustCompile(localFamilyPattern),
			handler: func(matches []string) (*ImageFile, error) {
				return populateImageFile(image, matches, 0)
			},
		},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(inputPath)
		if matches != nil {
			log.Printf("Matched pattern: %s", pattern.name)
			return pattern.handler(matches)
		}
	}

	// If no pattern matches, log and skip the file
	log.Printf("No pattern matched for path: %s", inputPath)
	return nil, nil
}

// Populate ImageFile properties based on regex matches
func populateImageFile(image *ImageFile, matches []string, startIndex int) (*ImageFile, error) {
	log.Printf("populateImageFile: matches=%v, startIndex=%d", matches, startIndex)

	// Ensure the matches length is sufficient
	requiredLength := startIndex + 5
	if len(matches) < requiredLength {
		return nil, fmt.Errorf(
			"unexpected match length: got %d, expected at least %d. Matches: %v",
			len(matches), requiredLength, matches,
		)
	}

	// Populate fields based on match indices
	image.IsValid = true
	image.Contributor = matches[startIndex+1] // Adjusted to capture group index
	if matches[startIndex+2] != "previews" {
		image.ProductType = &matches[startIndex+2]
	}
	image.FamilyUniqueID = matches[startIndex+3]
	if matches[startIndex+4] != "" {
		image.SetUniqueID = &matches[startIndex+4]
	}
	image.Filename = matches[startIndex+5]
	image.Extension = path.Ext(image.Filename)[1:]
	image.Slug = strings.TrimSuffix(path.Base(image.Filename), "."+image.Extension)
	image.Stem = image.stripSizeFromSlug()
	image.ObjectKey = image.buildObjectKey()
	image.OptimizedImageKey = image.getOptimizedImageKey()
	image.FileType = image.Extension

	return image, nil
}

// getOptimizedImageKey generates the optimized image key
func (image *ImageFile) getOptimizedImageKey() string {
	if !image.IsValid {
		return ""
	}
	if image.ProductType != nil && image.SetUniqueID != nil {
		return image.Contributor + "/" + *image.ProductType + "/" + image.FamilyUniqueID + "/" + *image.SetUniqueID + "/" + image.Stem + ".webp"
	}
	return image.Contributor + "/preview/" + image.FamilyUniqueID + "/" + image.Stem + ".webp"
}

// buildObjectKey constructs the object key
func (image *ImageFile) buildObjectKey() string {
	if !image.IsValid {
		return ""
	}
	if image.ProductType != nil && image.SetUniqueID != nil {
		return image.Contributor + "/" + *image.ProductType + "/" + image.FamilyUniqueID + "/" + *image.SetUniqueID + "/" + image.Filename
	}
	return image.Contributor + "/preview/" + image.FamilyUniqueID + "/" + image.Filename
}

// stripSizeFromSlug removes size-related suffixes from the slug
func (image *ImageFile) stripSizeFromSlug() string {
	if !image.IsValid {
		return ""
	}

	if !strings.Contains(image.Slug, "@") && !strings.Contains(image.Slug, "-") {
		return image.Slug
	}

	if strings.Contains(image.Slug, "@") {
		return strings.Split(image.Slug, "@")[0]
	}

	// Remove patterns like @2x, @2x3x, -2x, -10x20, etc.
	sizePattern := regexp.MustCompile(`(@\d+x\d+-\dx|@\d+x\d+|@\dx|-\d+x\d+|-\d+)$`)
	return sizePattern.ReplaceAllString(image.Slug, "")
}

// ToJSON converts the object to a JSON-compatible map
func (image *ImageFile) ToJSON() map[string]interface{} {
	if !image.IsValid {
		return map[string]interface{}{
			"url":   image.URL,
			"error": image.Error,
		}
	}
	return map[string]interface{}{
		"url":               image.URL,
		"objectKey":         image.ObjectKey,
		"contributor":       image.Contributor,
		"productType":       image.ProductType,
		"familyUniqueID":    image.FamilyUniqueID,
		"setUniqueID":       image.SetUniqueID,
		"filename":          image.Filename,
		"extension":         image.Extension,
		"slug":              image.Slug,
		"stem":              image.Stem,
		"isValid":           image.IsValid,
		"optimizedImageKey": image.OptimizedImageKey,
		"fileType":          image.FileType,
	}
}

// ToSnakeCase converts the object to a snake_case JSON-compatible map
func (image *ImageFile) ToSnakeCase() map[string]interface{} {
	return map[string]interface{}{
		"url":                 image.URL,
		"object_key":          image.ObjectKey,
		"contributor":         image.Contributor,
		"product_type":        image.ProductType,
		"family_unique_id":    image.FamilyUniqueID,
		"set_unique_id":       image.SetUniqueID,
		"filename":            image.Filename,
		"extension":           image.Extension,
		"slug":                image.Slug,
		"stem":                image.Stem,
		"is_valid":            image.IsValid,
		"optimized_image_key": image.OptimizedImageKey,
	}
}

// Exists checks if the image file exists
func (image *ImageFile) Exists() (bool, error) {
	info, err := os.Stat(image.InputPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}
