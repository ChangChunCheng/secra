#!/bin/bash

# Secra Backup Script
# Usage: ./backup.sh <output_directory>

set -e

CONTAINER_NAME="secra-web"
OUT_DIR=$1

if [ -z "$OUT_DIR" ]; then
    echo "❌ Error: Please specify an output directory."
    echo "Usage: $0 <output_dir>"
    exit 1
fi

if [ ! "$(docker ps -q -f name=${CONTAINER_NAME})" ]; then
    echo "❌ Error: Container ${CONTAINER_NAME} is not running."
    exit 1
fi

# Fetch version dynamically from the CLI
VERSION=$(docker compose exec web secra version | tr -d '\r')
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

mkdir -p "$OUT_DIR"
FILENAME="secra_v${VERSION}_${TIMESTAMP}.tar.gz"
TMP_PATH="/tmp/${FILENAME}"
FINAL_PATH="${OUT_DIR}/${FILENAME}"

echo "📦 Initializing backup for version v${VERSION}..."
docker compose exec web secra backup create -o "${TMP_PATH}"

echo "🚚 Copying backup to host: ${FINAL_PATH}"
docker cp "${CONTAINER_NAME}:${TMP_PATH}" "${FINAL_PATH}"

echo "✅ Backup completed: ${FINAL_PATH}"
echo "💡 Note: Temporary file in container /tmp will persist until container restart."
