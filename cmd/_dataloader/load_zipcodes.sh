#!/bin/bash

# Load US Cities ZipCode data into BlueConfig database
# This script builds and runs the zipcode loader

set -e

echo "=== BlueConfig ZipCode Loader ==="
echo ""

# Check if USCities.json exists
if [ ! -f "../configadmin/USCities.json" ]; then
    echo "Error: USCities.json not found at ../configadmin/USCities.json"
    echo "Please ensure the file exists before running this script"
    exit 1
fi

# Build the loader
echo "Building zipcode loader..."
go build -o load_zipcodes load_zipcodes.go

# Run the loader
echo "Running zipcode loader..."
./load_zipcodes

# Show database info
echo ""
echo "Database created at: ./stores/zipcodes.db"
echo ""
echo "You can now use this database for testing and benchmarking."
echo ""
echo "To run benchmarks:"
echo "  cd ../.."
echo "  go test -bench=BenchmarkLoadUSCities -benchmem"
echo ""
