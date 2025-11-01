#!/bin/bash

# BlueConfig Admin - Start Script
# This script builds and starts the BlueConfig Admin server

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "=================================="
echo "BlueConfig Admin - Start Script"
echo "=================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Check if npm is installed
if ! command -v npm &> /dev/null; then
    echo "Error: npm is not installed"
    exit 1
fi

# Build frontend if needed
if [ ! -f "public/bundle.js" ]; then
    echo "Building frontend..."
    npm run build
    echo ""
fi

# Build backend
echo "Building backend..."
go build -o configadmin
echo ""

# Start server
echo "Starting BlueConfig Admin Server..."
echo "Server will be available at: http://localhost:8213"
echo ""
echo "Press Ctrl+C to stop"
echo "=================================="
echo ""

./configadmin
