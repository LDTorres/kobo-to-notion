#!/bin/bash

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

# Remove the files
rm -rf "$ADDS_DIR/notion_sync"
rm -rf "$NM_DIR/notion_sync.conf"

echo "Uninstall complete!"
echo "Now secure disconnect your Kobo and check the Nickel menu."

# We're done!
exit 0
