package processor

import (
	"context"
	"errors"
	"testing"
	"time"
	
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/transcription-service/internal/elevenlabs"
	"github.com/yourusername/transcription-service/internal/model"
)

// Mock S3 operations
type MockS3Operations struct {
	mock.Mock
}

func (m *MockS3Operations) DownloadFile(ctx context.Context, bucket, key string) (string, error) {
	args := m.Called(ctx, bucket, key)
	return args.String(0), args.Error(1)
}

func (m *MockS3Operations) GeneratePresignedURL(ctx context.Context, bucket, key string, expirationSeconds int) (string, error) {
	args := m.Called(ctx, bucket, key, expirationSeconds)
	return args.String(0), args.Error(1)
}

func (m *MockS3Operations) UploadText(ctx context.Context, bucket, key, content string) error {
	args := m.Called(ctx, bucket, key, content)
	return args.Error(0)
}

// Mock DynamoDB operations
type MockDynamoDBOperations struct {
	mock.Mock
}

func (m *MockDynamoDBOperations) CreateTranscriptionItem(ctx context.Context, item *model.TranscriptionItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockDynamoDBOperations) UpdateTranscriptionItemStatus(
	ctx context.Context,
	fileIdentifier string,
	status model.TranscriptionStatus,
	transcriptText string,
	outputLocation string,
	errorMessage string,
	processingTime float64,
) error {
	args := m.Called(ctx, fileIdentifier, status, transcriptText, outputLocation, errorMessage, processingTime)
	return args.Error(0)
}

func (m *MockDynamoDBOperations) GetTranscriptionItem(ctx context.Context, fileIdentifier string) (*model.TranscriptionItem, error) {
	args := m.Called(ctx, fileIdentifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TranscriptionItem), args.Error(1)
}

// Mock ElevenLabs client
type MockElevenLabsClient struct {
	mock.Mock
}

func (m *MockElevenLabsClient) TranscribeAudio(ctx context.Context, audioURL string) (*model.ElevenLabsResponse, error) {
	args := m.Called(ctx, audioURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ElevenLabsResponse), args.Error(1)
}

// Test the ProcessFile method
func TestProcessFile(t *testing.T) {
	// Create mocks
	mockS3Client := &s3.Client{}
	mockDynamoDBClient := &dynamodb.Client{}
	mockS3Ops := new(MockS3Operations)
	mockDynamoDBOps := new(MockDynamoDBOperations)
	mockElevenLabsClient := new(MockElevenLabsClient)
	
	// Create processor with mocks
	processor := &Processor{
		s3Client:           mockS3Client,
		dynamoDBClient:     mockDynamoDBClient,
		elevenlabsClient:   mockElevenLabsClient,
		s3Operations:       mockS3Ops,
		dynamoDBOperations: mockDynamoDBOps,
		tableName:          "test-table",
		outputBucket:       "test-output-bucket",
	}
	
	// Test case: New file, successful transcription
	ctx := context.Background()
	bucket := "test-bucket"
	key := "audio/test-file.aac"
	
	// Setup mock expectations
	mockDynamoDBOps.On("GetTranscriptionItem", ctx, key).Return(nil, nil)
	mockDynamoDBOps.On("CreateTranscriptionItem", ctx, mock.MatchedBy(func(item *model.TranscriptionItem) bool {
		return item.FileIdentifier == key && 
			   item.Status == model.StatusInProgress &&
			   item.SourceBucket == bucket &&
			   item.SourceKey == key
	})).Return(nil)
	
	mockS3Ops.On("GeneratePresignedURL", ctx, bucket, key, 3600).Return("https://presigned-url", nil)
	
	mockElevenLabsClient.On("TranscribeAudio", ctx, "https://presigned-url").Return(&model.ElevenLabsResponse{
		ID:      "test-id",
		Text:    "This is a test transcription.",
		Success: true,
	}, nil)
	
	mockS3Ops.On("UploadText", ctx, "test-output-bucket", "transcripts/test-file.txt", "This is a test transcription.").Return(nil)
	
	mockDynamoDBOps.On("UpdateTranscriptionItemStatus", 
		ctx, 
		key, 
		model.StatusCompleted, 
		"This is a test transcription.",
		"s3://test-output-bucket/transcripts/test-file.txt",
		"",
		mock.MatchedBy(func(pt float64) bool { return pt > 0 }),
	).Return(nil)
	
	// Call method
	err := processor.ProcessFile(ctx, bucket, key)
	
	// Verify
	assert.NoError(t, err)
	mockDynamoDBOps.AssertExpectations(t)
	mockS3Ops.AssertExpectations(t)
	mockElevenLabsClient.AssertExpectations(t)
}

