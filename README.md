# S3 Event Trigger Lambda Code Sample

This project is a code sample to handle S3 Event to Copy to other bucket.  
Uses io.pipe to copy dynamically from the event source bucket to specified destination.  


Tested with a large file (upto 2GigByte) with 256MByte Memory Lambda. 



## Building and testing the code
### Build
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main ./... &&zip -r main.zip main
```
### Test
```bash
go test -coverprofile=coverage.out |go tool cover -html=coverage.out
```

## Required Lambda Configuration
Environment Variable **"destinationBucket"** is a mandatory requirement. Please add it on your lambda configuration.


## How to setup a S3 Bucket Trigger Configurations
1. Create Lambda, upload the zipped code, setup environment variable "destinationBucket"
2. Make sure Lambda has assigned with the IAM Role with GetObject to the source bucket, and PutItem to destination bucket.  
3. Goto Source Bucket, and configure [Event notifications](https://docs.aws.amazon.com/AmazonS3/latest/userguide/NotificationHowTo.html)
   and add Put event, point to ARN or the name of the lambda you saved in the step 1.  
4. Upload a test file to the source bucket and see it gets copied to the destination bucket.

