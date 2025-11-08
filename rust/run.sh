#!/bin/bash

# Build and run Rust benchmark

# Add timestamp to each line of output
add_timestamp() {
    while IFS= read -r line; do
        echo "[$(date '+%H:%M:%S')] $line"
    done
}

echo "Building Rust project..."
cargo build --release

echo ""
echo "Running benchmark..."
cargo run --release 2>&1 | add_timestamp

