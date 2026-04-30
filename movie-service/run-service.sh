#!/bin/bash

# Movie Service Runner Script

echo "Starting Movie Service..."
cd "$(dirname "$0")"

# Set GOPROXY to avoid corporate proxy issues
export GOPROXY=https://proxy.golang.org,direct

# Run the service
go run main.go
