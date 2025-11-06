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
    
    total_time = (time.time() - total_start) * 1000
    
    conn.close()
    
    print("\n=== Results ===")
    print(f"Batch Insert:    {batch_insert_time:>8.0f}ms")
    print(f"Single Inserts:  {single_insert_time:>8.0f}ms")
    print(f"Simple Select:   {simple_select_time:>8.0f}ms")
    print(f"Complex Select:  {complex_select_time:>8.0f}ms")
    print(f"Batch Update:    {batch_update_time:>8.0f}ms")
    print(f"Batch Delete:    {batch_delete_time:>8.0f}ms")
    print("─────────────────────────")
    print(f"Total Time:      {total_time:>8.0f}ms")


if __name__ == "__main__":
    main()

