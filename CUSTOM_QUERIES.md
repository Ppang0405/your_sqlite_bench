# Custom Queries Benchmark Results

## Overview

This benchmark suite tests real-world query patterns simulating a DVD catalog application with complex joins, filtering, and pagination. The queries represent a realistic user journey: browsing an index page, clicking on a DVD, viewing its details, relationships, and similar recommendations.

## Test Configuration

- **Database**: `r18_25_11_04.sqlite` (production database with real data)
- **Iterations**: 10 per language
- **Queries per iteration**: 4 queries (40 total queries per language)
- **Connection mode**: Read-only
- **Statement preparation**: All statements prepared once before loop execution

## The Four Queries

### Query 1: Index Page (DVD Listing with Filters)

**Purpose**: Fetch paginated DVD listings with complex filtering and search capabilities.

**Features**:
- Multiple LEFT OUTER JOINs (actress, category tables)
- Full-text search simulation using LIKE across multiple columns
- Date filtering (release_date <= random date)
- Pagination with LIMIT/OFFSET
- ORDER BY release_date DESC
- DISTINCT to handle join duplicates

**Parameters**:
- Random release date: 2020-2025
- Fixed LIMIT: 100 records
- Random page: 0-49 (OFFSET = page * 100)

**Average Results**: ~100 rows per query

```sql
SELECT DISTINCT derived_video.dvd_id, derived_video.jacket_full_url, derived_video.release_date 
FROM derived_video 
LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
WHERE derived_video.dvd_id IS NOT NULL 
  AND derived_video.dvd_id IS NOT '' 
  AND derived_video.release_date IS NOT NULL 
  AND derived_video.release_date <= ? 
  AND derived_video.jacket_full_url IS NOT NULL 
  AND (lower(derived_video.dvd_id) LIKE lower('%%') 
       OR lower(derived_actress.name_romaji) LIKE lower('%%') 
       OR lower(derived_actress.name_kanji) LIKE lower('%%') 
       OR lower(derived_actress.name_kana) LIKE lower('%%') 
       OR lower(derived_category.name_en) LIKE lower('%%') 
       OR lower(derived_category.name_ja) LIKE lower('%%')) 
ORDER BY derived_video.release_date DESC
LIMIT ? OFFSET ?
```

---

### Query 2: DVD Detail Page

**Purpose**: Fetch complete information for a specific DVD.

**Features**:
- Single table query (derived_video)
- Retrieves 20 columns of detailed information
- Uses dvd_id from Query 1 results (simulates user clicking on a DVD)

**Average Results**: 1 row per query

```sql
SELECT derived_video.content_id, derived_video.dvd_id, derived_video.title_en, derived_video.title_ja, 
       derived_video.comment_en, derived_video.comment_ja, derived_video.runtime_mins, derived_video.release_date, 
       derived_video.sample_url, derived_video.maker_id, derived_video.label_id, derived_video.series_id, 
       derived_video.jacket_full_url, derived_video.jacket_thumb_url, derived_video.gallery_full_first, 
       derived_video.gallery_full_last, derived_video.gallery_thumb_first, derived_video.gallery_thumb_last, 
       derived_video.site_id, derived_video.service_code 
FROM derived_video 
WHERE derived_video.dvd_id IS NOT NULL 
  AND derived_video.dvd_id != '' 
  AND derived_video.release_date IS NOT NULL 
  AND derived_video.dvd_id = ?
```

---

### Query 3: DVD Relationships (Categories & Actresses)

**Purpose**: Fetch all categories and actresses associated with a DVD.

**Features**:
- Multiple LEFT OUTER JOINs (video_category, category, video_actress, actress)
- Returns denormalized data (multiple rows per DVD)
- Uses same dvd_id from Query 1

**Average Results**: 8-15 rows per query (varies by DVD)

```sql
SELECT derived_video.content_id, derived_category.id AS cat_id, derived_category.name_en AS cat_name_en, 
       derived_category.name_ja AS cat_name_ja, derived_actress.id AS act_id, derived_actress.name_romaji, 
       derived_actress.name_kana, derived_actress.name_kanji, derived_actress.image_url AS act_image_url 
FROM derived_video 
LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
WHERE derived_video.dvd_id = ?
```

---

### Query 4: Similar DVDs (Same Year)

**Purpose**: Find DVDs from the same release year for recommendations.

**Features**:
- Subquery to get release date of current DVD
- CAST and STRFTIME for year extraction
- Random ordering for variety
- Fixed LIMIT of 6 recommendations
- Uses same dvd_id from Query 1

**Average Results**: 6 rows per query

```sql
SELECT derived_video.dvd_id, derived_video.jacket_full_url, derived_video.release_date 
FROM derived_video, (SELECT derived_video.release_date AS release_date 
                     FROM derived_video 
                     WHERE derived_video.dvd_id = ?) AS anon_1 
WHERE CAST(STRFTIME('%Y', derived_video.release_date) AS INTEGER) = CAST(STRFTIME('%Y', anon_1.release_date) AS INTEGER) 
  AND derived_video.dvd_id IS NOT NULL 
  AND derived_video.dvd_id != '' 
  AND derived_video.release_date IS NOT NULL 
  AND derived_video.jacket_full_url IS NOT NULL 
ORDER BY random() 
LIMIT 6 OFFSET 0
```

---

## Performance Results

### Benchmark Results (4 queries × 10 iterations)

