#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

echo "Running tests..."

go test -v ./...

echo "Tests completed successfully."