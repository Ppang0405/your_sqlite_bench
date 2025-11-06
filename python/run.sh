#!/bin/bash

# Run Python benchmark

echo "Running benchmark..."
if command -v python3 >/dev/null 2>&1; then
    python3 benchmark.py
elif command -v python >/dev/null 2>&1; then
    python benchmark.py
else
    echo "Python not found!"
    exit 1
fi

