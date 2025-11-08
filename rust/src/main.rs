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

/// Performs custom queries benchmark on existing database
/// Tests 3 different query patterns: index page, DVD detail, and DVD relationships
fn benchmark_custom_query(db_path: &str, iterations: usize) -> Result<u128> {
    let conn = Connection::open_with_flags(
        db_path,
        rusqlite::OpenFlags::SQLITE_OPEN_READ_ONLY,
    )?;
    
    // Query 1: Index page query (listing with filters)
    let index_query = "
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
    
    // Query 2: DVD detail page query
    let detail_query = "
        SELECT derived_video.content_id, derived_video.dvd_id, derived_video.title_en, derived_video.title_ja, 
               derived_video.comment_en, derived_video.comment_ja, derived_video.runtime_mins, derived_video.release_date, 
               derived_video.sample_url, derived_video.maker_id, derived_video.label_id, derived_video.series_id, 
               derived_video.jacket_full_url, derived_video.jacket_thumb_url, derived_video.gallery_full_first, 
               derived_video.gallery_full_last, derived_video.gallery_thumb_first, derived_video.gallery_thumb_last, 
               derived_video.site_id, derived_video.service_code 
        FROM derived_video 
        WHERE derived_video.dvd_id IS NOT NULL 
        AND derived_video.dvd_id IS NOT '' 
        AND derived_video.release_date IS NOT NULL 
        AND derived_video.dvd_id = ?1
    ";
    
    // Query 3: DVD relationships query (categories and actresses)
    let relationships_query = "
        SELECT derived_video.content_id, derived_category.id AS cat_id, derived_category.name_en AS cat_name_en, 
               derived_category.name_ja AS cat_name_ja, derived_actress.id AS act_id, derived_actress.name_romaji, 
               derived_actress.name_kana, derived_actress.name_kanji, derived_actress.image_url AS act_image_url 
        FROM derived_video 
        LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
        LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
        WHERE derived_video.dvd_id = ?1
    ";
    
    // Query 4: Similar DVDs (same year, random order)
    let similar_query = "
        SELECT derived_video.dvd_id, derived_video.jacket_full_url, derived_video.release_date 
        FROM derived_video, (SELECT derived_video.release_date AS release_date 
                             FROM derived_video 
                             WHERE derived_video.dvd_id = ?1) AS anon_1 
        WHERE CAST(STRFTIME('%Y', derived_video.release_date) AS INTEGER) = CAST(STRFTIME('%Y', anon_1.release_date) AS INTEGER) 
        AND derived_video.dvd_id IS NOT NULL 
        AND derived_video.dvd_id != '' 
        AND derived_video.release_date IS NOT NULL 
        AND derived_video.jacket_full_url IS NOT NULL 
        ORDER BY random() 
        LIMIT 6 OFFSET 0
    ";
    
    let mut stmt1 = conn.prepare(index_query)?;
    let mut stmt2 = conn.prepare(detail_query)?;
    let mut stmt3 = conn.prepare(relationships_query)?;
    let mut stmt4 = conn.prepare(similar_query)?;
    
    let start = Instant::now();
    
    let mut total_rows1 = 0;
    let mut total_rows2 = 0;
    let mut total_rows3 = 0;
    let mut total_rows4 = 0;
    
    for i in 0..iterations {
        // Query 1: Index page with random parameters
        let random_year = 2020 + (i % 6);
        let random_month = 1 + ((i * 7) % 12);
        let random_day = 1 + ((i * 11) % 28);
        let random_date = format!("{:04}-{:02}-{:02}", random_year, random_month, random_day);
        let page_number = (i * 13) % 50; // Random page 0-49
        let limit = 100;
        let offset = page_number * 100;
        
        // Collect Query 1 results
        let mut rows1 = stmt1.query(params![random_date, limit, offset])?;
        let mut query1_results: Vec<String> = Vec::new();
        while let Some(row) = rows1.next()? {
            let dvd_id: String = row.get(0)?;
            query1_results.push(dvd_id);
        }
        total_rows1 += query1_results.len();
        
        // Query 2, 3, 4: Use a random dvd_id from Query 1 results
        if query1_results.is_empty() {
            continue;
        }
        let random_dvd_id = &query1_results[i % query1_results.len()];
        
        let mut rows2 = stmt2.query(params![random_dvd_id])?;
        let mut count2 = 0;
        while rows2.next()?.is_some() {
            count2 += 1;
        }
        total_rows2 += count2;
        
        let mut rows3 = stmt3.query(params![random_dvd_id])?;
        let mut count3 = 0;
        while rows3.next()?.is_some() {
            count3 += 1;
        }
        total_rows3 += count3;
        
        let mut rows4 = stmt4.query(params![random_dvd_id])?;
        let mut count4 = 0;
        while rows4.next()?.is_some() {
            count4 += 1;
        }
        total_rows4 += count4;
    }
    
    let duration = start.elapsed().as_millis();
    
    println!("  → Query 1 (Index): {} iterations, avg {} rows", iterations, total_rows1 / iterations);
    println!("  → Query 2 (Detail): {} iterations, avg {} rows", iterations, total_rows2 / iterations);
    println!("  → Query 3 (Relations): {} iterations, avg {} rows", iterations, total_rows3 / iterations);
    println!("  → Query 4 (Similar): {} iterations, avg {} rows", iterations, total_rows4 / iterations);
    
    Ok(duration)
}

fn main() -> Result<()> {
    // Check for --custom-queries flag
    let args: Vec<String> = std::env::args().collect();
    let custom_queries_only = args.len() > 1 && args[1] == "--custom-queries";

    if custom_queries_only {
        println!("=== Rust SQLite Benchmark - Custom Queries Only ===\n");
        
        let total_start = Instant::now();
        
        // Custom Queries Benchmark on existing database
        println!("Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ");
        let custom_query_time = benchmark_custom_query("../r18_25_11_04.sqlite", 10)?;
        println!("   Total: {}ms", custom_query_time);
        
        let total_time = total_start.elapsed().as_millis();
        
        println!("\n=== Results ===");
        println!("Custom Query:    {:>8}ms", custom_query_time);
        println!("─────────────────────────");
        println!("Total Time:      {:>8}ms", total_time);
        
        return Ok(());
    }

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
    
    // Custom Queries Benchmark on existing database
    println!("\n7. Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ");
    let custom_query_time = benchmark_custom_query("../r18_25_11_04.sqlite", 10)?;
    println!("   Total: {}ms", custom_query_time);
    
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

