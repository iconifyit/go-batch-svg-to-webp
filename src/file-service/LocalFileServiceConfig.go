package fileservice

import (
	"github.com/aws/aws-sdk-go/aws/session"
)

type LocalFileServiceConfig struct {
	UUID       string
	Session    *session.Session
	Config     FileServiceConfig
	SourceRoot string
	TargetRoot string
}
