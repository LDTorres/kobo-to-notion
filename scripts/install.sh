#!/bin/bash

# Check if it's running from vscode task
if [[ "$1" == "from-vscode" ]]; then
    echo "Running from vscode task."
	BASE_PATH=$(pwd)/out
elif [[ "$1" == "from-scripts" ]]; then
    echo "Running from scripts folder."
	BASE_PATH=$(pwd)/../out
else
    echo "Running from terminal."
	BASE_PATH=$(pwd)
fi

# Function to clean up temporary resources
cleanup() {
	# Check if the folder exists
	if [[ -z "$BASE_PATH/$FOLDER_NAME" ]] ; then
		exit 1
	fi

    rm -rf "$FOLDER_NAME"
}
trap cleanup EXIT  # Ensure cleanup runs when the script exits, even if it fails

# We're ultimately going to need unzip...
if ! unzip -v &>/dev/null ; then
	echo "This script relies on unzip! Please install unzip and try again."
	exit 1
fi

# Check if the release file exists
if [[ ! -f "$BASE_PATH/release.zip" ]]; then
    echo "release.zip file not found! Please left it on the same folder as this script."
    exit 1
fi

# Plugin folder name
FOLDER_NAME=$BASE_PATH/notion_sync

# Unzip the release file
unzip -qq "$BASE_PATH/release.zip" -d $FOLDER_NAME

# Are we on Linux or macOS?
PLATFORM="$(uname -s)"

# Find out where the Kobo is mounted...
KOBO_MOUNT_POINT="/dev/null"

case "${PLATFORM}" in
	"Linux" )
		# Use findmnt, it's in util-linux, which should be present in every sane distro.
		if ! findmnt -V &>/dev/null ; then
			echo "This script relies on findmnt, from util-linux!"
			exit 255
		fi

		# Match on the FS Label, which is common to all models.
		KOBO_MOUNT_POINT="$(findmnt -nlo TARGET LABEL=KOBOeReader)"
	;;
	"Darwin" )
		# Same idea, via diskutil
		KOBO_MOUNT_POINT="$(diskutil info -plist "KOBOeReader" | grep -A1 "MountPoint" | tail -n 1 | cut -d'>' -f2 | cut -d'<' -f1)"
	;;
	* )
		echo "Unsupported OS!"
		exit 255
	;;
esac

# Sanity check...
if [[ -z "${KOBO_MOUNT_POINT}" ]] ; then
	echo "Couldn't find a Kobo eReader volume! Is one actually mounted?"
	exit 1
fi

ADDS_DIR="${KOBO_MOUNT_POINT}/.adds"
if [[ ! -d "${ADDS_DIR}" ]] ; then
	echo "Can't find a .adds directory, ${KOBO_MOUNT_POINT} doesn't appear to point to a Kobo eReader... Is one actually mounted?"
	exit 1
fi

NM_DIR="${KOBO_MOUNT_POINT}/.adds/nm"
if [[ ! -d "${NM_DIR}" ]] ; then
	echo "Can't find a .adds/nm directory, ${KOBO_MOUNT_POINT} doesn't appear to point to a Kobo eReader. Is nickel menu installed?"
	exit 1
fi

# Ask for NOTION_TOKEN
read -p "Enter your NOTION_TOKEN: " NOTION_TOKEN

# Ask for NOTION_DATABASE_ID
read -p "Enter your NOTION_DATABASE_ID: " NOTION_DATABASE_ID

# Replace first line of .env file with NOTION_TOKEN
sed -i '' "1s/.*/NOTION_TOKEN=${NOTION_TOKEN}/" "$FOLDER_NAME/notion_sync/.env"

# Replace second line of .env file with NOTION_DATABASE_ID
sed -i '' "2s/.*/NOTION_DATABASE_ID=${NOTION_DATABASE_ID}/" "$FOLDER_NAME/notion_sync/.env"

# Copy the files to the Kobo
cp -r "$FOLDER_NAME/notion_sync" "$ADDS_DIR"
cp -r "$FOLDER_NAME/notion_sync.conf" "$NM_DIR"

echo "Installation complete!"
echo "Now secure disconnect your Kobo and check the Nickel menu for the new plugin called 'Notion Sync'."

# We're done!
exit 0
