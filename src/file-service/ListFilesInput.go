package fileservice

type ListFilesInput struct {
	SourceRoot string `json:"sourceRoot"`
	FileType   string `json:"fileType"`
	Include    []string
	Exclude    []string
}
