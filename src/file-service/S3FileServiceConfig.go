package fileservice

import "github.com/aws/aws-sdk-go/aws/session"

type S3FileServiceConfig struct {
	UUID         string
	Session      *session.Session
	Config       FileServiceConfig
	BucketName   string
	ObjectKey    string
	SourceBucket string
	TargetBucket string
}
