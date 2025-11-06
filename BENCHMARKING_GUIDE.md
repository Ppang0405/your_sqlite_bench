# SQLite Benchmarking Guide

## Understanding the Benchmarks

This benchmark suite tests SQLite performance across four programming languages using identical test scenarios. Here's what each test measures:

### Test Operations

1. **Batch Insert (10,000 records)**
   - Inserts 10,000 records within a single transaction
   - Tests write throughput with transaction batching
   - Key metric for bulk data import scenarios

2. **Single Inserts (1,000 records)**
   - Inserts 1,000 records individually (auto-commit mode)
   - Tests overhead of individual transactions
   - Demonstrates importance of batching

3. **Simple Select (WHERE clause)**
   - Query all users where age > 30
   - Tests basic filtering performance
   - Measures sequential scan efficiency

4. **Complex Select (aggregation)**
   - Uses GROUP BY, AVG(), COUNT() with range filter
   - Tests query optimizer and aggregation performance
   - Includes ORDER BY and LIMIT clauses

5. **Batch Update (5,000 records)**
   - Updates 5,000 records in a single transaction
   - Tests update performance with transaction batching

6. **Batch Delete (5,000 records)**
   - Deletes 5,000 records in a single transaction
   - Tests delete performance and cleanup

## Language-Specific Details

### Rust
- **Library**: `rusqlite` (v0.31)
- **Features**: Bundled SQLite library
- **Compilation**: Release mode with optimizations
- **Strengths**: Zero-cost abstractions, memory safety

### Go
- **Library**: `github.com/mattn/go-sqlite3`
- **Implementation**: CGO binding to SQLite C library
- **Strengths**: Efficient concurrency primitives, fast compilation

### Python
- **Library**: `sqlite3` (built-in)
- **Implementation**: C extension wrapping SQLite
- **Strengths**: Simple API, no external dependencies

### JavaScript (Node.js)
- **Library**: `better-sqlite3`
- **Implementation**: Native C++ binding
- **Strengths**: Synchronous API, efficient for single-threaded workloads

## Performance Factors

### What Affects Performance

1. **Language Runtime**
   - Compiled vs interpreted languages
   - Garbage collection overhead
   - Memory management efficiency

2. **SQLite Bindings**
   - Overhead of FFI (Foreign Function Interface)
   - Statement preparation and caching
   - Parameter binding efficiency

3. **Transaction Management**
   - Explicit vs implicit transactions
   - Transaction commit overhead
   - Write-ahead logging (WAL) mode

4. **I/O Patterns**
   - Filesystem caching
   - Disk sync operations
   - Buffer pool management

### Expected Performance Patterns

- **Batch operations** should be significantly faster than single operations
- **Compiled languages** (Rust, Go) typically show lower overhead
- **Query operations** are generally faster than write operations
- **Transaction batching** can provide 10-100x speedup for inserts

## Running Reproducible Benchmarks

### Best Practices

1. **Warm-up Runs**: Run benchmarks 2-3 times to warm up filesystem caches
2. **Clean State**: Each run creates a fresh database
3. **Single Core**: Run on a single CPU core to reduce variability
4. **Background Activity**: Minimize other system activity
5. **Multiple Runs**: Average results across 3-5 runs for statistical validity

### Environment Considerations

```bash
# Disable CPU frequency scaling (Linux)
sudo cpupower frequency-set --governor performance

# Clear filesystem caches (Linux)
sudo sh -c "echo 3 > /proc/sys/vm/drop_caches"

# Increase process priority (Unix-like)
nice -n -20 ./run_all.sh
```

## Interpreting Results

### Comparing Languages

When comparing results across languages, consider:

1. **Absolute Performance**: Raw millisecond measurements
2. **Relative Performance**: Ratio between languages for same operation
3. **Consistency**: Variance between runs (lower is better)
4. **Trade-offs**: Development speed vs execution speed

### Typical Results (Order of Magnitude)

Based on similar benchmarks, expected ordering (fastest to slowest):

1. **Rust / C++**: ~100-200ms total
2. **Go**: ~150-300ms total
3. **JavaScript (Node.js)**: ~200-400ms total
4. **Python**: ~500-1500ms total

*Note: Actual results vary significantly based on hardware and system load*

### Key Insights

- **Python's overhead** is most noticeable in loops (single inserts)
- **Compiled languages** show more consistent performance
- **better-sqlite3** (Node.js) is surprisingly competitive due to synchronous API
- **Transaction batching** matters more than language choice for bulk operations

## Optimization Tips

### General SQLite Optimizations

```sql
-- Enable Write-Ahead Logging
PRAGMA journal_mode = WAL;

-- Disable synchronous mode (risky!)
PRAGMA synchronous = OFF;

-- Increase cache size (in pages)
PRAGMA cache_size = -64000;  -- 64MB

-- Use memory for temp tables
PRAGMA temp_store = MEMORY;
```

### Language-Specific Optimizations

**Rust**
- Use `unchecked_transaction()` for performance
- Reuse prepared statements with caching
- Consider `unsafe` for zero-copy operations

**Go**
- Connection pooling with `SetMaxOpenConns(1)` for SQLite
- Use prepared statements with placeholder caching
- Enable `PRAGMA busy_timeout` for concurrent writes

**Python**
- Use `executemany()` for batch operations
- Consider `apsw` library for better performance
- Use context managers for transaction safety

**JavaScript**
- `better-sqlite3` is faster than `sqlite3` (async)
- Use `transaction()` helper for automatic rollback
- Prepare statements outside of loops

## Further Reading

- [SQLite Optimization Tips](https://www.sqlite.org/optoverview.html)
- [Database Performance Benchmarking](https://use-the-index-luke.com/)
- [Language Performance Comparisons](https://benchmarksgame-team.pages.debian.net/benchmarksgame/)

## Contributing

To add new benchmarks:

1. Maintain test parity across all languages
2. Document any language-specific limitations
3. Update this guide with new test descriptions
4. Follow existing code style for each language

