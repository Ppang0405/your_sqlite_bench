use rusqlite::{Connection, Result, params};
use std::time::Instant;
use std::fs;

/// Creates and initializes the database with the users table
fn setup_database(conn: &Connection) -> Result<()> {
    conn.execute(
        "CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            age INTEGER NOT NULL
        )",
        [],
    )?;
    Ok(())
}

/// Performs batch insert of records within a transaction
fn benchmark_batch_insert(conn: &Connection, count: usize) -> Result<u128> {
    let start = Instant::now();
    
    let tx = conn.unchecked_transaction()?;
    
    for i in 0..count {
        tx.execute(
            "INSERT INTO users (name, email, age) VALUES (?1, ?2, ?3)",
            params![
                format!("User{}", i),
                format!("user{}@example.com", i),
                20 + (i % 50) as i32
            ],
        )?;
    }
    
    tx.commit()?;
    
    Ok(start.elapsed().as_millis())
}

/// Performs single inserts without explicit transaction
fn benchmark_single_inserts(conn: &Connection, count: usize) -> Result<u128> {
    let start = Instant::now();
    
    for i in 0..count {
        conn.execute(
            "INSERT INTO users (name, email, age) VALUES (?1, ?2, ?3)",
            params![
                format!("SingleUser{}", i),
                format!("single{}@example.com", i),
                25 + (i % 40) as i32
            ],
        )?;
    }
    
    Ok(start.elapsed().as_millis())
}

/// Performs simple SELECT query with WHERE clause
fn benchmark_simple_select(conn: &Connection) -> Result<u128> {
    let start = Instant::now();
    
    let mut stmt = conn.prepare("SELECT * FROM users WHERE age > ?1")?;
    let mut rows = stmt.query([30])?;
    
    let mut count = 0;
    while rows.next()?.is_some() {
        count += 1;
    }
    
    let duration = start.elapsed().as_millis();
    println!("  → Found {} records", count);
    
    Ok(duration)
}

/// Performs complex SELECT query with aggregation
fn benchmark_complex_select(conn: &Connection) -> Result<u128> {
    let start = Instant::now();
    
    let mut stmt = conn.prepare(
        "SELECT age, COUNT(*) as count, AVG(age) as avg_age 
         FROM users 
         WHERE age BETWEEN ?1 AND ?2 
         GROUP BY age 
         ORDER BY count DESC 
         LIMIT 10"
    )?;
    
    let mut rows = stmt.query([25, 50])?;
    
    let mut count = 0;
    while rows.next()?.is_some() {
        count += 1;
    }
    
    let duration = start.elapsed().as_millis();
    println!("  → Aggregated {} groups", count);
    
    Ok(duration)
}

/// Performs batch update within a transaction
fn benchmark_batch_update(conn: &Connection, count: usize) -> Result<u128> {
    let start = Instant::now();
    
    let tx = conn.unchecked_transaction()?;
    
    for i in 0..count {
        tx.execute(
            "UPDATE users SET age = ?1 WHERE id = ?2",
            params![30 + (i % 30) as i32, i + 1],
        )?;
    }
    
    tx.commit()?;
    
    Ok(start.elapsed().as_millis())
}

/// Performs batch delete within a transaction
fn benchmark_batch_delete(conn: &Connection, count: usize) -> Result<u128> {
    let start = Instant::now();
    
    let tx = conn.unchecked_transaction()?;
    
    tx.execute("DELETE FROM users WHERE id <= ?1", params![count])?;
    
    tx.commit()?;
    
    Ok(start.elapsed().as_millis())
}

