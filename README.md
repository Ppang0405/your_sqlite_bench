# SQLite Performance Benchmark

Cross-language SQLite performance comparison for **Rust**, **Go**, **Python**, **JavaScript** (Node.js), and **TypeScript** (Bun).

## Quick Start

```bash
# Run all benchmarks
./run_all.sh
```

## Benchmark Tests

Each implementation performs identical operations:

1. **Batch Insert** - 10,000 records in a transaction
2. **Single Inserts** - 1,000 individual inserts
3. **Simple Select** - Query with WHERE clause
4. **Complex Select** - Aggregation with GROUP BY
5. **Batch Update** - 5,000 records update
6. **Batch Delete** - 5,000 records deletion

## Results Summary

| Language       | Total Time | Winner In            |
|----------------|-----------|----------------------|
| **Go**         | 634ms     | Overall & Inserts    |
| **Rust**       | 755ms     | SELECT queries       |
| **Python**     | 763ms     | Batch operations     |
| **JavaScript** | 869ms     | Updates              |

👉 **See [RESULTS.md](RESULTS.md) for detailed analysis and recommendations**

## Prerequisites

Install at least one:

- **Rust**: https://rustup.rs/
- **Go**: https://go.dev/dl/
- **Python**: 3.8+ (usually pre-installed)
- **Node.js**: https://nodejs.org/
- **Bun**: https://bun.sh (for TypeScript version)

## Run Individual Benchmarks

```bash
# Rust
cd rust && cargo run --release

# Go
cd go && go run main.go

# Python
cd python && python3 benchmark.py

# JavaScript (Node.js)
cd javascript && node benchmark.js

# TypeScript (Bun)
cd typescript && bun benchmark.ts
```

## Documentation

- **[RESULTS.md](RESULTS.md)** - Detailed benchmark results and analysis
- **[QUICK_START.md](QUICK_START.md)** - Installation and setup guide
- **[BENCHMARKING_GUIDE.md](BENCHMARKING_GUIDE.md)** - Methodology and optimization tips

## Libraries Used

- **Rust**: `rusqlite` (bundled SQLite)
- **Go**: `mattn/go-sqlite3` (CGO binding)
- **Python**: `sqlite3` (built-in)
- **JavaScript**: `better-sqlite3` (native binding)
- **TypeScript**: `bun:sqlite` (Bun's built-in SQLite module)

## Key Findings

✨ **Go** is the fastest overall (634ms)  
✨ **Rust** has sub-millisecond SELECT queries  
✨ **Python** performs surprisingly well (only 20% slower)  
✨ **Transaction batching** is 50-60× faster than single operations  

## References

Based on: https://github.com/antonputra/tutorials/tree/223/lessons/223

