#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

echo "Building the project..."
go build -v -o notion_sync

echo "Build completed successfully."

