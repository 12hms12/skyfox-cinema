#!/bin/bash

# Script to run the backend locally
# This script handles the GOPROXY issue and provides instructions for PostgreSQL

echo "=========================================="
echo "Backend Server Startup Script"
echo "=========================================="
echo ""

# Check if server binary exists
if [ ! -f "out/server" ]; then
    echo "Building the server..."
    GOPROXY=https://proxy.golang.org,direct go build -o out/server ./main.go
    if [ $? -ne 0 ]; then
        echo "Build failed! Please check the errors above."
        exit 1
    fi
    echo "Build successful!"
    echo ""
fi

# Check if local PostgreSQL is running
LOCAL_PG_RUNNING=$(ps aux | grep -E "(postgres|Postgres)" | grep -v grep | wc -l)

if [ $LOCAL_PG_RUNNING -gt 0 ]; then
    echo "⚠️  WARNING: Local PostgreSQL is running on port 5432"
    echo ""
    echo "The backend needs to connect to the Docker PostgreSQL (postgresdb)."
    echo "You have two options:"
    echo ""
    echo "Option 1: Stop local PostgreSQL temporarily"
    echo "  - Stop Postgres.app from the menu bar"
    echo "  - Or run: brew services stop postgresql@14"
    echo ""
    echo "Option 2: Run the backend in Docker (recommended)"
    echo "  - Use docker-compose to run the full stack"
    echo ""
    read -p "Do you want to continue anyway? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Exiting. Please stop local PostgreSQL and try again."
        exit 1
    fi
fi

# Check if Docker PostgreSQL is running
DOCKER_PG_RUNNING=$(docker ps --filter "name=postgresdb" --format "{{.Names}}" | wc -l)

if [ $DOCKER_PG_RUNNING -eq 0 ]; then
    echo "❌ ERROR: Docker PostgreSQL container 'postgresdb' is not running!"
    echo "Please start it first."
    exit 1
fi

echo "✅ Docker PostgreSQL is running"
echo ""
echo "Starting backend server on port 8080..."
echo "API Documentation: http://localhost:8080/swagger/index.html"
echo ""
echo "Press Ctrl+C to stop the server"
echo "=========================================="
echo ""

# Run the server
./out/server -configFile=./config/config-local.yml
