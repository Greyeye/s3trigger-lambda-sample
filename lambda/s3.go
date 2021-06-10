package main

import (
	"context"
	"fmt"
	"github.com/Greyeye/s3trigger-lambda-sample/pkg/awsclient"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	"log"
	"net/url"
	"sync"
)

// FakeWriterAt returns the writer for a single threat writer.
// credit to
// https://dev.to/flowup/using-io-reader-io-writer-in-go-to-stream-data-3i7b
type FakeWriterAt struct {
	w io.Writer
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we forced sequential downloads
	return fw.w.Write(p)
}

// s3handler is the handler function to start download/upload goroutine.
func s3handler(ctx context.Context, a *awsclient.Clients, record events.S3EventRecord, destinationBucket string) error {
	s3rec := record.S3
	fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3rec.Bucket.Name, s3rec.Object.Key)
	pr, pw := io.Pipe()
	wg := sync.WaitGroup{}
	wg.Add(2)
	// run downloader and uploader in tandem.
	// it is connected via io.Pipe() and stream byte data between each other.
	// eliminating need the step to download S3 object to local storage then upload.
	go downloader(ctx, a, pw, &wg, s3rec)
	go uploader(ctx, a, &wg, pr, destinationBucket, s3rec)
	wg.Wait()

	return nil
}

// downloader object from S3 and send to PipeWriter.
func downloader(ctx context.Context, a *awsclient.Clients, pw *io.PipeWriter, wg *sync.WaitGroup, s3rec events.S3Entity) {
	defer func() {
		wg.Done()
		pw.Close()
	}()
	// submitted payload may escape email at sign `@` with urlescape, need to unescape it.
	keyEscapaed, err := url.QueryUnescape(s3rec.Object.Key)
	if err != nil {
		log.Println("failed to parse URL key: ", err)
		return
	}
	file := &s3.GetObjectInput{Bucket: aws.String(s3rec.Bucket.Name), Key: aws.String(keyEscapaed)}
	downloader := manager.NewDownloader(a.S3client)
	// limit concurrency to 1, otherwise downloaded stream will be out of orders.
	downloader.Concurrency = 1
	_, err = downloader.Download(ctx, FakeWriterAt{pw}, file)
	if err != nil {
		fmt.Println("Download failed:", err.Error())
	}
}

// uploader receive io.pipe buffer to destination S3 bucket.
func uploader(ctx context.Context, a *awsclient.Clients, wg *sync.WaitGroup, pr *io.PipeReader, destinationBucket string, s3rec events.S3Entity) {
	uploader := manager.NewUploader(a.S3client)
	uploader.Concurrency = 1
	defer func() {
		wg.Done()
		pr.Close()
	}()
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Body:   pr,
		Bucket: aws.String(destinationBucket),
		Key:    aws.String(s3rec.Object.Key),
		// ContentType: aws.String("text/csv"),
		// ideally, content Type should be set, but events.S3Entity does not contain metadata.
		// hardcode the content type, or build another parameters to setup contentType.
		// default ContentType will be used without specifying the value. (eg application/octet-stream)
	})
	if err != nil {
		fmt.Println("uploading failed: ", s3rec.Object.Key)
		fmt.Println(err.Error())
	}
}
