package awsclient

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3")

// S3Operations provides operations for working with S3
type S3Operations struct {
	client *s3.Client
}

// NewS3Operations creates a new S3Operations instance
func NewS3Operations(client *s3.Client) *S3Operations {
	return &S3Operations{
		client: client,
	}
}

// DownloadFile downloads a file from S3 to a local temp file
func (s *S3Operations) DownloadFile(ctx context.Context, bucket, key string) (string, error) {
	log.Printf("Downloading file from s3://%s/%s", bucket, key)
	
	// Create a temporary file
	tmpDir := os.TempDir()
	fileName := filepath.Base(key)
	tempFilePath := filepath.Join(tmpDir, fmt.Sprintf("download-%s", fileName))
	
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()
	
	// Make sure we clean up the temp file if something goes wrong
	defer func() {
		if err != nil {
			os.Remove(tempFilePath)
		}
	}()
	
	// Download file from S3
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer resp.Body.Close()
	
	// Copy the response body to the temp file
	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy S3 object to temp file: %w", err)
	}
	
	log.Printf("Downloaded %d bytes to %s", written, tempFilePath)
	return tempFilePath, nil