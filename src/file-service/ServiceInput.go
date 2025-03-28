package fileservice

import "github.com/aws/aws-sdk-go/aws/session"

type ServiceInput struct {
	UUID           string           `json:"uuid"`
	IsLocal        bool             `json:"is_local"`
	UploadToS3     bool             `json:"upload_to_s3"`
	Session        *session.Session `json:"-"`
	SourceRoot     string           `json:"sourceRoot"`
	TargetRoot     string           `json:"targetRoot"`
	SourceFilePath string           `json:"sourceFilePath"`
	TargetFilePath string           `json:"targetFilePath"`
	File           string           `json:"file"`
}
