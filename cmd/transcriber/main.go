package main

import (
	"context"
	"log"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	log.Printf("Received S3 event with %d records", len(s3Event.Records))
	for _, record := range s3Event.Records {
		log.Printf("Processing file: s3://%s/%s", record.S3.Bucket.Name, record.S3.Object.Key)
	}
	return nil
}

func main() {
	log.Println("Starting Lambda function")
	lambda.Start(HandleRequest)
}
