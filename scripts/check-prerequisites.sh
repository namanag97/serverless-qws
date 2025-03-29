#!/bin/bash

echo "Checking prerequisites..."

# Check AWS SAM CLI
if command -v sam &> /dev/null; then
    echo "✅ AWS SAM CLI is installed"
    sam --version
else
    echo "❌ AWS SAM CLI is not installed"
    echo "Please install it from: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html"
    exit 1
fi

# Check Docker
if command -v docker &> /dev/null; then
    echo "✅ Docker is installed"
    docker --version
    if docker info &> /dev/null; then
        echo "✅ Docker daemon is running"
    else
        echo "❌ Docker daemon is not running"
        echo "Please start Docker and try again"
        exit 1
    fi
else
    echo "❌ Docker is not installed"
    echo "Please install it from: https://docs.docker.com/get-docker/"
    exit 1
fi

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go is installed"
    go version
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [ "$(printf '%s\n' "1.18" "$GO_VERSION" | sort -V | head -n1)" = "1.18" ]; then
        echo "✅ Go version is 1.18 or higher"
    else
        echo "❌ Go version is below 1.18"
        echo "Please install Go 1.18 or higher from: https://golang.org/dl/"
        exit 1
    fi
else
    echo "❌ Go is not installed"
    echo "Please install it from: https://golang.org/dl/"
    exit 1
fi

# Check AWS credentials
if aws sts get-caller-identity &> /dev/null; then
    echo "✅ AWS credentials are configured"
    echo "Using AWS account: $(aws sts get-caller-identity --query Account --output text)"
else
    echo "❌ AWS credentials are not configured"
    echo "Please configure AWS credentials using: aws configure"
    exit 1
fi

echo "✅ All prerequisites are met!" 