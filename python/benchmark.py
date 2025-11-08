#!/usr/bin/env python3
"""
SQLite benchmark for Python using the built-in sqlite3 module.
"""

import sqlite3
import os
import time
from typing import Tuple


def setup_database(conn: sqlite3.Connection) -> None:
    """Creates the users table in the database."""
    conn.execute("""
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            age INTEGER NOT NULL
        )
    """)
    conn.commit()


def benchmark_batch_insert(conn: sqlite3.Connection, count: int) -> float:
    """Performs batch insert within a transaction."""
    start = time.time()
    
    cursor = conn.cursor()
    cursor.execute("BEGIN TRANSACTION")
    
    for i in range(count):
        cursor.execute(
            "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
            (f"User{i}", f"user{i}@example.com", 20 + (i % 50))
        )
    
    conn.commit()
    
    return (time.time() - start) * 1000


def benchmark_single_inserts(conn: sqlite3.Connection, count: int) -> float:
    """Performs single inserts without explicit transaction."""
    start = time.time()
    
    cursor = conn.cursor()
    
    for i in range(count):
        cursor.execute(
            "INSERT INTO users (name, email, age) VALUES (?, ?, ?)",
            (f"SingleUser{i}", f"single{i}@example.com", 25 + (i % 40))
        )
        conn.commit()
    
    return (time.time() - start) * 1000


def benchmark_simple_select(conn: sqlite3.Connection) -> float:
    """Performs simple SELECT query with WHERE clause."""
    start = time.time()
    
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM users WHERE age > ?", (30,))
    
    rows = cursor.fetchall()
    count = len(rows)
    
    duration = (time.time() - start) * 1000
    print(f"  → Found {count} records")
    
    return duration


def benchmark_complex_select(conn: sqlite3.Connection) -> float:
    """Performs complex SELECT query with aggregation."""
    start = time.time()
    
    cursor = conn.cursor()
    cursor.execute("""
        SELECT age, COUNT(*) as count, AVG(age) as avg_age 
        FROM users 
        WHERE age BETWEEN ? AND ? 
        GROUP BY age 
        ORDER BY count DESC 
        LIMIT 10
    """, (25, 50))
    
    rows = cursor.fetchall()
    count = len(rows)
    
    duration = (time.time() - start) * 1000
    print(f"  → Aggregated {count} groups")
    
    return duration


def benchmark_batch_update(conn: sqlite3.Connection, count: int) -> float:
    """Performs batch update within a transaction."""
    start = time.time()
    
    cursor = conn.cursor()
    cursor.execute("BEGIN TRANSACTION")
    
    for i in range(count):
        cursor.execute(
            "UPDATE users SET age = ? WHERE id = ?",
            (30 + (i % 30), i + 1)
        )
    
    conn.commit()
    
    return (time.time() - start) * 1000


def benchmark_batch_delete(conn: sqlite3.Connection, count: int) -> float:
    """Performs batch delete within a transaction."""
    start = time.time()
    
    cursor = conn.cursor()
    cursor.execute("BEGIN TRANSACTION")
    cursor.execute("DELETE FROM users WHERE id <= ?", (count,))
    conn.commit()
    
    return (time.time() - start) * 1000


def benchmark_custom_query(db_path: str, iterations: int) -> float:
    """Performs custom queries benchmark on existing database.
    Tests 3 different query patterns: index page, DVD detail, and DVD relationships.
    """
    import random
    
    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    cursor = conn.cursor()
    
    # Query 1: Index page query (listing with filters)
    index_query = """
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
    """
    
    # Query 2: DVD detail page query
    detail_query = """
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
        AND derived_video.dvd_id = ?
    """
    
    # Query 3: DVD relationships query (categories and actresses)
    relationships_query = """
        SELECT derived_video.content_id, derived_category.id AS cat_id, derived_category.name_en AS cat_name_en, 
               derived_category.name_ja AS cat_name_ja, derived_actress.id AS act_id, derived_actress.name_romaji, 
               derived_actress.name_kana, derived_actress.name_kanji, derived_actress.image_url AS act_image_url 
        FROM derived_video 
        LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
        LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
        WHERE derived_video.dvd_id = ?
    """
    
    # Query 4: Similar DVDs (same year, random order)
    similar_query = """
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
    """
    
    # Prepare statements (not directly supported in sqlite3, but we can reuse cursor)
    # Python sqlite3 caches the last query, but explicit preparation helps
    start = time.time()
    
    total_rows1 = 0
    total_rows2 = 0
    total_rows3 = 0
    total_rows4 = 0
    
    for _ in range(iterations):
        # Query 1: Index page with random parameters
        random_year = random.randint(2020, 2025)
        random_month = random.randint(1, 12)
        random_day = random.randint(1, 28)
        random_date = f"{random_year:04d}-{random_month:02d}-{random_day:02d}"
        page_number = random.randint(0, 49)  # Random page 0-49
        limit = 100
        offset = page_number * 100
        
        cursor.execute(index_query, (random_date, limit, offset))
        rows1 = cursor.fetchall()
        total_rows1 += len(rows1)
        
        # Query 2, 3, 4: Use a random dvd_id from Query 1 results
        if len(rows1) > 0:
            random_dvd_id = random.choice(rows1)[0]  # Get dvd_id from first column
            
            cursor.execute(detail_query, (random_dvd_id,))
            rows2 = cursor.fetchall()
            total_rows2 += len(rows2)
            
            cursor.execute(relationships_query, (random_dvd_id,))
            rows3 = cursor.fetchall()
            total_rows3 += len(rows3)
            
            cursor.execute(similar_query, (random_dvd_id,))
            rows4 = cursor.fetchall()
            total_rows4 += len(rows4)
    
    duration = (time.time() - start) * 1000
    
    conn.close()
    
    print(f"  → Query 1 (Index): {iterations} iterations, avg {total_rows1 // iterations} rows")
    print(f"  → Query 2 (Detail): {iterations} iterations, avg {total_rows2 // iterations} rows")
    print(f"  → Query 3 (Relations): {iterations} iterations, avg {total_rows3 // iterations} rows")
    print(f"  → Query 4 (Similar): {iterations} iterations, avg {total_rows4 // iterations} rows")
    
    return duration


