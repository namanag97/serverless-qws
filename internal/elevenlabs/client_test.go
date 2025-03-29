package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/transcription-service/internal/model"
)

// Mock Secrets Manager client
type MockSecretsManagerClient struct {
	mock.Mock
}

func (m *MockSecretsManagerClient) GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)
}

func TestTranscribeAudio(t *testing.T) {
	// Set up mock secrets manager
	mockSecretsClient := new(MockSecretsManagerClient)
	apiKey := "test-api-key"
	secretName := "test-secret"
	
	mockSecretsClient.On("GetSecretValue", mock.Anything, &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}).Return(&secretsmanager.GetSecretValueOutput{
		SecretString: &apiKey,
	}, nil)
	
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/transcribe", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, apiKey, r.Header.Get("xi-api-key"))
		
		// Read and verify request body
		var req model.ElevenLabsRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/audio.aac", req.AudioURL)
		
		// Write response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model.ElevenLabsResponse{
			ID:      "test-id",
			Text:    "Hello, world!",
			Success: true,
		})
	}))
	defer server.Close()
	
	// Create client and override baseURL to point to test server
	client, err := NewClient(context.Background(), mockSecretsClient, secretName)
	assert.NoError(t, err)
	client.baseURL = server.URL + "/v1"
	
	// Test the TranscribeAudio method
	resp, err := client.TranscribeAudio(context.Background(), "https://example.com/audio.aac")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "Hello, world!", resp.Text)
	assert.True(t, resp.Success)
	
	mockSecretsClient.AssertExpectations(t)
}

func TestTranscribeAudio_Error(t *testing.T) {
	// Set up mock secrets manager
	mockSecretsClient := new(MockSecretsManagerClient)
	apiKey := "test-api-key"
	secretName := "test-secret"
	
	mockSecretsClient.On("GetSecretValue", mock.Anything, &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}).Return(&secretsmanager.GetSecretValueOutput{
		SecretString: &apiKey,
	}, nil)
	
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(model.ElevenLabsResponse{
			Error:   "Invalid audio format",
			Success: false,
		})
	}))
	defer server.Close()
	
	// Create client and override baseURL
	client, err := NewClient(context.Background(), mockSecretsClient, secretName)
	assert.NoError(t, err)
	client.baseURL = server.URL + "/v1"
	
	// Test the TranscribeAudio method with error response
	_, err = client.TranscribeAudio(context.Background(), "https://example.com/invalid.aac")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid audio format")
	
	mockSecretsClient.AssertExpectations(t)
}