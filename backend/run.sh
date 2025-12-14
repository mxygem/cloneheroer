#!/bin/bash

# Clone Hero Score Tracker - Run Script
# This script helps set up and run the service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Clone Hero Score Tracker${NC}"
echo "================================"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}Warning: .env file not found${NC}"
    echo "Creating .env from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo -e "${GREEN}.env file created. Please edit it with your configuration.${NC}"
    else
        echo -e "${RED}Error: .env.example not found${NC}"
        exit 1
    fi
    echo ""
    echo "Please edit .env file and set:"
    echo "  - WATCH_DIR: Directory containing screenshot images"
    echo "  - DATABASE_URL: PostgreSQL connection string"
    echo ""
    read -p "Press Enter after editing .env file..."
fi

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Check required environment variables
if [ -z "$WATCH_DIR" ]; then
    echo -e "${RED}Error: WATCH_DIR not set${NC}"
    exit 1
fi

if [ -z "$DATABASE_URL" ]; then
    echo -e "${RED}Error: DATABASE_URL not set${NC}"
    exit 1
fi

# Check if watch directory exists
if [ ! -d "$WATCH_DIR" ]; then
    echo -e "${YELLOW}Warning: Watch directory does not exist: $WATCH_DIR${NC}"
    read -p "Create it? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        mkdir -p "$WATCH_DIR"
        echo -e "${GREEN}Created watch directory${NC}"
    else
        exit 1
    fi
fi

# Check if Tesseract is installed
if ! command -v tesseract &> /dev/null; then
    echo -e "${RED}Error: Tesseract OCR not found${NC}"
    echo "Please install Tesseract:"
    echo "  Ubuntu/Debian: sudo apt-get install tesseract-ocr"
    echo "  macOS: brew install tesseract"
    exit 1
fi

echo -e "${GREEN}Configuration looks good!${NC}"
echo ""
echo "Starting service..."
echo "  Watch Directory: $WATCH_DIR"
echo "  Database: ${DATABASE_URL%%\?*}"  # Hide password
echo "  Port: ${PORT:-3000}"
echo ""

# Run the service
go run cmd/server/main.go


