#!/bin/bash

# Script to run all tests and exit with non-zero status if any test fails
# Useful for CI environments

# Exit immediately if a command exits with a non-zero status
set -e

echo "Running all tests in CI environment..."

# Clean test cache to ensure fresh test runs
go clean -testcache

# Create output directory if needed
mkdir -p out/

# Run tests with coverage
echo "Running tests with coverage..."
# Store the exit code of the test command
go test $(go list ./...) -v -coverprofile=out/coverage.out
TEST_EXIT_CODE=$?

# Generate coverage report regardless of test success/failure
if [ -f out/coverage.out ]; then
  go tool cover -html=out/coverage.out -o out/coverage.html
  echo "Coverage report generated at out/coverage.html"
fi

# If tests failed, exit with the same code
if [ $TEST_EXIT_CODE -ne 0 ]; then
  echo "❌ Tests failed with exit code $TEST_EXIT_CODE"
  exit $TEST_EXIT_CODE
fi

echo "✅ All tests passed successfully!"
exit 0