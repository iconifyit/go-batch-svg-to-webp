package fileservice

type TransferInput struct {
	Bucket         string `json:"bucket"`
	File           string `json:"file"`
	SourceFilePath string `json:"sourceFilePath"`
	TargetFilePath string `json:"targetFilePath"`
}
