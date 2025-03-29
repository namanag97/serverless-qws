package processor

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/yourusername/transcription-service/internal/awsclient"
	"github.com/yourusername/transcription-service/internal/elevenlabs"
	"github.com/yourusername/transcription-service/internal/model"
)

// Processor handles the transcription business logic
type Processor struct {
	s3Client            *s3.Client
	dynamoDBClient      *dynamodb.Client
	elevenlabsClient    *elevenlabs.Client
	s3Operations        *awsclient.S3Operations
	dynamoDBOperations  *awsclient.DynamoDBOperations
	tableName           string
	outputBucket        string
}

// NewProcessor creates a new processor instance
func NewProcessor(
	s3Client *s3.Client,
	dynamoDBClient *dynamodb.Client,
	elevenlabsClient *elevenlabs.Client,
	tableName string,
	outputBucket string,
) *Processor {
	return &Processor{
		s3Client:           s3Client,
		dynamoDBClient:     dynamoDBClient,
		elevenlabsClient:   elevenlabsClient,
		s3Operations:       awsclient.NewS3Operations(s3Client),
		dynamoDBOperations: awsclient.NewDynamoDBOperations(dynamoDBClient, tableName),
		tableName:          tableName,
		outputBucket:       outputBucket,
	}
}

// ProcessFile processes an audio file from S3 for transcription
func (p *Processor) ProcessFile(ctx context.Context, bucket, key string) error {
	startTime := time.Now()
	fileID := key // Using the S3 key as the file identifier
	
	log.Printf("Starting processing of file: %s", fileID)
	
	// Check if this file has already been processed
	existingItem, err := p.dynamoDBOperations.GetTranscriptionItem(ctx, fileID)
	if err != nil {
		return fmt.Errorf("error checking for existing transcription: %w", err)
	}
	
	if existingItem != nil && (existingItem.Status == model.StatusCompleted || existingItem.Status == model.StatusInProgress) {
		log.Printf("File %s is already processed or in progress, skipping", fileID)
		return nil
	}
	
	// Create or update the DynamoDB item to indicate processing has started
	if existingItem == nil {
		// Create new item
		item := &model.TranscriptionItem{
			FileIdentifier: fileID,
			Status:         model.StatusInProgress,
			SourceBucket:   bucket,
			SourceKey:      key,
		}
		
		err = p.dynamoDBOperations.CreateTranscriptionItem(ctx, item)
		if err != nil {
			return fmt.Errorf("failed to create DynamoDB item: %w", err)
		}
	} else {
		// Update existing item
		err = p.dynamoDBOperations.UpdateTranscriptionItemStatus(
			ctx, fileID, model.StatusInProgress, "", "", "", 0)
		if err != nil {
			return fmt.Errorf("failed to update DynamoDB item status: %w", err)
		}
	}
	
	// Generate a pre-signed URL for the audio file
	audioURL, err := p.s3Operations.GeneratePresignedURL(ctx, bucket, key, 3600) // 1 hour expiration
	if err != nil {
		// Update DynamoDB to indicate failure
		updateErr := p.dynamoDBOperations.UpdateTranscriptionItemStatus(
			ctx, fileID, model.StatusFailed, "", "", fmt.Sprintf("Failed to generate pre-signed URL: %v", err), 0)
		if updateErr != nil {
			log.Printf("Failed to update DynamoDB item status: %v", updateErr)
		}
		
		return fmt.Errorf("failed to generate pre-signed URL: %w", err)
	}
	
	// Call ElevenLabs API for transcription
	log.Printf("Sending audio to ElevenLabs API for transcription")
	transcriptionResp, err := p.elevenlabsClient.TranscribeAudio(ctx, audioURL)
	if err != nil {
		// Update DynamoDB to indicate failure
		updateErr := p.dynamoDBOperations.UpdateTranscriptionItemStatus(
			ctx, fileID, model.StatusFailed, "", "", fmt.Sprintf("Transcription API error: %v", err), 0)
		if updateErr != nil {
			log.Printf("Failed to update DynamoDB item status: %v", updateErr)
		}
		
		return fmt.Errorf("transcription API error: %w", err)
	}
	
	// Calculate processing time
	processingTime := time.Since(startTime).Seconds()
	
	// If output bucket is specified, store the transcript in S3
	var outputLocation string
	if p.outputBucket != "" && transcriptionResp.Text != "" {
		// Generate output key based on input filename
		baseName := filepath.Base(key)
		extLess := baseName[:len(baseName)-len(filepath.Ext(baseName))]
		outputKey := fmt.Sprintf("transcripts/%s.txt", extLess)
		
		// Upload transcript to S3
		err = p.s3Operations.UploadText(ctx, p.outputBucket, outputKey, transcriptionResp.Text)
		if err != nil {
			log.Printf("Warning: Failed to upload transcript to S3: %v", err)
			// Continue processing instead of failing
		} else {
			outputLocation = fmt.Sprintf("s3://%s/%s", p.outputBucket, outputKey)
			log.Printf("Uploaded transcript to %s", outputLocation)
		}
	}
	
	// Update DynamoDB with successful result
	err = p.dynamoDBOperations.UpdateTranscriptionItemStatus(
		ctx,
		fileID,
		model.StatusCompleted,
		transcriptionResp.Text,
		outputLocation,
		"", // No error message
		processingTime,
	)
	if err != nil {
		log.Printf("Warning: Failed to update DynamoDB with successful result: %v", err)
		// Continue despite error since transcription was successful
	}
	
	log.Printf("Successfully processed file %s in %.2f seconds", fileID, processingTime)
	return nil