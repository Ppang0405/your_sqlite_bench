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

// benchmarkCustomQuery performs custom queries benchmark on existing database
// Tests 3 different query patterns: index page, DVD detail, and DVD relationships
func benchmarkCustomQuery(dbPath string, iterations int) (int64, error) {
	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	// Query 1: Index page query (listing with filters)
	indexQuery := `
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

	// Query 2: DVD detail page query
	detailQuery := `
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
	`

	// Query 3: DVD relationships query (categories and actresses)
	relationshipsQuery := `
		SELECT derived_video.content_id, derived_category.id AS cat_id, derived_category.name_en AS cat_name_en, 
		       derived_category.name_ja AS cat_name_ja, derived_actress.id AS act_id, derived_actress.name_romaji, 
		       derived_actress.name_kana, derived_actress.name_kanji, derived_actress.image_url AS act_image_url 
		FROM derived_video 
		LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
		LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
		LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
		LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
		WHERE derived_video.dvd_id = ?
	`

	// Query 4: Similar DVDs (same year, random order)
	similarQuery := `
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
	`

	// Prepare statements once before the loop
	stmt1, err := db.Prepare(indexQuery)
	if err != nil {
		return 0, err
	}
	defer stmt1.Close()

	stmt2, err := db.Prepare(detailQuery)
	if err != nil {
		return 0, err
	}
	defer stmt2.Close()

	stmt3, err := db.Prepare(relationshipsQuery)
	if err != nil {
		return 0, err
	}
	defer stmt3.Close()

	stmt4, err := db.Prepare(similarQuery)
	if err != nil {
		return 0, err
	}
	defer stmt4.Close()

	start := time.Now()

	var totalRows1, totalRows2, totalRows3, totalRows4 int

	for i := 0; i < iterations; i++ {
		// Query 1: Index page with random parameters
		randomYear := 2020 + (i % 6)
		randomMonth := 1 + (i*7 % 12)
		randomDay := 1 + (i*11 % 28)
		randomDate := fmt.Sprintf("%d-%02d-%02d", randomYear, randomMonth, randomDay)
		pageNumber := (i * 13) % 50 // Random page 0-49
		limit := 100
		offset := pageNumber * 100

		rows1, err := stmt1.Query(randomDate, limit, offset)
		if err != nil {
			return 0, err
		}

		// Collect Query 1 results
		var query1Results []string
		for rows1.Next() {
			var dvdID, jacketURL, releaseDate sql.NullString
			if err := rows1.Scan(&dvdID, &jacketURL, &releaseDate); err != nil {
				rows1.Close()
				return 0, err
			}
			if dvdID.Valid {
				query1Results = append(query1Results, dvdID.String)
			}
		}
		rows1.Close()
		totalRows1 += len(query1Results)

		// Query 2, 3, 4: Use a random dvd_id from Query 1 results
		if len(query1Results) == 0 {
			continue
		}
		randomDvdID := query1Results[i%len(query1Results)]

		rows2, err := stmt2.Query(randomDvdID)
		if err != nil {
			return 0, err
		}

		count2 := 0
		for rows2.Next() {
			var contentID, dvdID, titleEN, titleJA, commentEN, commentJA sql.NullString
			var runtimeMins, makerID, labelID, seriesID, siteID sql.NullInt64
			var releaseDate, sampleURL, jacketFullURL, jacketThumbURL sql.NullString
			var galleryFullFirst, galleryFullLast, galleryThumbFirst, galleryThumbLast sql.NullString
			var serviceCode sql.NullString

			if err := rows2.Scan(&contentID, &dvdID, &titleEN, &titleJA, &commentEN, &commentJA,
				&runtimeMins, &releaseDate, &sampleURL, &makerID, &labelID, &seriesID,
				&jacketFullURL, &jacketThumbURL, &galleryFullFirst, &galleryFullLast,
				&galleryThumbFirst, &galleryThumbLast, &siteID, &serviceCode); err != nil {
				rows2.Close()
				return 0, err
			}
			count2++
		}
		rows2.Close()
		totalRows2 += count2

		rows3, err := stmt3.Query(randomDvdID)
		if err != nil {
			return 0, err
		}

		count3 := 0
		for rows3.Next() {
			var contentID, catID, catNameEN, catNameJA, actID, nameRomaji, nameKana, nameKanji, actImageURL sql.NullString
			if err := rows3.Scan(&contentID, &catID, &catNameEN, &catNameJA, &actID,
				&nameRomaji, &nameKana, &nameKanji, &actImageURL); err != nil {
				rows3.Close()
				return 0, err
			}
			count3++
		}
		rows3.Close()
		totalRows3 += count3

		rows4, err := stmt4.Query(randomDvdID)
		if err != nil {
			return 0, err
		}

		count4 := 0
		for rows4.Next() {
			var dvdID, jacketURL, releaseDate sql.NullString
			if err := rows4.Scan(&dvdID, &jacketURL, &releaseDate); err != nil {
				rows4.Close()
				return 0, err
			}
			count4++
		}
		rows4.Close()
		totalRows4 += count4
	}

	duration := time.Since(start).Milliseconds()

	fmt.Printf("  → Query 1 (Index): %d iterations, avg %d rows\n", iterations, totalRows1/iterations)
	fmt.Printf("  → Query 2 (Detail): %d iterations, avg %d rows\n", iterations, totalRows2/iterations)
	fmt.Printf("  → Query 3 (Relations): %d iterations, avg %d rows\n", iterations, totalRows3/iterations)
	fmt.Printf("  → Query 4 (Similar): %d iterations, avg %d rows\n", iterations, totalRows4/iterations)

	return duration, nil
}

func main() {
	// Check for --custom-queries flag
	customQueriesOnly := len(os.Args) > 1 && os.Args[1] == "--custom-queries"

	if customQueriesOnly {
		fmt.Println("=== Go SQLite Benchmark - Custom Queries Only ===\n")

		totalStart := time.Now()

		// Custom Queries Benchmark on existing database
		fmt.Println("Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ")
		customQueryTime, err := benchmarkCustomQuery("../r18_25_11_04.sqlite", 10)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   Total: %dms\n", customQueryTime)

		totalTime := time.Since(totalStart).Milliseconds()

		fmt.Println("\n=== Results ===")
		fmt.Printf("Custom Query:    %8dms\n", customQueryTime)
		fmt.Println("─────────────────────────")
		fmt.Printf("Total Time:      %8dms\n", totalTime)
		return
	}

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

	// Custom Queries Benchmark on existing database
	fmt.Println("\n7. Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ")
	customQueryTime, err := benchmarkCustomQuery("../r18_25_11_04.sqlite", 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Total: %dms\n", customQueryTime)

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

