#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

# Variables
BASE_PATH=$(pwd)
TEMP_FOLDER="$BASE_PATH/temp"
DOCKER_IMAGE_NAME="arm-go-builder:latest"
BINARY_NAME="sync.arm"
RELEASE_ZIP="release.zip"

# Function to clean up temporary resources
cleanup() {
    echo "Cleaning up temporary resources..."
    rm -rf "$TEMP_FOLDER"
}
trap cleanup EXIT  # Ensure cleanup runs when the script exits, even if it fails

# Create directory structure
echo "Creating directory structure..."
mkdir -p "$TEMP_FOLDER/notion_sync"

# Copy necessary files
echo "Copying release files..."
cp -r ./release_files/* "$TEMP_FOLDER/notion_sync/"
cp ./release_files/notion_sync.conf "$TEMP_FOLDER"
cp ./release_files/readme.txt "$TEMP_FOLDER"

# Download certificates
curl -o cacert.pem https://curl.se/ca/cacert.pem

echo "Copying certificates..."
mv -f cacert.pem "$TEMP_FOLDER/notion_sync/certs/"

# Build Docker image if it does not exist
if [[ "$(docker images -q $DOCKER_IMAGE_NAME 2> /dev/null)" == "" ]]; then
    echo "Building Docker image..."
    docker build -t "$DOCKER_IMAGE_NAME" ..
fi

# Compile the binary for ARM
echo "Compiling binary for ARM (This may take a while)"
cd ../
docker run --rm -it -v $(pwd):/app $DOCKER_IMAGE_NAME bash -c "CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags '-extldflags -static' -o $BINARY_NAME"

# Move the compiled binary to the release directory
echo "Moving compiled binary to release directory..."
mv $BINARY_NAME "$TEMP_FOLDER/notion_sync/"

# Create a zip archive for the release
echo "Creating release zip file..."
cd $TEMP_FOLDER
zip -r $RELEASE_ZIP .

mv $RELEASE_ZIP $BASE_PATH/../

echo "Process completed successfully."