#!/bin/bash

# Quick script to check if the service is running and test the API

set -e

PORT="${PORT:-3000}"
API_URL="http://localhost:$PORT"

echo "Checking Clone Hero Score Tracker service..."
echo "=============================================="
echo ""

# Check if service is running
echo "1. Health check:"
if curl -s -f "$API_URL/health" > /dev/null; then
    echo "   ✓ Service is running"
    curl -s "$API_URL/health" | python3 -m json.tool 2>/dev/null || curl -s "$API_URL/health"
else
    echo "   ✗ Service is not responding"
    echo "   Make sure the service is running on port $PORT"
    exit 1
fi

echo ""
echo "2. Recent scores:"
SCORES=$(curl -s "$API_URL/scores?limit=5")
if [ -n "$SCORES" ]; then
    echo "$SCORES" | python3 -m json.tool 2>/dev/null || echo "$SCORES"
    COUNT=$(echo "$SCORES" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data) if isinstance(data, list) else 0)" 2>/dev/null || echo "?")
    echo ""
    echo "   Found $COUNT score(s)"
else
    echo "   No scores found"
fi

echo ""
echo "3. To add a test image, run:"
echo "   ./test-image.sh"


