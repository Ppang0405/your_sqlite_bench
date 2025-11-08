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

// benchmarkCustomQuery performs custom complex query on existing database (1000 iterations)
// Uses random LIMIT, OFFSET, and release_date values for realistic testing
func benchmarkCustomQuery(dbPath string, iterations int) (int64, error) {
	start := time.Now()

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	queryTemplate := `
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
	`

	var totalRows int
	for i := 0; i < iterations; i++ {
		// Generate random parameters
		randomYear := 2020 + (i % 6)                                  // 2020-2025
		randomMonth := 1 + (i*7 % 12)                                 // 1-12
		randomDay := 1 + (i*11 % 28)                                  // 1-28
		randomDate := fmt.Sprintf("%d-%02d-%02d", randomYear, randomMonth, randomDay)
		randomLimit := 50 + (i*13 % 150)  // 50-199
		randomOffset := (i * 37) % 5000   // 0-4999

		rows, err := db.Query(queryTemplate, randomDate, randomLimit, randomOffset)
		if err != nil {
			return 0, err
		}

		count := 0
		for rows.Next() {
			var dvdID, jacketURL, releaseDate sql.NullString
			if err := rows.Scan(&dvdID, &jacketURL, &releaseDate); err != nil {
				rows.Close()
				return 0, err
			}
			count++
		}
		rows.Close()

		if err := rows.Err(); err != nil {
			return 0, err
		}
		totalRows += count
	}

	duration := time.Since(start).Milliseconds()
	avgRows := totalRows / iterations
	fmt.Printf("  → Executed %d times, avg %d rows per query\n", iterations, avgRows)

	return duration, nil
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

	// Custom Query Benchmark on existing database
	fmt.Println("\n7. Custom Query (10 iterations on r18_25_11_04.sqlite)... ")
	customQueryTime, err := benchmarkCustomQuery("../r18_25_11_04.sqlite", 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %dms\n", customQueryTime)

	totalTime := time.Since(totalStart).Milliseconds()

	fmt.Println("\n=== Results ===")
	fmt.Printf("Batch Insert:    %8dms\n", batchInsertTime)
	fmt.Printf("Single Inserts:  %8dms\n", singleInsertTime)
	fmt.Printf("Simple Select:   %8dms\n", simpleSelectTime)
	fmt.Printf("Complex Select:  %8dms\n", complexSelectTime)
	fmt.Printf("Batch Update:    %8dms\n", batchUpdateTime)
	fmt.Printf("Batch Delete:    %8dms\n", batchDeleteTime)
	fmt.Printf("Custom Query:    %8dms\n", customQueryTime)
	fmt.Println("─────────────────────────")
	fmt.Printf("Total Time:      %8dms\n", totalTime)
}

