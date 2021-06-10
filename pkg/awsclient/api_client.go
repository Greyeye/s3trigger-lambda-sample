package awsclient

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Clients struct {
	S3client S3clientIface
}

// S3clientIface is interface to mock aws s3 services
type S3clientIface interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	UploadPart(context.Context, *s3.UploadPartInput, ...func(*s3.Options)) (*s3.UploadPartOutput, error)
	CreateMultipartUpload(context.Context, *s3.CreateMultipartUploadInput, ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error)
	CompleteMultipartUpload(context.Context, *s3.CompleteMultipartUploadInput, ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error)
	AbortMultipartUpload(context.Context, *s3.AbortMultipartUploadInput, ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error)
}

// New will setup the AWS client service with s3client interface
func New(ctx context.Context) (*Clients, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	// init all client connections

	return &Clients{
		S3client: s3.NewFromConfig(cfg),
	}, nil
}
