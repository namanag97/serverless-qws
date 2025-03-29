package awsclient

import (
	"context"
	
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// Clients holds all AWS service clients
type Clients struct {
	s3Client         *s3.Client
	dynamoDBClient   *dynamodb.Client
	secretsClient    *secretsmanager.Client
}

// NewClients initializes all AWS service clients
func NewClients(region string) (*Clients, error) {
	// Load AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), 
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}
	
	// Create service clients
	s3Client := s3.NewFromConfig(cfg)
	dynamoDBClient := dynamodb.NewFromConfig(cfg)
	secretsClient := secretsmanager.NewFromConfig(cfg)
	
	return &Clients{
		s3Client:         s3Client,
		dynamoDBClient:   dynamoDBClient,
		secretsClient:    secretsClient,
	}, nil
}

// GetS3 returns the S3 client
func (c *Clients) GetS3() *s3.Client {
	return c.s3Client
}

// GetDynamoDB returns the DynamoDB client
func (c *Clients) GetDynamoDB() *dynamodb.Client {
	return c.dynamoDBClient
}

// GetSecretsManager returns the Secrets Manager client
func (c *Clients) GetSecretsManager() *secretsmanager.Client {
	return c.secretsClient
}

// GetClients is a utility function to create clients directly
// Useful for testing and mock replacement
func GetClients(region string) (*s3.Client, *dynamodb.Client, *secretsmanager.Client) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), 
		config.WithRegion(region),
	)
	if err != nil {
		return nil, nil, nil
	}
	
	s3Client := s3.NewFromConfig(cfg)
	dynamoDBClient := dynamodb.NewFromConfig(cfg)
	secretsClient := secretsmanager.NewFromConfig(cfg)
	
	return s3Client, dynamoDBClient, secretsClient
}