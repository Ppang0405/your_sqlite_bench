package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupDatabase creates the users table
func setupDatabase(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER NOT NULL
		)
	`)
	return err
}

// benchmarkBatchInsert performs batch insert within a transaction
func benchmarkBatchInsert(db *sql.DB, count int) (int64, error) {
	start := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO users (name, email, age) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for i := 0; i < count; i++ {
		_, err = stmt.Exec(
			fmt.Sprintf("User%d", i),
			fmt.Sprintf("user%d@example.com", i),
			20+(i%50),
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return time.Since(start).Milliseconds(), nil
}

// benchmarkSingleInserts performs single inserts without explicit transaction
func benchmarkSingleInserts(db *sql.DB, count int) (int64, error) {
	start := time.Now()

	stmt, err := db.Prepare("INSERT INTO users (name, email, age) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for i := 0; i < count; i++ {
		_, err = stmt.Exec(
			fmt.Sprintf("SingleUser%d", i),
			fmt.Sprintf("single%d@example.com", i),
			25+(i%40),
		)
		if err != nil {
			return 0, err
		}
	}

	return time.Since(start).Milliseconds(), nil
}

// benchmarkSimpleSelect performs simple SELECT query with WHERE clause
func benchmarkSimpleSelect(db *sql.DB) (int64, error) {
	start := time.Now()

	rows, err := db.Query("SELECT * FROM users WHERE age > ?", 30)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, age int
		var name, email string
		if err := rows.Scan(&id, &name, &email, &age); err != nil {
			return 0, err
		}
		count++
	}

	duration := time.Since(start).Milliseconds()
	fmt.Printf("  → Found %d records\n", count)

	return duration, rows.Err()
}

// benchmarkComplexSelect performs complex SELECT query with aggregation
func benchmarkComplexSelect(db *sql.DB) (int64, error) {
	start := time.Now()

	rows, err := db.Query(`
		SELECT age, COUNT(*) as count, AVG(age) as avg_age 
		FROM users 
		WHERE age BETWEEN ? AND ? 
		GROUP BY age 
		ORDER BY count DESC 
		LIMIT 10
	`, 25, 50)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var age, cnt int
		var avgAge float64
		if err := rows.Scan(&age, &cnt, &avgAge); err != nil {
			return 0, err
		}
		count++
	}

	duration := time.Since(start).Milliseconds()
	fmt.Printf("  → Aggregated %d groups\n", count)

	return duration, rows.Err()
}

// benchmarkBatchUpdate performs batch update within a transaction
func benchmarkBatchUpdate(db *sql.DB, count int) (int64, error) {
	start := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE users SET age = ? WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for i := 0; i < count; i++ {
		_, err = stmt.Exec(30+(i%30), i+1)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return time.Since(start).Milliseconds(), nil
}

// benchmarkBatchDelete performs batch delete within a transaction
func benchmarkBatchDelete(db *sql.DB, count int) (int64, error) {
	start := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM users WHERE id <= ?", count)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return time.Since(start).Milliseconds(), nil
}

func main() {
	fmt.Println("=== Go SQLite Benchmark ===\n")

	// Remove old database file if exists
	os.Remove("benchmark.db")

	db, err := sql.Open("sqlite3", "benchmark.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := setupDatabase(db); err != nil {
		log.Fatal(err)
	}

	totalStart := time.Now()

	// Batch Insert
	fmt.Print("1. Batch Insert (10,000 records)... ")
	batchInsertTime, err := benchmarkBatchInsert(db, 10_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dms\n", batchInsertTime)

	// Single Inserts
	fmt.Print("2. Single Inserts (1,000 records)... ")
	singleInsertTime, err := benchmarkSingleInserts(db, 1_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dms\n", singleInsertTime)

	// Simple Select
	fmt.Print("3. Simple Select (age > 30)... ")
	simpleSelectTime, err := benchmarkSimpleSelect(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dms\n", simpleSelectTime)

	// Complex Select
	fmt.Print("4. Complex Select (aggregation)... ")
	complexSelectTime, err := benchmarkComplexSelect(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dms\n", complexSelectTime)

	// Batch Update
	fmt.Print("5. Batch Update (5,000 records)... ")
	batchUpdateTime, err := benchmarkBatchUpdate(db, 5_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dms\n", batchUpdateTime)

	// Batch Delete
	fmt.Print("6. Batch Delete (5,000 records)... ")
	batchDeleteTime, err := benchmarkBatchDelete(db, 5_000)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%dms\n", batchDeleteTime)

	totalTime := time.Since(totalStart).Milliseconds()

	fmt.Println("\n=== Results ===")
	fmt.Printf("Batch Insert:    %8dms\n", batchInsertTime)
	fmt.Printf("Single Inserts:  %8dms\n", singleInsertTime)
	fmt.Printf("Simple Select:   %8dms\n", simpleSelectTime)
	fmt.Printf("Complex Select:  %8dms\n", complexSelectTime)
	fmt.Printf("Batch Update:    %8dms\n", batchUpdateTime)
	fmt.Printf("Batch Delete:    %8dms\n", batchDeleteTime)
	fmt.Println("─────────────────────────")
	fmt.Printf("Total Time:      %8dms\n", totalTime)
}

