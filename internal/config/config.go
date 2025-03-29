package config

import (
	"errors"
	"os"
)

// Config holds the application configuration
type Config struct {
	// AWS region for all service clients
	AWSRegion string
	
	// DynamoDB table name for state tracking
	DynamoDBTableName string
	
	// Secrets Manager secret name containing the ElevenLabs API key
	ElevenLabsSecretName string
	
	// Optional output S3 bucket (if storing full transcripts separately)
	OutputS3Bucket string
	
	// ElevenLabs API base URL
	ElevenLabsBaseURL string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		// Default to us-east-1 if not specified
		region = "us-east-1"
	}
	
	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		return nil, errors.New("DYNAMODB_TABLE_NAME environment variable is required")
	}
	
	secretName := os.Getenv("ELEVENLABS_SECRET_NAME")
	if secretName == "" {
		return nil, errors.New("ELEVENLABS_SECRET_NAME environment variable is required")
	}
	
	// Optional values with defaults
	outputBucket := os.Getenv("OUTPUT_S3_BUCKET")
	
	// Default API URL
	elevenLabsBaseURL := os.Getenv("ELEVENLABS_BASE_URL")
	if elevenLabsBaseURL == "" {
		elevenLabsBaseURL = "https://api.elevenlabs.io/v1"
	}
	
	return &Config{
		AWSRegion:           region,
		DynamoDBTableName:   tableName,
		ElevenLabsSecretName: secretName,
		OutputS3Bucket:      outputBucket,
		ElevenLabsBaseURL:   elevenLabsBaseURL,
	}, nil
}