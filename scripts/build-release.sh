#!/bin/bash
set -e  # Exit immediately if a command exits with a non-zero status

checkDockerIsRunning(){
    # Check if docker is installed
    if ! docker --version &>/dev/null; then
        echo "Docker is not installed. Please install Docker and try again."
        exit 1
    fi

    # Check if docker is running
    echo "Checking if docker is running..."

    docker info > /dev/null 2>&1 &
    DOCKER_PID=$!

    TIMEOUT=5
    elapsed=0

    while kill -0 "$DOCKER_PID" 2>/dev/null && [ $elapsed -lt $TIMEOUT ]; do
        sleep 1
        elapsed=$((elapsed+1))
    done

    if kill -0 "$DOCKER_PID" 2>/dev/null; then
        kill "$DOCKER_PID" 2>/dev/null
        echo "Docker is not running or not responding. Please start Docker and try again."
        exit 1
    fi

    echo "Docker is running."
}

checkDockerIsRunning

# Check if the script is running on scripts folder
# if pwd includes scripts, then use the scripts folder as base path
if [[ "$(pwd)" == *"scripts"* ]]; then
    BASE_PATH=$(pwd)
else
    BASE_PATH=$(pwd)/scripts
fi

# Variables
TEMP_FOLDER="$BASE_PATH/../out/temp"
ASSETS_PATH="$BASE_PATH/../assets"
DOCKER_IMAGE_NAME="arm-go-builder:latest"
BINARY_NAME="sync.arm"
RELEASE_ZIP="release.zip"

echo Base path: $BASE_PATH
echo Temp folder: $TEMP_FOLDER
echo Docker image name: $DOCKER_IMAGE_NAME
echo Binary name: $BINARY_NAME
echo Release zip: $RELEASE_ZIP

# Function to clean up temporary resources
cleanup() {
    echo "Cleaning up temporary resources..."
    rm -rf "$TEMP_FOLDER"
}
trap cleanup EXIT  # Ensure cleanup runs when the script exits, even if it fails

# Create directory structure
echo "Creating directory structure..."
mkdir -p "$TEMP_FOLDER/notion_sync"
mkdir -p "$TEMP_FOLDER/notion_sync/certs"

# Copy necessary files
echo "Copying release files..."
cp "$ASSETS_PATH/release_files/notion_sync.conf" "$TEMP_FOLDER"
cp "$ASSETS_PATH/release_files/readme.txt" "$TEMP_FOLDER"
cp "$ASSETS_PATH/release_files/.env.example" "$TEMP_FOLDER/notion_sync/.env"
cp "$ASSETS_PATH/release_files/start.sh" "$TEMP_FOLDER/notion_sync/"

# Download certificates
curl -o cacert.pem https://curl.se/ca/cacert.pem

echo "Copying certificates..."
mv -f cacert.pem "$TEMP_FOLDER/notion_sync/certs/"

# Build Docker image if it does not exist
echo "Building Docker image..."
docker build -t $DOCKER_IMAGE_NAME .

# Compile the binary for ARM
echo "Compiling binary for ARM (This may take a while)"
docker run --rm -it -v $(pwd):/app $DOCKER_IMAGE_NAME bash -c "CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags '-extldflags -static' -o $BINARY_NAME"

# Move the compiled binary to the release directory
echo "Moving compiled binary to release directory..."
mv $BINARY_NAME "$TEMP_FOLDER/notion_sync/"

# Create a zip archive for the release
echo "Creating release zip file..."
cd $TEMP_FOLDER
zip -qq -r $RELEASE_ZIP .

mkdir -p $BASE_PATH/../out
mv $RELEASE_ZIP $BASE_PATH/../out

echo "Process completed successfully."
