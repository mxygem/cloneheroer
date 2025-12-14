#!/bin/bash

# Quick test script to copy a sample image to WATCH_DIR
# Usage: ./test-image.sh [image-name]

set -e

# Get WATCH_DIR from environment or use default
WATCH_DIR="${WATCH_DIR:-/tmp/clonehero-screenshots}"

# Get the project root (assuming we're in backend/)
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TESTDATA_DIR="$PROJECT_ROOT/testdata/screenshots"

# Check if WATCH_DIR is set
if [ -z "$WATCH_DIR" ]; then
    echo "Error: WATCH_DIR not set"
    echo "Set it with: export WATCH_DIR=/path/to/watch"
    exit 1
fi

# Create watch directory if it doesn't exist
mkdir -p "$WATCH_DIR"

# List available test images
echo "Available test images:"
ls -1 "$TESTDATA_DIR"/*.png 2>/dev/null | xargs -n1 basename || echo "No test images found"

echo ""
echo "Watch directory: $WATCH_DIR"
echo ""

# If image name provided, use it; otherwise use first available
if [ -n "$1" ]; then
    IMAGE_NAME="$1"
    SOURCE="$TESTDATA_DIR/$IMAGE_NAME"
else
    SOURCE=$(ls "$TESTDATA_DIR"/*.png 2>/dev/null | head -1)
    IMAGE_NAME=$(basename "$SOURCE")
fi

if [ ! -f "$SOURCE" ]; then
    echo "Error: Image not found: $SOURCE"
    exit 1
fi

# Copy with a unique name to trigger file watcher
TIMESTAMP=$(date +%s)
DEST="$WATCH_DIR/test-$TIMESTAMP-$(basename "$SOURCE")"

echo "Copying: $IMAGE_NAME"
echo "To: $DEST"
cp "$SOURCE" "$DEST"

echo ""
echo "✓ Image copied! The service should process it automatically."
echo "Check the service logs for processing status."


