#!/bin/bash

# Install dependencies and run JavaScript benchmark

# Add timestamp to each line of output
add_timestamp() {
    while IFS= read -r line; do
        echo "[$(date '+%H:%M:%S')] $line"
    done
}

if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
    echo ""
fi

echo "Running benchmark..."
node benchmark.js 2>&1 | add_timestamp

