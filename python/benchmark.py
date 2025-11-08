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
    """Performs custom complex query on existing database (1000 iterations).
    Uses random LIMIT, OFFSET, and release_date values for realistic testing.
    """
    import random
    
    start = time.time()
    
    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    cursor = conn.cursor()
    
    query_template = """
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
    
    total_rows = 0
    for _ in range(iterations):
        # Generate random parameters
        random_year = random.randint(2020, 2025)
        random_month = random.randint(1, 12)
        random_day = random.randint(1, 28)
        random_date = f"{random_year:04d}-{random_month:02d}-{random_day:02d}"
        random_limit = random.randint(50, 199)  # 50-199
        random_offset = random.randint(0, 4999)  # 0-4999
        
        cursor.execute(query_template, (random_date, random_limit, random_offset))
        rows = cursor.fetchall()
        total_rows += len(rows)
    
    conn.close()
    
    duration = (time.time() - start) * 1000
    avg_rows = total_rows // iterations
    print(f"  → Executed {iterations} times, avg {avg_rows} rows per query")
    
    return duration


def main():
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
    
    # Custom Query Benchmark on existing database
    print("\n7. Custom Query (10 iterations on r18_25_11_04.sqlite)... ")
    custom_query_time = benchmark_custom_query("../r18_25_11_04.sqlite", 10)
    print(f"   {custom_query_time:.0f}ms")
    
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

