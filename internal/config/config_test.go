package config

import (
	"os"
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables to restore after test
	originalTableName := os.Getenv("DYNAMODB_TABLE_NAME")
	originalSecretName := os.Getenv("ELEVENLABS_SECRET_NAME")
	originalRegion := os.Getenv("AWS_REGION")
	originalOutputBucket := os.Getenv("OUTPUT_S3_BUCKET")
	originalBaseURL := os.Getenv("ELEVENLABS_BASE_URL")
	
	// Restore environment variables after test
	defer func() {
		os.Setenv("DYNAMODB_TABLE_NAME", originalTableName)
		os.Setenv("ELEVENLABS_SECRET_NAME", originalSecretName)
		os.Setenv("AWS_REGION", originalRegion)
		os.Setenv("OUTPUT_S3_BUCKET", originalOutputBucket)
		os.Setenv("ELEVENLABS_BASE_URL", originalBaseURL)
	}()
	
	// Test case 1: Missing required environment variables
	os.Unsetenv("DYNAMODB_TABLE_NAME")
	os.Unsetenv("ELEVENLABS_SECRET_NAME")
	
	_, err := LoadConfig()
	assert.Error(t, err, "Expected error when required environment variables are missing")
	
	// Test case 2: All required environment variables provided
	os.Setenv("DYNAMODB_TABLE_NAME", "test-table")
	os.Setenv("ELEVENLABS_SECRET_NAME", "test-secret")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("OUTPUT_S3_BUCKET", "test-output-bucket")
	os.Setenv("ELEVENLABS_BASE_URL", "https://test-api.elevenlabs.io/v1")
	
	config, err := LoadConfig()
	assert.NoError(t, err, "Expected no error when all required environment variables are provided")
	assert.Equal(t, "test-table", config.DynamoDBTableName)
	assert.Equal(t, "test-secret", config.ElevenLabsSecretName)
	assert.Equal(t, "us-west-2", config.AWSRegion)
	assert.Equal(t, "test-output-bucket", config.OutputS3Bucket)
	assert.Equal(t, "https://test-api.elevenlabs.io/v1", config.ElevenLabsBaseURL)
	
	// Test case 3: Default values when optional variables not provided
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("ELEVENLABS_BASE_URL")
	os.Unsetenv("OUTPUT_S3_BUCKET")
	
	config, err = LoadConfig()
	assert.NoError(t, err, "Expected no error when only required environment variables are provided")
	assert.Equal(t, "us-east-1", config.AWSRegion, "Expected default AWS region")
	assert.Equal(t, "https://api.elevenlabs.io/v1", config.ElevenLabsBaseURL, "Expected default ElevenLabs API URL")
	assert.Equal(t, "", config.OutputS3Bucket, "Expected empty output bucket")
}