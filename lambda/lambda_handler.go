package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"log"
)

func (h *lambdaHandler) handler(ctx context.Context, S3events events.S3Event) error {

	// S3 trigger may contain multiple events.
	for _, record := range S3events.Records {
		err := s3handler(ctx, h.awsClient, record, h.destinationBucket)
		if err != nil {
			log.Println("S3 Handling error")
			return err
		}

	}
	return nil

}
