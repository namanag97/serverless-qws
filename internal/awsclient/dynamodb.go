package awsclient

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/yourusername/transcription-service/internal/model"
)

// DynamoDBOperations provides operations for working with DynamoDB
type DynamoDBOperations struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBOperations creates a new DynamoDBOperations instance
func NewDynamoDBOperations(client *dynamodb.Client, tableName string) *DynamoDBOperations {
	return &DynamoDBOperations{
		client:    client,
		tableName: tableName,
	}
}

// CreateTranscriptionItem creates a new transcription item in DynamoDB
func (d *DynamoDBOperations) CreateTranscriptionItem(ctx context.Context, item *model.TranscriptionItem) error {
	// Set timestamps
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	
	// Marshal item to DynamoDB attribute values
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}
	
	// Put item in DynamoDB
	_, err = d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      av,
	})
	
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}
	
	log.Printf("Created DynamoDB item for file: %s", item.FileIdentifier)
	return nil
}

// UpdateTranscriptionItemStatus updates the status of a transcription item
func (d *DynamoDBOperations) UpdateTranscriptionItemStatus(
	ctx context.Context, 
	fileIdentifier string, 
	status model.TranscriptionStatus,
	transcriptText string,
	outputLocation string,
	errorMessage string,
	processingTime float64,
) error {
	// Build update expression
	updateExpression := "SET #status = :status, #updatedAt = :updatedAt"
	expressionAttributeNames := map[string]string{
		"#status":    "Status",
		"#updatedAt": "UpdatedAt",
	}
	expressionAttributeValues := map[string]types.AttributeValue{
		":status":    &types.AttributeValueMemberS{Value: string(status)},
		":updatedAt": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
	}
	
	// Add optional attributes if provided
	if transcriptText != "" {
		updateExpression += ", #transcriptText = :transcriptText"
		expressionAttributeNames["#transcriptText"] = "TranscriptText"
		expressionAttributeValues[":transcriptText"] = &types.AttributeValueMemberS{Value: transcriptText}
	}
	
	if outputLocation != "" {
		updateExpression += ", #outputLocation = :outputLocation"
		expressionAttributeNames["#outputLocation"] = "OutputLocation"
		expressionAttributeValues[":outputLocation"] = &types.AttributeValueMemberS{Value: outputLocation}
	}
	
	if errorMessage != "" {
		updateExpression += ", #errorMessage = :errorMessage"
		expressionAttributeNames["#errorMessage"] = "ErrorMessage"
		expressionAttributeValues[":errorMessage"] = &types.AttributeValueMemberS{Value: errorMessage}
	}
	
	if processingTime > 0 {
		updateExpression += ", #processingTime = :processingTime"
		expressionAttributeNames["#processingTime"] = "ProcessingTime"
		expressionAttributeValues[":processingTime"] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%f", processingTime)}
	}
	
	// Update item in DynamoDB
	_, err := d.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"FileIdentifier": &types.AttributeValueMemberS{Value: fileIdentifier},
		},
		UpdateExpression:          aws.String(updateExpression),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	})
	
	if err != nil {
		return fmt.Errorf("failed to update item in DynamoDB: %w", err)
	}
	
	log.Printf("Updated DynamoDB item status to %s for file: %s", status, fileIdentifier)
	return nil
}

// GetTranscriptionItem gets a transcription item by fileIdentifier
func (d *DynamoDBOperations) GetTranscriptionItem(ctx context.Context, fileIdentifier string) (*model.TranscriptionItem, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"FileIdentifier": &types.AttributeValueMemberS{Value: fileIdentifier},
		},
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}
	
	// Check if item was found
	if result.Item == nil {
		return nil, nil
	}
	
	// Unmarshal item
	var item model.TranscriptionItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}
	
	return &item, nil
}