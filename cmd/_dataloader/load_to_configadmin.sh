#!/bin/bash

# Load US Cities ZipCode data into ConfigAdmin stores directory
# This makes the zipcode database available as a configuration store in ConfigAdmin

set -e

echo "=== BlueConfig - Load ZipCodes to ConfigAdmin ==="
echo ""

# Check if USCities.json exists
if [ ! -f "../configadmin/USCities.json" ]; then
    echo "Error: USCities.json not found at ../configadmin/USCities.json"
    echo "Please ensure the file exists before running this script"
    exit 1
fi

# Check if configadmin stores directory exists
if [ ! -d "../configadmin/stores" ]; then
    echo "Creating stores directory..."
    mkdir -p ../configadmin/stores
fi

# Build the loader
echo "Building zipcode loader for ConfigAdmin..."
go build -o load_to_configadmin load_to_configadmin.go

# Run the loader
echo ""
echo "Loading zipcode data into ConfigAdmin stores..."
echo ""
./load_to_configadmin

# Cleanup
rm -f load_to_configadmin

echo ""
echo "=== Setup Complete ==="
echo ""
echo "The 'US Zip Codes' store is now available in ConfigAdmin!"
echo ""
echo "To access it:"
echo "  1. cd ../configadmin"
echo "  2. ./start.sh"
echo "  3. Open your browser and select 'US Zip Codes' from the store selector"
echo ""
