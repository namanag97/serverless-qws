package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProcessor is a mock implementation of the processor interface
type MockProcessor struct {
	mock.Mock
}

// ProcessFile mocks the ProcessFile method
func (m *MockProcessor) ProcessFile(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}

func TestHandleS3Event(t *testing.T) {
	// Create mock processor
	mockProc := new(MockProcessor)
	
	// Create handler with mock processor
	handler := NewHandler(mockProc)
	
	// Test case 1: Single valid file
	mockProc.On("ProcessFile", mock.Anything, "test-bucket", "audio/test-file.aac").Return(nil)
	
	event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: "test-bucket",
					},
					Object: events.S3Object{
						Key:           "audio/test-file.aac",
						URLDecodedKey: "audio/test-file.aac",
					},
				},
			},
		},
	}
	
	err := handler.HandleS3Event(context.Background(), event)
	assert.NoError(t, err)
	mockProc.AssertExpectations(t)
	
	// Test case 2: Multiple files, one fails
	mockProc = new(MockProcessor)
	handler = NewHandler(mockProc)
	
	mockProc.On("ProcessFile", mock.Anything, "test-bucket", "audio/file1.aac").Return(nil)
	mockProc.On("ProcessFile", mock.Anything, "test-bucket", "audio/file2.aac").Return(errors.New("processing error"))
	mockProc.On("ProcessFile", mock.Anything, "test-bucket", "audio/file3.aac").Return(nil)
	
	event = events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-bucket"},
					Object: events.S3Object{Key: "audio/file1.aac", URLDecodedKey: "audio/file1.aac"},
				},
			},
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-bucket"},
					Object: events.S3Object{Key: "audio/file2.aac", URLDecodedKey: "audio/file2.aac"},
				},
			},
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-bucket"},
					Object: events.S3Object{Key: "audio/file3.aac", URLDecodedKey: "audio/file3.aac"},
				},
			},
		},
	}
	
	err = handler.HandleS3Event(context.Background(), event)
	assert.NoError(t, err, "Handler should continue processing even if one file fails")
	mockProc.AssertExpectations(t)
	
	// Test case 3: Unsupported file extension
	mockProc = new(MockProcessor)
	handler = NewHandler(mockProc)
	
	// No expectations for ProcessFile, as it should be skipped
	
	event = events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-bucket"},
					Object: events.S3Object{Key: "document.pdf", URLDecodedKey: "document.pdf"},
				},
			},
		},
	}
	
	err = handler.HandleS3Event(context.Background(), event)
	assert.NoError(t, err, "Handler should skip unsupported file types")
	mockProc.AssertExpectations(t)
}

func TestIsValidAudioFile(t *testing.T) {
	handler := NewHandler(nil)
	
	// Test valid extensions
	assert.True(t, handler.isValidAudioFile("test.aac"))
	assert.True(t, handler.isValidAudioFile("path/to/file.mp3"))
	assert.True(t, handler.isValidAudioFile("UPPERCASE.WAV"))
	assert.True(t, handler.isValidAudioFile("with spaces.flac"))
	
	// Test invalid extensions
	assert.False(t, handler.isValidAudioFile("document.pdf"))
	assert.False(t, handler.isValidAudioFile("image.jpg"))
	assert.False(t, handler.isValidAudioFile("noextension"))
	assert.False(t, handler.isValidAudioFile(".htaccess"))
}