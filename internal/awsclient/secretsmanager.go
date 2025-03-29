package awsclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// SecretsManagerOperations provides operations for working with AWS Secrets Manager
type SecretsManagerOperations struct {
	client *secretsmanager.Client
}

// NewSecretsManagerOperations creates a new SecretsManagerOperations instance
func NewSecretsManagerOperations(client *secretsmanager.Client) *SecretsManagerOperations {
	return &SecretsManagerOperations{
		client: client,
	}
}

// GetSecretString retrieves a plain string secret
func (s *SecretsManagerOperations) GetSecretString(ctx context.Context, secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	
	result, err := s.client.GetSecretValue(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get secret from Secrets Manager: %w", err)
	}
	
	if result.SecretString == nil {
		return "", fmt.Errorf("secret %s not found or has no string value", secretName)
	}
	
	return *result.SecretString, nil
}

// GetSecretJSON retrieves a JSON secret and unmarshals it into the provided target
func (s *SecretsManagerOperations) GetSecretJSON(ctx context.Context, secretName string, target interface{}) error {
	secretString, err := s.GetSecretString(ctx, secretName)
	if err != nil {
		return err
	}
	
	err = json.Unmarshal([]byte(secretString), target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal secret JSON: %w", err)
	}
	
	return nil
}