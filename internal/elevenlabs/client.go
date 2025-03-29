package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/yourusername/transcription-service/internal/model"
)

// Client provides methods for interacting with the ElevenLabs API
type Client struct {
	httpClient  *http.Client
	baseURL     string
	apiKey      string
}

// defaultHTTPClient returns a properly configured HTTP client
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

// TranscribeAudio sends an audio file URL to the ElevenLabs API for transcription
// Note: This implementation will be overridden by client_local.go when running in local mode
func (c *Client) TranscribeAudio(ctx context.Context, audioURL string) (*model.ElevenLabsResponse, error) {
	return c.sendTranscriptionRequest(ctx, audioURL)
}

// sendTranscriptionRequest is the actual implementation that makes the HTTP request
func (c *Client) sendTranscriptionRequest(ctx context.Context, audioURL string) (*model.ElevenLabsResponse, error) {
	// Construct the API endpoint
	endpoint := fmt.Sprintf("%s/transcribe", c.baseURL)
	
	// Create request body
	requestBody := model.ElevenLabsRequest{
		AudioURL: audioURL,
	}
	
	// Marshal request to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", c.apiKey)
	
	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ElevenLabs: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ElevenLabs API returned non-200 status code: %d, body: %s", 
			resp.StatusCode, string(respBody))
	}
	
	// Parse response
	var response model.ElevenLabsResponse
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	// Check for API-level errors
	if !response.Success {
		return nil, fmt.Errorf("ElevenLabs API returned error: %s", response.Error)
	}
	
	return &response, nil
}