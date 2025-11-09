# TypeScript (Bun) SQLite Benchmark

This benchmark uses [Bun](https://bun.sh) runtime with its built-in `bun:sqlite` module for high-performance SQLite operations.

## Prerequisites

**Bun runtime** (v1.0.0 or higher)

Install Bun:
```bash
curl -fsSL https://bun.sh/install | bash
```

Or on macOS via Homebrew:
```bash
brew install oven-sh/bun/bun
```

## Why Bun?

- **Built-in SQLite**: No need for external native modules like `better-sqlite3`
- **Fast**: Bun's SQLite implementation is written in Zig and highly optimized
- **TypeScript Native**: First-class TypeScript support without transpilation
- **Minimal Dependencies**: Uses Bun's built-in APIs
- **Modern Syntax**: Tagged template literals for SQL queries

## Installation

```bash
cd typescript
bun install
```

## Running the Benchmark

Run all benchmarks:
```bash
bun benchmark.ts
```

Run only custom queries benchmark:
```bash
bun benchmark.ts --custom-queries
```

Or use the convenience script:
```bash
./run.sh                    # All benchmarks
./run.sh --custom-queries   # Custom queries only
```

## Performance

Bun's SQLite performance using the `SQL` API with tagged template literals:

**Custom Queries Benchmark** (4 queries × 10 iterations):
- **Bun/TypeScript (SQL API)**: ~7.5 seconds ⚡ **FASTEST!**
- **Go**: ~11.3 seconds
- **Node.js (better-sqlite3)**: ~11.4 seconds
- **Python**: ~12.4 seconds
- **Rust**: ~12.5 seconds

**🎉 Bun is 34% faster than Node.js!** The `SQL` API from `"bun"` with tagged template literals provides exceptional performance. Even though it uses async/await syntax, Bun optimizes SQLite queries to run synchronously under the hood with zero overhead.

*Performance may vary based on system configuration and SQLite optimizations.*

## API Comparison

### Bun (`SQL` from `"bun"`)
```typescript
import { SQL } from "bun";

const db = new SQL("sqlite://database.db");
const rows = await db`SELECT * FROM users WHERE age > ${30}`;
await db.close();
```

**Features:**
- Connection string format: `sqlite://path.db`
- Tagged template literals for queries
- Automatic parameterization with `${}`
- Async/await API
- Works synchronously under the hood (optimized for SQLite)

### Node.js (`better-sqlite3`)
```javascript
import Database from 'better-sqlite3';

const db = new Database('database.db');
const stmt = db.prepare('SELECT * FROM users WHERE age > ?');
const rows = stmt.all(30);
db.close();
```

**Bun's advantage:** Tagged template literals provide cleaner syntax and automatic SQL injection protection!

## Features Tested

1. **Batch Insert** - Transaction-based bulk inserts (10,000 records)
2. **Single Inserts** - Individual insert operations (1,000 records)
3. **Simple Select** - Basic SELECT with WHERE clause
4. **Complex Select** - Aggregation with GROUP BY and ORDER BY
5. **Batch Update** - Transaction-based bulk updates (5,000 records)
6. **Batch Delete** - Transaction-based bulk deletes (5,000 records)
7. **Custom Queries** - Real-world complex queries with JOINs and pagination

## Custom Queries

The custom queries benchmark simulates a DVD catalog application:

1. **Index Page Query** - Complex JOIN with filtering and pagination
2. **DVD Detail Query** - Fetch complete DVD information
3. **Relationships Query** - Get categories and actresses for a DVD
4. **Similar DVDs Query** - Find DVDs from the same year

See `CUSTOM_QUERIES.md` in the root directory for detailed information.

## Notes

- Bun's `SQL` API is primarily designed for PostgreSQL/MySQL but supports SQLite
- Works synchronously under the hood (optimized for SQLite)
- No need to install or compile native modules
- TypeScript types are included out of the box
- Tagged template literals provide clean, modern syntax
- Automatic SQL injection protection

## Troubleshooting

### Error: "Bun not found"
Make sure Bun is installed and in your PATH:
```bash
bun --version
```

### Error: "Cannot find module"
Install dependencies:
```bash
bun install
```

## Resources

- [Bun Documentation](https://bun.sh/docs)
- [Bun SQLite API](https://bun.sh/docs/api/sqlite)
- [Bun GitHub](https://github.com/oven-sh/bun)

