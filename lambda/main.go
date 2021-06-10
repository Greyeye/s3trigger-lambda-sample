package main

import (
	"context"
	"github.com/Greyeye/s3trigger-lambda-sample/pkg/awsclient"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"os"
)

type lambdaHandler struct {
	awsClient         *awsclient.Clients
	destinationBucket string
}

func main() {
	ac, err := awsclient.New(context.Background())
	if err != nil {
		log.Fatal("aws sdk client config error ", err.Error())
		return
	}
	if os.Getenv("destinationBucket") == "" {
		log.Fatal("please check destinationBucket environment variables")
		return
	}
	lh := &lambdaHandler{
		awsClient:         ac,
		destinationBucket: os.Getenv("destinationBucket"),
	}
	lambda.Start(lh.handler)
}
