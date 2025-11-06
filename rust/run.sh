#!/bin/bash

# Build and run Rust benchmark

echo "Building Rust project..."
cargo build --release

echo ""
echo "Running benchmark..."
cargo run --release

