#!/bin/bash

# Install dependencies and run JavaScript benchmark

if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    npm install
    echo ""
fi

echo "Running benchmark..."
node benchmark.js

