package main

import (
	"bytes"
	"context"
	"github.com/Greyeye/s3trigger-lambda-sample/pkg/awsclient"
	"github.com/Greyeye/s3trigger-lambda-sample/pkg/awsclient/mocks"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

func TestDownloader(t *testing.T) {
	testFilename := "test.txt"
	testWordList := "Hi there, this is a test file"
	// prepare mock for GetObject
	mockAWS := new(mocks.S3clientIface)
	mockAWS.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
		ContentLength: 10,
		Body:          nopCloser{bytes.NewBufferString(testWordList)},
	}, nil)

	ac := &awsclient.Clients{}
	// override s3Client with mock AWS interface
	ac.S3client = mockAWS
	pr, pw := io.Pipe()
	wg := sync.WaitGroup{}

	wg.Add(2)
	// start the process with dummy S3Object metadata.
	// Object's content is submitted as part of unit test. (see testWordList)
	go downloader(context.TODO(), ac, pw, &wg,
		events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "s3buckettest",
			},
			Object: events.S3Object{
				Key:  "dummyfile.txt",
				Size: 10,
			},
		})
	// create temp file to store data.
	f, err := os.Create(testFilename)
	if err != nil {
		t.Fatal(err)
	}
	// use io.copy to write data from downloader to the file.
	go func() {
		io.Copy(f, pr)
		wg.Done()
	}()
	wg.Wait()

	//open unzipped file and verify the content.
	data, err := ioutil.ReadFile(testFilename)
	assert.Nil(t, err)
	assert.Equal(t, testWordList, string(data))
	// remove file to clean up
	os.Remove(testFilename)
}

func TestBufferUploader(t *testing.T) {
	//testFilename := "test.txt"
	testWordList := "Hi there, this is a test file"
	// prepare mock for GetObject
	mockAWS := new(mocks.S3clientIface)
	mockAWS.On("PutObject", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		&s3.PutObjectOutput{
			ETag:      aws.String("dummyTag"),
			VersionId: aws.String("dummy version"),
		}, nil)
	mockAWS.On("GetObject", mock.Anything, mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{
		ContentLength: 10,
		Body:          nopCloser{bytes.NewBufferString(testWordList)},
	}, nil)
	//CreateMultipartUpload(context.Context, *s3.CreateMultipartUploadInput, ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput
	ac := &awsclient.Clients{}
	// override s3Client with mock AWS interface
	ac.S3client = mockAWS
	pr, pw := io.Pipe()
	wg := sync.WaitGroup{}

	wg.Add(2)

	go downloader(context.TODO(), ac, pw, &wg,
		events.S3Entity{
			Bucket: events.S3Bucket{
				Name: "s3buckettest",
			},
			Object: events.S3Object{
				Key:  "dummyfile.txt",
				Size: 10,
			},
		})

	go uploader(context.TODO(), ac, &wg, pr, "destinationBucket", events.S3Entity{
		Bucket: events.S3Bucket{
			Name: "s3buckettest",
		},
		Object: events.S3Object{
			Key:  "dummyfile.txt",
			Size: 5,
		},
	})

	// use io.copy to write data from downloader to the file.
	wg.Wait()

}
