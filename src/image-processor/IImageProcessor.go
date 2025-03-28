package imageprocessor

import imagefile "github.com/iconifyit/go-batch-svg-to-webp/src/image-file"

type IImageProcessor interface {
	IsLocalRun() bool
	SetupLogging() error
	ShouldInclude(filePath *string) bool
	ListFiles() ([]string, error)
	ListDirs(rootdir string) ([]string, error)
	IsValidContributor(username string) bool
	ListFilesForContributor(contributor string) ([]string, error)
	ImageFiles(files []string) []imagefile.ImageFile
	ProcessFiles() error
	Run() error
}
