package model

import "time"

// TranscriptionStatus indicates the current status of a transcription job
type TranscriptionStatus string

const (
	// StatusPending indicates the transcription is queued but not processed
	StatusPending TranscriptionStatus = "PENDING"
	
	// StatusInProgress indicates the transcription is currently being processed
	StatusInProgress TranscriptionStatus = "IN_PROGRESS"
	
	// StatusCompleted indicates the transcription was successfully completed
	StatusCompleted TranscriptionStatus = "COMPLETED"
	
	// StatusFailed indicates the transcription encountered an error
	StatusFailed TranscriptionStatus = "FAILED"
)

// TranscriptionItem represents an item in the DynamoDB table
type TranscriptionItem struct {
	// FileIdentifier is the unique identifier (usually the S3 key)
	FileIdentifier string `json:"fileIdentifier" dynamodbav:"FileIdentifier"`
	
	// Status is the current status of the transcription
	Status TranscriptionStatus `json:"status" dynamodbav:"Status"`
	
	// SourceBucket is the S3 bucket containing the source audio file
	SourceBucket string `json:"sourceBucket" dynamodbav:"SourceBucket"`
	
	// SourceKey is the S3 key of the source audio file
	SourceKey string `json:"sourceKey" dynamodbav:"SourceKey"`
	
	// TranscriptText contains the transcribed text (if completed)
	TranscriptText string `json:"transcriptText,omitempty" dynamodbav:"TranscriptText,omitempty"`
	
	// OutputLocation contains the S3 URL to the transcript (if stored separately)
	OutputLocation string `json:"outputLocation,omitempty" dynamodbav:"OutputLocation,omitempty"`
	
	// ErrorMessage contains error details if the transcription failed
	ErrorMessage string `json:"errorMessage,omitempty" dynamodbav:"ErrorMessage,omitempty"`
	
	// CreatedAt is when the record was first created
	CreatedAt time.Time `json:"createdAt" dynamodbav:"CreatedAt"`
	
	// UpdatedAt is when the record was last updated
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"UpdatedAt"`
	
	// ProcessingTime is how long the transcription took in seconds
	ProcessingTime float64 `json:"processingTime,omitempty" dynamodbav:"ProcessingTime,omitempty"`
}

// ElevenLabsRequest represents a request to the ElevenLabs API
type ElevenLabsRequest struct {
	AudioURL string `json:"audio_url"`
}

// ElevenLabsResponse represents a response from the ElevenLabs API
type ElevenLabsResponse struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}