#!/bin/bash

# Secra Restore Script
# Usage: ./restore.sh <backup_file.tar.gz>

set -e

CONTAINER_NAME="secra-web"
BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "❌ Error: Please specify a backup file to restore."
    echo "Usage: $0 <backup_file>"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ Error: Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

if [ ! "$(docker ps -q -f name=${CONTAINER_NAME})" ]; then
    echo "❌ Error: Container ${CONTAINER_NAME} is not running."
    exit 1
fi

FILENAME=$(basename "$BACKUP_FILE")
TMP_PATH="/tmp/${FILENAME}"

echo "🚚 Transferring backup to container..."
docker cp "$BACKUP_FILE" "${CONTAINER_NAME}:${TMP_PATH}"

echo "📥 Starting restoration and migration..."
docker compose exec web secra backup restore "${TMP_PATH}"

echo "✅ Restoration successful!"
