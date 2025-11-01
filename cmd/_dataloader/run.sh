#!/bin/bash

# BlueConfig Data Loader - Run Script
# This script builds and runs the data loader to populate the BenchTest store

set -e

echo "========================================"
echo "BlueConfig Data Loader - BenchTest"
echo "========================================"
echo ""

# Navigate to the dataloader directory
cd "$(dirname "$0")"

# Build the data loader
echo "Building data loader..."
go build -o dataloader main.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"
echo ""

# Run the data loader
echo "Running data loader..."
echo ""
./dataloader

echo ""
echo "Moving benchtest.db to configadmin stores directory..."
mv ./stores/benchtest.db ../configadmin/stores/

echo ""
echo "========================================"
echo "Data loader finished!"
echo "========================================"
echo ""
echo "The BenchTest store is ready at: ../configadmin/stores/benchtest.db"
echo ""
echo "To use it:"
echo "1. Start configadmin: cd ../configadmin && ./start.sh"
echo "2. Open http://localhost:8214/landing"
echo "3. Select 'BenchTest' store"
echo ""