// Test file already processed
func TestProcessFile_AlreadyProcessed(t *testing.T) {
	// Create mocks
	mockS3Client := &s3.Client{}
	mockDynamoDBClient := &dynamodb.Client{}
	mockS3Ops := new(MockS3Operations)
	mockDynamoDBOps := new(MockDynamoDBOperations)
	mockElevenLabsClient := new(MockElevenLabsClient)
	
	// Create processor with mocks
	processor := &Processor{
		s3Client:           mockS3Client,
		dynamoDBClient:     mockDynamoDBClient,
		elevenlabsClient:   mockElevenLabsClient,
		s3Operations:       mockS3Ops,
		dynamoDBOperations: mockDynamoDBOps,
		tableName:          "test-table",
		outputBucket:       "test-output-bucket",
	}
	
	// Test case: File already processed
	ctx := context.Background()
	bucket := "test-bucket"
	key := "audio/test-file.aac"
	
	// Create a completed item
	existingItem := &model.TranscriptionItem{
		FileIdentifier: key,
		Status:         model.StatusCompleted,
		SourceBucket:   bucket,
		SourceKey:      key,
		TranscriptText: "Existing transcription",
		CreatedAt:      time.Now().Add(-1 * time.Hour),
	}
	
	// Setup mock expectations
	mockDynamoDBOps.On("GetTranscriptionItem", ctx, key).Return(existingItem, nil)
	
	// Call method
	err := processor.ProcessFile(ctx, bucket, key)
	
	// Verify
	assert.NoError(t, err)
	mockDynamoDBOps.AssertExpectations(t)
	mockS3Ops.AssertNotCalled(t, "GeneratePresignedURL")
	mockElevenLabsClient.AssertNotCalled(t, "TranscribeAudio")
}

// Test transcription API error
func TestProcessFile_TranscriptionError(t *testing.T) {
	// Create mocks
	mockS3Client := &s3.Client{}
	mockDynamoDBClient := &dynamodb.Client{}
	mockS3Ops := new(MockS3Operations)
	mockDynamoDBOps := new(MockDynamoDBOperations)
	mockElevenLabsClient := new(MockElevenLabsClient)
	
	// Create processor with mocks
	processor := &Processor{
		s3Client:           mockS3Client,
		dynamoDBClient:     mockDynamoDBClient,
		elevenlabsClient:   mockElevenLabsClient,
		s3Operations:       mockS3Ops,
		dynamoDBOperations: mockDynamoDBOps,
		tableName:          "test-table",
		outputBucket:       "test-output-bucket",
	}
	
	// Test case: Transcription API error
	ctx := context.Background()
	bucket := "test-bucket"
	key := "audio/test-file.aac"
	
	// Setup mock expectations
	mockDynamoDBOps.On("GetTranscriptionItem", ctx, key).Return(nil, nil)
	mockDynamoDBOps.On("CreateTranscriptionItem", ctx, mock.Anything).Return(nil)
	
	mockS3Ops.On("GeneratePresignedURL", ctx, bucket, key, 3600).Return("https://presigned-url", nil)
	
	apiError := errors.New("API error")
	mockElevenLabsClient.On("TranscribeAudio", ctx, "https://presigned-url").Return(nil, apiError)
	
	mockDynamoDBOps.On("UpdateTranscriptionItemStatus", 
		ctx, 
		key, 
		model.StatusFailed, 
		"",
		"",
		"Transcription API error: API error",
		float64(0),
	).Return(nil)
	
	// Call method
	err := processor.ProcessFile(ctx, bucket, key)
	
	// Verify
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transcription API error")
	mockDynamoDBOps.AssertExpectations(t)
	mockS3Ops.AssertExpectations(t)
	mockElevenLabsClient.AssertExpectations(t)
}