| Language | Total Time | Avg per Iteration | Relative Speed |
|----------|-----------|-------------------|----------------|
| **JavaScript (Node.js)** | **7,434 ms** | **743 ms** | **1.00x** (fastest) |
| **Rust** | 10,252 ms | 1,025 ms | 1.38x |
| **Go** | 10,261 ms | 1,026 ms | 1.38x |
| **Python** | 11,946 ms | 1,195 ms | 1.61x |

### Query-Level Breakdown

| Query | JavaScript | Rust | Go | Python |
|-------|-----------|------|-----|--------|
| Query 1 (Index) | 100 rows | 100 rows | 100 rows | 100 rows |
| Query 2 (Detail) | 1 row | 1 row | 1 row | 1 row |
| Query 3 (Relations) | 8 rows | 15 rows | 15 rows | 12 rows |
| Query 4 (Similar) | 6 rows | 6 rows | 6 rows | 6 rows |

*Note: Query 3 row count varies because different DVDs have different numbers of categories/actresses.*

---

## Why JavaScript is Fastest

### 1. **better-sqlite3 Library Advantages**

JavaScript uses the `better-sqlite3` library, which is a highly optimized C++ binding:
- **Synchronous I/O**: No async/await overhead
- **Direct memory access**: Minimal data copying
- **Optimized statement caching**: Built into the library
- **Native object mapping**: Direct conversion to JavaScript objects

### 2. **Efficient API Design**

```javascript
const rows = stmt.all(params);  // Single call, all rows
```

vs. other languages requiring manual iteration:

```go
// Go - manual iteration
for rows.Next() {
    rows.Scan(&col1, &col2, ...)
}
```

```rust
// Rust - manual iteration
while let Some(row) = rows.next()? {
    // process row
}
```

### 3. **Statement Preparation**

All implementations now prepare statements once before the loop:

```javascript
// JavaScript
const stmt = db.prepare(query);
for (let i = 0; i < iterations; i++) {
    const rows = stmt.all(params);
}
```

```go
// Go - NOW OPTIMIZED
stmt, err := db.Prepare(query)
defer stmt.Close()
for i := 0; i < iterations; i++ {
    rows, err := stmt.Query(params)
}
```

```rust
// Rust
let mut stmt = conn.prepare(query)?;
for i in 0..iterations {
    let mut rows = stmt.query(params)?;
}
```

---

## User Journey Simulation

The benchmark simulates a realistic user flow:

1. **User browses index page** → Query 1 executes with random date/pagination
2. **User sees 100 DVDs** → Query 1 returns 100 results
3. **User clicks on a random DVD** → Pick random `dvd_id` from Query 1 results
4. **User views DVD detail page** → Query 2 executes for that `dvd_id`
5. **User sees categories/actresses** → Query 3 executes for that `dvd_id`
6. **User sees recommendations** → Query 4 executes for that `dvd_id`

This flow ensures queries 2, 3, and 4 always use valid data from Query 1, creating a **connected query workflow** that mirrors real application behavior.

---

## Optimization Details

### Go Optimizations Applied

**Before**: Used `db.Query()` which prepares statement on every call
```go
rows, err := db.Query(indexQuery, params...)  // Slow!
```

**After**: Prepare once, execute many times
```go
stmt, err := db.Prepare(indexQuery)
defer stmt.Close()
for i := 0; i < iterations; i++ {
    rows, err := stmt.Query(params...)  // Fast!
}
```

**Result**: Reduced Go time from ~12s to ~10s

### Python Limitations

Python's built-in `sqlite3` module has limited statement caching:
- No explicit `prepare()` method
- Relies on internal caching (last query)
- Cannot match the performance of C++ bindings

### Rust Performance

Rust's `rusqlite` crate:
- Excellent type safety
- Proper statement preparation
- Close to JavaScript performance
- Slight overhead from manual iteration

---

## Key Takeaways

1. **Statement preparation is critical**: Preparing statements once before loops provides significant speedup
2. **JavaScript/Node.js is fastest**: `better-sqlite3` is exceptionally well-optimized
3. **All languages are within 60% of fastest**: 7-12 seconds for 40 complex queries is excellent
4. **Real-world patterns matter**: Testing with realistic query chains reveals true performance
5. **SQLite is fast**: Even complex joins with 100+ row results execute in ~1 second per iteration

---

## Running the Benchmark

Run all languages:
```bash
./run_all.sh --custom-queries
```

Run individual language:
```bash
# JavaScript
cd javascript && npm install && node benchmark.js --custom-queries

# Go
cd go && go run main.go --custom-queries

# Rust
cd rust && cargo run --release -- --custom-queries

# Python
cd python && python3 benchmark.py --custom-queries
```

---

## Database Schema Notes

The queries operate on these tables:
- `derived_video`: Main DVD table (content_id, dvd_id, titles, dates, URLs, etc.)
- `derived_actress`: Actress information (id, names in multiple languages)
- `derived_category`: Category/genre information (id, names)
- `derived_video_actress`: Junction table (content_id, actress_id)
- `derived_video_category`: Junction table (content_id, category_id)

**Recommended indexes** (see `optimize_indexes.sql`):
- Primary keys on all id columns
- Foreign keys on junction tables
- Composite index on (release_date, dvd_id)
- Index on dvd_id for WHERE clauses

---

*Last updated: 2025-11-08*

