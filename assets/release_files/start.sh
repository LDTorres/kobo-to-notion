#!/bin/sh
cd /mnt/onboard/.adds/notion_sync

# Change permissions
chmod +x sync.arm

echo "Starting sync at $(date)" >> sync.log
sync.arm >> sync.log 2>&1
echo "Finished sync with exit code $?" >> sync.log

exit 0