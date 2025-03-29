#!/bin/bash
set -e

# Colors for better readability
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Starting Local Testing Script ===${NC}"

# Step 1: Check required tools
echo -e "\n${YELLOW}Checking required tools...${NC}"

# Check for AWS SAM CLI
if ! command -v sam &> /dev/null; then
    echo -e "${RED}AWS SAM CLI not found. Please install it first.${NC}"
    echo "Follow instructions at https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html"
    exit 1
fi

# Check for Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker not found. Please install it first.${NC}"
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    echo -e "${RED}Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check for Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go not found. Please install Go first.${NC}"
    exit 1
fi

echo -e "${GREEN}All required tools are available.${NC}"

# Step 2: Create necessary directories
echo -e "\n${YELLOW}Creating test directories and files...${NC}"

# Create events directory if it doesn't exist
mkdir -p events

# Create test S3 event if it doesn't exist
if [ ! -f "events/s3-event.json" ]; then
    cat > events/s3-event.json << 'EOF'
{
  "Records": [
    {
      "eventVersion": "2.1",
      "eventSource": "aws:s3",
      "awsRegion": "us-east-1",
      "eventTime": "2023-05-15T12:34:56.789Z",
      "eventName": "ObjectCreated:Put",
      "userIdentity": {
        "principalId": "EXAMPLE"
      },
      "requestParameters": {
        "sourceIPAddress": "127.0.0.1"
      },
      "responseElements": {
        "x-amz-request-id": "EXAMPLE123456789",
        "x-amz-id-2": "EXAMPLE123/abcdefghijklmno/123456789"
      },
      "s3": {
        "s3SchemaVersion": "1.0",
        "configurationId": "testConfigRule",
        "bucket": {
          "name": "local-test-bucket",
          "ownerIdentity": {
            "principalId": "EXAMPLE"
          },
          "arn": "arn:aws:s3:::local-test-bucket"
        },
        "object": {
          "key": "test-audio.mp3",
          "size": 1024,
          "eTag": "0123456789abcdef0123456789abcdef",
          "sequencer": "0A1B2C3D4E5F678901"
        }
      }
    }
  ]
}
EOF
    echo -e "${GREEN}Created events/s3-event.json${NC}"
else
    echo -e "${GREEN}events/s3-event.json already exists${NC}"
fi

# Create environment variables file
if [ ! -f "env.json" ]; then
    cat > env.json << 'EOF'
{
  "TranscriptionFunction": {
    "DYNAMODB_TABLE_NAME": "local-transcription-table",
    "OUTPUT_S3_BUCKET": "local-output-bucket",
    "ELEVENLABS_SECRET_NAME": "ElevenLabsApiKey",
    "AWS_REGION": "us-east-1",
    "AWS_SAM_LOCAL": "true"
  }
}
EOF
    echo -e "${GREEN}Created env.json${NC}"
else
    echo -e "${GREEN}env.json already exists${NC}"
fi

# Step 3: Run unit tests
echo -e "\n${YELLOW}Running unit tests...${NC}"
go test ./... -v || {
    echo -e "${RED}Unit tests failed.${NC}"
    # Continue anyway
}

# Step 4: Build the Lambda function
echo -e "\n${YELLOW}Building the Lambda function...${NC}"
cd deployments
sam build || {
    echo -e "${RED}Build failed.${NC}"
    exit 1
}
cd ..

# Step 5: Invoke the function locally
echo -e "\n${YELLOW}Invoking the function locally...${NC}"
cd deployments
sam local invoke TranscriptionFunction \
  --event ../events/s3-event.json \
  --env-vars ../env.json || {
    echo -e "${RED}Function invocation failed.${NC}"
    exit 1
}
cd ..

echo -e "\n${GREEN}Local testing completed successfully!${NC}"
echo -e "Function output is shown above."
echo -e "${YELLOW}NOTE:${NC} Since this is a local test, no actual file was transcribed."
echo -e "The mock ElevenLabs client returned a placeholder response."
echo -e "\nTo test with actual AWS services, you can deploy the function using:"
echo -e "${YELLOW}make deploy${NC}"