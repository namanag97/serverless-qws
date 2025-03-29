package handler

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/yourusername/transcription-service/internal/processor"
)

// Handler manages the Lambda function handler
type Handler struct {
	processor *processor.Processor
}

// NewHandler creates a new handler instance
func NewHandler(proc *processor.Processor) *Handler {
	return &Handler{
		processor: proc,
	}
}

// HandleS3Event processes S3 events from Lambda
func (h *Handler) HandleS3Event(ctx context.Context, s3Event events.S3Event) error {
	log.Printf("Received %d record(s) from S3", len(s3Event.Records))
	
	for i, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.URLDecodedKey
		
		log.Printf("[%d/%d] Processing file: s3://%s/%s", i+1, len(s3Event.Records), bucket, key)
		
		// Validate file extension
		if !h.isValidAudioFile(key) {
			log.Printf("Skipping file with unsupported extension: %s", key)
			continue
		}
		
		// Process the file
		err := h.processor.ProcessFile(ctx, bucket, key)
		if err != nil {
			log.Printf("ERROR processing %s: %v", key, err)
			// Decision: Return error to trigger Lambda retry, or continue with next file?
			// Here we continue with next files rather than failing the entire batch
			continue
		}
		
		log.Printf("Successfully processed file: %s", key)
	}
	
	return nil
}

// isValidAudioFile checks if the file has a supported audio extension
func (h *Handler) isValidAudioFile(key string) bool {
	ext := strings.ToLower(filepath.Ext(key))
	
	// List of supported audio extensions
	supportedExts := map[string]bool{
		".aac":  true,
		".mp3":  true,
		".wav":  true,
		".flac": true,
		".ogg":  true,
		".m4a":  true,
	}
	
	return supportedExts[ext]
}