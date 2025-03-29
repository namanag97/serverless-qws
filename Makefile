.PHONY: build clean deploy test validate update

# Variables
STACK_NAME ?= transcription-lambda
AWS_REGION ?= us-east-1
DYNAMODB_TABLE ?= TranscriptionState
INPUT_BUCKET ?= transcription-input-$(shell date +%Y%m%d%H%M%S)
OUTPUT_BUCKET ?= transcription-output-$(shell date +%Y%m%d%H%M%S)
SECRET_NAME ?= ElevenLabsApiKey

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build the Lambda function"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run unit tests"
	@echo "  validate    - Validate SAM template"
	@echo "  deploy      - Deploy the Lambda function to AWS"
	@echo "  update      - Update an existing deployment"
	@echo ""
	@echo "Configuration:"
	@echo "  STACK_NAME     = $(STACK_NAME)"
	@echo "  AWS_REGION     = $(AWS_REGION)"
	@echo "  DYNAMODB_TABLE = $(DYNAMODB_TABLE)"
	@echo "  INPUT_BUCKET   = $(INPUT_BUCKET)"
	@echo "  OUTPUT_BUCKET  = $(OUTPUT_BUCKET)"
	@echo "  SECRET_NAME    = $(SECRET_NAME)"

# Build the Lambda function
build:
	@echo "Building Lambda function..."
	cd deployments && sam build

# Clean up build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf .aws-sam

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Validate the SAM template
validate:
	@echo "Validating SAM template..."
	cd deployments && sam validate

# Deploy the Lambda function to AWS
deploy: build
	@echo "Deploying Lambda function to AWS..."
	cd deployments && sam deploy \
		--stack-name $(STACK_NAME) \
		--parameter-overrides \
		DynamoDBTableName=$(DYNAMODB_TABLE) \
		InputBucketName=$(INPUT_BUCKET) \
		OutputBucketName=$(OUTPUT_BUCKET) \
		ElevenLabsSecretName=$(SECRET_NAME) \
		--capabilities CAPABILITY_IAM \
		--region $(AWS_REGION) \
		--guided

# Update an existing deployment
update: build
	@echo "Updating Lambda function deployment..."
	cd deployments && sam deploy \
		--stack-name $(STACK_NAME) \
		--parameter-overrides \
		DynamoDBTableName=$(DYNAMODB_TABLE) \
		InputBucketName=$(INPUT_BUCKET) \
		OutputBucketName=$(OUTPUT_BUCKET) \
		ElevenLabsSecretName=$(SECRET_NAME) \
		--capabilities CAPABILITY_IAM \
		--region $(AWS_REGION) \
		--no-confirm-changeset