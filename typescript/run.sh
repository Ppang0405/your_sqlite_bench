#!/bin/bash

# TypeScript (Bun) SQLite Benchmark Runner

set -e

# Change to script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# Add timestamp to each line of output
add_timestamp() {
    while IFS= read -r line; do
        echo "[$(date '+%H:%M:%S')] $line"
    done
}

echo -e "${BLUE}TypeScript (Bun) SQLite Benchmark${NC}"
echo ""

# Check if Bun is installed
if ! command -v bun &> /dev/null; then
    echo -e "${RED}Error: Bun is not installed${NC}"
    echo "Please install Bun from: https://bun.sh"
    echo "Run: curl -fsSL https://bun.sh/install | bash"
    exit 1
fi

# Install dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing dependencies..."
    bun install 2>&1 | add_timestamp
    echo ""
fi

# Run benchmark
if [ "$1" = "--custom-queries" ]; then
    bun benchmark.ts --custom-queries 2>&1 | add_timestamp
else
    bun benchmark.ts 2>&1 | add_timestamp
fi

