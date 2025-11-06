#!/bin/bash

# SQLite Benchmark Runner - Runs all language implementations

set -e

echo "=============================================="
echo "  SQLite Performance Benchmark Suite"
echo "=============================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Run Rust benchmark
run_rust() {
    echo -e "${BLUE}[1/4] Running Rust Benchmark...${NC}"
    echo ""
    if command_exists cargo; then
        cd rust
        if [ ! -d "target/release" ]; then
            echo "Building Rust project (first time)..."
            cargo build --release
            if [ $? -ne 0 ]; then
                echo -e "${RED}Failed to build Rust project${NC}"
                cd ..
                return 1
            fi
        fi
        cargo run --release --quiet
        cd ..
        echo ""
    else
        echo -e "${RED}Rust not found. Please install from https://rustup.rs/${NC}"
        echo ""
    fi
}

# Run Go benchmark
run_go() {
    echo -e "${BLUE}[2/4] Running Go Benchmark...${NC}"
    echo ""
    if command_exists go; then
        cd go
        if [ ! -f "go.sum" ]; then
            echo "Downloading Go dependencies..."
            go mod download
            if [ $? -ne 0 ]; then
                echo -e "${RED}Failed to download Go dependencies${NC}"
                cd ..
                return 1
            fi
        fi
        go run main.go
        cd ..
        echo ""
    else
        echo -e "${RED}Go not found. Please install from https://go.dev/dl/${NC}"
        echo ""
    fi
}

# Run Python benchmark
run_python() {
    echo -e "${BLUE}[3/4] Running Python Benchmark...${NC}"
    echo ""
    if command_exists python3; then
        cd python
        python3 benchmark.py
        cd ..
        echo ""
    elif command_exists python; then
        cd python
        python benchmark.py
        cd ..
        echo ""
    else
        echo -e "${RED}Python not found. Please install Python 3.8+${NC}"
        echo ""
    fi
}

# Run JavaScript benchmark
run_javascript() {
    echo -e "${BLUE}[4/4] Running JavaScript (Node.js) Benchmark...${NC}"
    echo ""
    if command_exists node; then
        cd javascript
        if [ ! -d "node_modules" ]; then
            echo "Installing Node.js dependencies..."
            npm install --quiet
            if [ $? -ne 0 ]; then
                echo -e "${RED}Failed to install Node.js dependencies${NC}"
                cd ..
                return 1
            fi
        fi
        node benchmark.js
        cd ..
        echo ""
    else
        echo -e "${RED}Node.js not found. Please install from https://nodejs.org/${NC}"
        echo ""
    fi
}

# Main execution
main() {
    # Store the original directory
    SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
    cd "$SCRIPT_DIR"
    
    run_rust
    run_go
    run_python
    run_javascript
    
    echo "=============================================="
    echo -e "${GREEN}  All benchmarks completed!${NC}"
    echo "=============================================="
    echo ""
    echo "Tip: You can run individual benchmarks by navigating"
    echo "to each language directory and running the respective"
    echo "command (see README.md for details)."
}

# Run the script
main

