# Quick Start Guide

Get started benchmarking SQLite across Rust, Go, Python, and JavaScript in minutes!

## 1. Prerequisites

Install the required tools for each language you want to test:

### Rust
```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

### Go
```bash
# macOS
brew install go

# Or download from: https://go.dev/dl/
```

### Python
```bash
# Python 3 is usually pre-installed on macOS/Linux
python3 --version  # Should be 3.8+
```

### Node.js
```bash
# macOS
brew install node

# Or download from: https://nodejs.org/
```

## 2. Run All Benchmarks

The easiest way to run all benchmarks:

```bash
./run_all.sh
```

This will automatically:
- Build and compile necessary dependencies
- Run each benchmark sequentially
- Display results for each language

## 3. Run Individual Benchmarks

If you want to test specific languages:

```bash
# Rust
cd rust && ./run.sh

# Go
cd go && ./run.sh

# Python
cd python && ./run.sh

# JavaScript
cd javascript && ./run.sh
```

Or run directly:

```bash
# Rust
cd rust && cargo run --release

# Go
cd go && go run main.go

# Python
cd python && python3 benchmark.py

# JavaScript
cd javascript && node benchmark.js
```

## 4. Understanding Results

Each benchmark displays:
- Individual operation times (milliseconds)
- Total execution time
- Record counts for queries

Example output:
```
=== Rust SQLite Benchmark ===

1. Batch Insert (10,000 records)... 45ms
2. Single Inserts (1,000 records)... 523ms
  → Found 7234 records
3. Simple Select (age > 30)... 12ms
  → Aggregated 10 groups
4. Complex Select (aggregation)... 8ms
5. Batch Update (5,000 records)... 34ms
6. Batch Delete (5,000 records)... 23ms

=== Results ===
Batch Insert:         45ms
Single Inserts:      523ms
Simple Select:        12ms
Complex Select:        8ms
Batch Update:         34ms
Batch Delete:         23ms
─────────────────────────
Total Time:          645ms
```

## 5. Comparing Results

Run the benchmark multiple times and compare:

```bash
# Run 3 times for average
./run_all.sh > results_1.txt
./run_all.sh > results_2.txt
./run_all.sh > results_3.txt
```

Key metrics to compare:
- **Batch Insert**: Tests bulk write performance
- **Single Inserts**: Shows transaction overhead
- **Total Time**: Overall efficiency

## 6. Typical Performance Rankings

Based on typical results (fastest to slowest):

1. 🥇 **Rust** - Lowest overhead, fastest execution
2. 🥈 **Go** - Close to Rust, excellent balance
3. 🥉 **JavaScript** - Surprisingly fast with better-sqlite3
4. 🏃 **Python** - Slower but simplest code

## 7. Troubleshooting

### Rust: "command not found: cargo"
```bash
source $HOME/.cargo/env
```

### Go: CGO compilation errors
```bash
# macOS: Install Xcode Command Line Tools
xcode-select --install
```

### JavaScript: Installation fails
```bash
cd javascript
rm -rf node_modules package-lock.json
npm install
```

### Python: Module errors
```bash
# Python's sqlite3 is built-in, no installation needed
python3 -m sqlite3 --version
```

## 8. Next Steps

- Read [BENCHMARKING_GUIDE.md](BENCHMARKING_GUIDE.md) for detailed methodology
- Modify test parameters in each language's source file
- Add custom benchmarks for your use case
- Enable SQLite optimizations (PRAGMA settings)

## 9. System Requirements

- **OS**: macOS, Linux, or Windows (WSL recommended)
- **RAM**: 1GB+ available
- **Disk**: 500MB for dependencies and build artifacts
- **CPU**: Any modern processor (results vary by hardware)

## 10. Need Help?

Common issues:
- **Permission denied**: Run `chmod +x *.sh`
- **Command not found**: Install missing language runtime
- **Compilation errors**: Check language version compatibility
- **Performance varies**: Normal - depends on system load

Happy benchmarking! 🚀