def main():
    import sys
    
    # Check for --custom-queries flag
    custom_queries_only = "--custom-queries" in sys.argv
    
    if custom_queries_only:
        print("=== Python SQLite Benchmark - Custom Queries Only ===\n")
        
        total_start = time.time()
        
        # Custom Queries Benchmark on existing database
        print("Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ")
        custom_query_time = benchmark_custom_query("../r18_25_11_04.sqlite", 10)
        print(f"   Total: {custom_query_time:.0f}ms")
        
        total_time = (time.time() - total_start) * 1000
        
        print("\n=== Results ===")
        print(f"Custom Query:    {custom_query_time:>8.0f}ms")
        print("─────────────────────────")
        print(f"Total Time:      {total_time:>8.0f}ms")
        return
    
    print("=== Python SQLite Benchmark ===\n")
    
    # Remove old database file if exists
    if os.path.exists("benchmark.db"):
        os.remove("benchmark.db")
    
    conn = sqlite3.connect("benchmark.db")
    setup_database(conn)
    
    total_start = time.time()
    
    # Batch Insert
    print("1. Batch Insert (10,000 records)... ", end="", flush=True)
    batch_insert_time = benchmark_batch_insert(conn, 10_000)
    print(f"{batch_insert_time:.0f}ms")
    
    # Single Inserts
    print("2. Single Inserts (1,000 records)... ", end="", flush=True)
    single_insert_time = benchmark_single_inserts(conn, 1_000)
    print(f"{single_insert_time:.0f}ms")
    
    # Simple Select
    print("3. Simple Select (age > 30)... ", end="", flush=True)
    simple_select_time = benchmark_simple_select(conn)
    print(f"{simple_select_time:.0f}ms")
    
    # Complex Select
    print("4. Complex Select (aggregation)... ", end="", flush=True)
    complex_select_time = benchmark_complex_select(conn)
    print(f"{complex_select_time:.0f}ms")
    
    # Batch Update
    print("5. Batch Update (5,000 records)... ", end="", flush=True)
    batch_update_time = benchmark_batch_update(conn, 5_000)
    print(f"{batch_update_time:.0f}ms")
    
    # Batch Delete
    print("6. Batch Delete (5,000 records)... ", end="", flush=True)
    batch_delete_time = benchmark_batch_delete(conn, 5_000)
    print(f"{batch_delete_time:.0f}ms")
    
    conn.close()
    
    # Custom Queries Benchmark on existing database
    print("\n7. Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ")
    custom_query_time = benchmark_custom_query("../r18_25_11_04.sqlite", 10)
    print(f"   Total: {custom_query_time:.0f}ms")
    
    total_time = (time.time() - total_start) * 1000
    
    print("\n=== Results ===")
    print(f"Batch Insert:    {batch_insert_time:>8.0f}ms")
    print(f"Single Inserts:  {single_insert_time:>8.0f}ms")
    print(f"Simple Select:   {simple_select_time:>8.0f}ms")
    print(f"Complex Select:  {complex_select_time:>8.0f}ms")
    print(f"Batch Update:    {batch_update_time:>8.0f}ms")
    print(f"Batch Delete:    {batch_delete_time:>8.0f}ms")
    print(f"Custom Query:    {custom_query_time:>8.0f}ms")
    print("─────────────────────────")
    print(f"Total Time:      {total_time:>8.0f}ms")


if __name__ == "__main__":
    main()

