#!/bin/bash

# Download dependencies and run Go benchmark

# Add timestamp to each line of output
add_timestamp() {
    while IFS= read -r line; do
        echo "[$(date '+%H:%M:%S')] $line"
    done
}

if [ ! -f "go.sum" ]; then
    echo "Downloading Go dependencies..."
    go mod download
    echo ""
fi

echo "Running benchmark..."
go run main.go 2>&1 | add_timestamp

