#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

echo "Downloading dependencies..."
go mod download 

echo "Tidying dependencies..."
go mod tidy

echo "Building the project..."
go build -v ./...

echo "Setup completed successfully."