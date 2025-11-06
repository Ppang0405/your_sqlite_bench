#!/bin/bash

# Download dependencies and run Go benchmark

if [ ! -f "go.sum" ]; then
    echo "Downloading Go dependencies..."
    go mod download
    echo ""
fi

echo "Running benchmark..."
go run main.go

