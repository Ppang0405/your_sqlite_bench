#!/bin/bash

# Run Python benchmark

# Add timestamp to each line of output
add_timestamp() {
    while IFS= read -r line; do
        echo "[$(date '+%H:%M:%S')] $line"
    done
}

echo "Running benchmark..."
if command -v python3 >/dev/null 2>&1; then
    python3 benchmark.py 2>&1 | add_timestamp
elif command -v python >/dev/null 2>&1; then
    python benchmark.py 2>&1 | add_timestamp
else
    echo "Python not found!"
    exit 1
fi

