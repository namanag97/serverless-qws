# Lambda Function

A Go-based AWS Lambda function for audio transcription and processing.

## Prerequisites

Before you begin, ensure you have the following installed and configured:

- [AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
- [Docker](https://docs.docker.com/get-docker/) (required for local Lambda testing)
- [Go](https://golang.org/dl/) (v1.18 or later)
- AWS credentials configured in your environment

## AWS Credentials Setup

1. Install and configure the AWS CLI:
```bash
aws configure
```

2. Enter your AWS credentials when prompted:
- AWS Access Key ID
- AWS Secret Access Key
- Default region (e.g., us-east-1)
- Default output format (json)

## Local Development Setup

1. Clone the repository:
```bash
git clone https://github.com/namanag97/my-lambda-function.git
cd my-lambda-function
```

2. Install dependencies:
```bash
go mod download
```

3. Build the function:
```bash
make build
```

4. Run tests:
```bash
make test
```

5. Run locally:
```bash
make local
```

## Project Structure

```
.
├── cmd
│   └── transcriber           # Main binary
│       └── main.go           # Entry point for Lambda
├── internal                  # Private application code
│   ├── handler               # Lambda handler logic
│   ├── processor             # Business logic
│   ├── awsclient            # AWS clients wrapper
│   ├── elevenlabs           # ElevenLabs API client
│   ├── config               # Configuration loading
│   └── model                # Shared data structures
└── deployments              # AWS SAM templates
```

## Deployment

1. Build the function:
```bash
make build
```

2. Deploy to AWS:
```bash
make deploy
```

## Testing

Run all tests:
```bash
make test
```

Run specific test:
```bash
go test ./internal/...
```

## Local Testing with SAM

1. Start local API:
```bash
sam local start-api
```

2. Test the endpoint:
```bash
curl http://localhost:3000/hello
```

## Environment Variables

Required environment variables:
- `AWS_REGION`: AWS region (e.g., us-east-1)
- `ELEVENLABS_API_KEY`: API key for ElevenLabs service

## License

MIT License