/// Performs custom complex query on existing database (1000 iterations)
/// Uses random LIMIT, OFFSET, and release_date values for realistic testing
fn benchmark_custom_query(db_path: &str, iterations: usize) -> Result<u128> {
    let start = Instant::now();
    
    let conn = Connection::open_with_flags(
        db_path,
        rusqlite::OpenFlags::SQLITE_OPEN_READ_ONLY,
    )?;
    
    let query_template = "
        SELECT DISTINCT derived_video.dvd_id, derived_video.jacket_full_url, derived_video.release_date 
        FROM derived_video 
        LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
        LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
        WHERE derived_video.dvd_id IS NOT NULL 
        AND derived_video.dvd_id IS NOT '' 
        AND derived_video.release_date IS NOT NULL 
        AND derived_video.release_date <= ?1
        AND derived_video.jacket_full_url IS NOT NULL 
        AND (lower(derived_video.dvd_id) LIKE lower('%%') 
             OR lower(derived_actress.name_romaji) LIKE lower('%%') 
             OR lower(derived_actress.name_kanji) LIKE lower('%%') 
             OR lower(derived_actress.name_kana) LIKE lower('%%') 
             OR lower(derived_category.name_en) LIKE lower('%%') 
             OR lower(derived_category.name_ja) LIKE lower('%%')) 
        ORDER BY derived_video.release_date DESC
        LIMIT ?2 OFFSET ?3
    ";
    
    // Prepare statement once outside the loop for efficiency
    let mut stmt = conn.prepare(query_template)?;
    
    let mut total_rows = 0;
    for i in 0..iterations {
        // Generate random parameters
        let random_year = 2020 + (i % 6);
        let random_month = 1 + ((i * 7) % 12);
        let random_day = 1 + ((i * 11) % 28);
        let random_date = format!("{:04}-{:02}-{:02}", random_year, random_month, random_day);
        let random_limit = 50 + ((i * 13) % 150); // 50-199
        let random_offset = (i * 37) % 5000; // 0-4999
        
        let mut rows = stmt.query(params![random_date, random_limit, random_offset])?;
        
        let mut count = 0;
        while rows.next()?.is_some() {
            count += 1;
        }
        total_rows += count;
    }
    
    let duration = start.elapsed().as_millis();
    let avg_rows = total_rows / iterations;
    println!("  → Executed {} times, avg {} rows per query", iterations, avg_rows);
    
    Ok(duration)
}

fn main() -> Result<()> {
    println!("=== Rust SQLite Benchmark ===\n");
    
    // Remove old database file if exists
    let _ = fs::remove_file("benchmark.db");
    
    let conn = Connection::open("benchmark.db")?;
    setup_database(&conn)?;
    
    let total_start = Instant::now();
    
    // Batch Insert
    print!("1. Batch Insert (10,000 records)... ");
    let batch_insert_time = benchmark_batch_insert(&conn, 10_000)?;
    println!("{}ms", batch_insert_time);
    
    // Single Inserts
    print!("2. Single Inserts (1,000 records)... ");
    let single_insert_time = benchmark_single_inserts(&conn, 1_000)?;
    println!("{}ms", single_insert_time);
    
    // Simple Select
    print!("3. Simple Select (age > 30)... ");
    let simple_select_time = benchmark_simple_select(&conn)?;
    println!("{}ms", simple_select_time);
    
    // Complex Select
    print!("4. Complex Select (aggregation)... ");
    let complex_select_time = benchmark_complex_select(&conn)?;
    println!("{}ms", complex_select_time);
    
    // Batch Update
    print!("5. Batch Update (5,000 records)... ");
    let batch_update_time = benchmark_batch_update(&conn, 5_000)?;
    println!("{}ms", batch_update_time);
    
    // Batch Delete
    print!("6. Batch Delete (5,000 records)... ");
    let batch_delete_time = benchmark_batch_delete(&conn, 5_000)?;
    println!("{}ms", batch_delete_time);
    
    // Custom Query Benchmark on existing database
    println!("\n7. Custom Query (10 iterations on r18_25_11_04.sqlite)... ");
    let custom_query_time = benchmark_custom_query("../r18_25_11_04.sqlite", 10)?;
    println!("   {}ms", custom_query_time);
    
    let total_time = total_start.elapsed().as_millis();
    
    println!("\n=== Results ===");
    println!("Batch Insert:    {:>8}ms", batch_insert_time);
    println!("Single Inserts:  {:>8}ms", single_insert_time);
    println!("Simple Select:   {:>8}ms", simple_select_time);
    println!("Complex Select:  {:>8}ms", complex_select_time);
    println!("Batch Update:    {:>8}ms", batch_update_time);
    println!("Batch Delete:    {:>8}ms", batch_delete_time);
    println!("Custom Query:    {:>8}ms", custom_query_time);
    println!("─────────────────────────");
    println!("Total Time:      {:>8}ms", total_time);
    
    Ok(())
}

