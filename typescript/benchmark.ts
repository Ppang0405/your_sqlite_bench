#!/usr/bin/env bun

/**
 * SQLite benchmark for Bun runtime using SQL from bun
 */

import { SQL } from "bun";
import { unlinkSync, existsSync } from "fs";

// Type alias for cleaner code
type DB = SQL;

/**
 * Creates the users table in the database
 */
async function setupDatabase(db: DB): Promise<void> {
  await db`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY,
      name TEXT NOT NULL,
      email TEXT NOT NULL,
      age INTEGER NOT NULL
    )
  `;
}

/**
 * Performs batch insert within a transaction
 */
async function benchmarkBatchInsert(db: DB, count: number): Promise<number> {
  const start = Date.now();
  
  await db.begin(async (tx) => {
    for (let i = 0; i < count; i++) {
      await tx`INSERT INTO users (name, email, age) VALUES (${`User${i}`}, ${`user${i}@example.com`}, ${20 + (i % 50)})`;
    }
  });
  
  return Date.now() - start;
}

/**
 * Performs single inserts without explicit transaction
 */
async function benchmarkSingleInserts(db: DB, count: number): Promise<number> {
  const start = Date.now();
  
  for (let i = 0; i < count; i++) {
    await db`INSERT INTO users (name, email, age) VALUES (${`SingleUser${i}`}, ${`single${i}@example.com`}, ${25 + (i % 40)})`;
  }
  
  return Date.now() - start;
}

/**
 * Performs simple SELECT query with WHERE clause
 */
async function benchmarkSimpleSelect(db: DB): Promise<number> {
  const start = Date.now();
  
  const rows = await db`SELECT * FROM users WHERE age > ${30}`;
  
  const duration = Date.now() - start;
  console.log(`  → Found ${rows.length} records`);
  
  return duration;
}

/**
 * Performs complex SELECT query with aggregation
 */
async function benchmarkComplexSelect(db: DB): Promise<number> {
  const start = Date.now();
  
  const rows = await db`
    SELECT age, COUNT(*) as count, AVG(age) as avg_age 
    FROM users 
    WHERE age BETWEEN ${25} AND ${50} 
    GROUP BY age 
    ORDER BY count DESC 
    LIMIT 10
  `;
  
  const duration = Date.now() - start;
  console.log(`  → Aggregated ${rows.length} groups`);
  
  return duration;
}

/**
 * Performs batch update within a transaction
 */
async function benchmarkBatchUpdate(db: DB, count: number): Promise<number> {
  const start = Date.now();
  
  await db.begin(async (tx) => {
    for (let i = 0; i < count; i++) {
      await tx`UPDATE users SET age = ${30 + (i % 30)} WHERE id = ${i + 1}`;
    }
  });
  
  return Date.now() - start;
}

/**
 * Performs batch delete within a transaction
 */
async function benchmarkBatchDelete(db: DB, count: number): Promise<number> {
  const start = Date.now();
  
  await db.begin(async (tx) => {
    await tx`DELETE FROM users WHERE id <= ${count}`;
  });
  
  return Date.now() - start;
}

/**
 * Performs custom queries benchmark on existing database
 * Tests 4 different query patterns: index page, DVD detail, DVD relationships, and similar DVDs
 */
async function benchmarkCustomQuery(dbPath: string, iterations: number): Promise<number> {
  const db = new SQL(`sqlite://${dbPath}`);
  
  const start = Date.now();
  
  let totalRows1 = 0, totalRows2 = 0, totalRows3 = 0, totalRows4 = 0;
  
  for (let i = 0; i < iterations; i++) {
    // Query 1: Index page with random parameters
    const randomYear = 2020 + Math.floor(Math.random() * 6);
    const randomMonth = 1 + Math.floor(Math.random() * 12);
    const randomDay = 1 + Math.floor(Math.random() * 28);
    const randomDate = `${randomYear}-${String(randomMonth).padStart(2, '0')}-${String(randomDay).padStart(2, '0')}`;
    const pageNumber = Math.floor(Math.random() * 50); // Random page 0-49
    const limit = 100;
    const offset = pageNumber * 100;
    
    const rows1 = await db`
      SELECT DISTINCT derived_video.dvd_id, derived_video.jacket_full_url, derived_video.release_date 
      FROM derived_video 
      LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
      LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
      LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
      LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
      WHERE derived_video.dvd_id IS NOT NULL 
      AND derived_video.dvd_id IS NOT '' 
      AND derived_video.release_date IS NOT NULL 
      AND derived_video.release_date <= ${randomDate} 
      AND derived_video.jacket_full_url IS NOT NULL 
      AND (lower(derived_video.dvd_id) LIKE lower('%%') 
           OR lower(derived_actress.name_romaji) LIKE lower('%%') 
           OR lower(derived_actress.name_kanji) LIKE lower('%%') 
           OR lower(derived_actress.name_kana) LIKE lower('%%') 
           OR lower(derived_category.name_en) LIKE lower('%%') 
           OR lower(derived_category.name_ja) LIKE lower('%%')) 
      ORDER BY derived_video.release_date DESC
      LIMIT ${limit} OFFSET ${offset}
    `;
    totalRows1 += rows1.length;
    
    // Query 2, 3, 4: Use a random dvd_id from Query 1 results
    if (rows1.length > 0) {
      const randomDvdId = rows1[Math.floor(Math.random() * rows1.length)].dvd_id;
      
      const rows2 = await db`
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
        AND derived_video.dvd_id = ${randomDvdId}
      `;
      totalRows2 += rows2.length;
      
      const rows3 = await db`
        SELECT derived_video.content_id, derived_category.id AS cat_id, derived_category.name_en AS cat_name_en, 
               derived_category.name_ja AS cat_name_ja, derived_actress.id AS act_id, derived_actress.name_romaji, 
               derived_actress.name_kana, derived_actress.name_kanji, derived_actress.image_url AS act_image_url 
        FROM derived_video 
        LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
        LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
        LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
        WHERE derived_video.dvd_id = ${randomDvdId}
      `;
      totalRows3 += rows3.length;
      
      const rows4 = await db`
        SELECT derived_video.dvd_id, derived_video.jacket_full_url, derived_video.release_date 
        FROM derived_video, (SELECT derived_video.release_date AS release_date 
                             FROM derived_video 
                             WHERE derived_video.dvd_id = ${randomDvdId}) AS anon_1 
        WHERE CAST(STRFTIME('%Y', derived_video.release_date) AS INTEGER) = CAST(STRFTIME('%Y', anon_1.release_date) AS INTEGER) 
        AND derived_video.dvd_id IS NOT NULL 
        AND derived_video.dvd_id != '' 
        AND derived_video.release_date IS NOT NULL 
        AND derived_video.jacket_full_url IS NOT NULL 
        ORDER BY random() 
        LIMIT 6 OFFSET 0
      `;
      totalRows4 += rows4.length;
    }
  }
  
  const duration = Date.now() - start;
  
  await db.close();
  
  console.log(`  → Query 1 (Index): ${iterations} iterations, avg ${Math.floor(totalRows1 / iterations)} rows`);
  console.log(`  → Query 2 (Detail): ${iterations} iterations, avg ${Math.floor(totalRows2 / iterations)} rows`);
  console.log(`  → Query 3 (Relations): ${iterations} iterations, avg ${Math.floor(totalRows3 / iterations)} rows`);
  console.log(`  → Query 4 (Similar): ${iterations} iterations, avg ${Math.floor(totalRows4 / iterations)} rows`);
  
  return duration;
}

async function main() {
  // Check for --custom-queries flag
  const customQueriesOnly = process.argv.includes('--custom-queries');
  
  if (customQueriesOnly) {
    console.log('=== TypeScript (Bun) SQLite Benchmark - Custom Queries Only ===\n');
    
    const totalStart = Date.now();
    
    // Custom Queries Benchmark on existing database
    console.log('Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ');
    const customQueryTime = await benchmarkCustomQuery('../r18_25_11_04.sqlite', 10);
    console.log(`   Total: ${customQueryTime}ms`);
    
    const totalTime = Date.now() - totalStart;
    
    console.log('\n=== Results ===');
    console.log(`Custom Query:    ${String(customQueryTime).padStart(8)}ms`);
    console.log('─────────────────────────');
    console.log(`Total Time:      ${String(totalTime).padStart(8)}ms`);
    return;
  }
  
  console.log('=== TypeScript (Bun) SQLite Benchmark ===\n');
  
  // Remove old database file if exists
  if (existsSync('benchmark.db')) {
    unlinkSync('benchmark.db');
  }
  
  const db = new SQL('sqlite://benchmark.db');
  await setupDatabase(db);
  
  const totalStart = Date.now();
  
  // Batch Insert
  process.stdout.write('1. Batch Insert (10,000 records)... ');
  const batchInsertTime = await benchmarkBatchInsert(db, 10_000);
  console.log(`${batchInsertTime}ms`);
  
  // Single Inserts
  process.stdout.write('2. Single Inserts (1,000 records)... ');
  const singleInsertTime = await benchmarkSingleInserts(db, 1_000);
  console.log(`${singleInsertTime}ms`);
  
  // Simple Select
  process.stdout.write('3. Simple Select (age > 30)... ');
  const simpleSelectTime = await benchmarkSimpleSelect(db);
  console.log(`${simpleSelectTime}ms`);
  
  // Complex Select
  process.stdout.write('4. Complex Select (aggregation)... ');
  const complexSelectTime = await benchmarkComplexSelect(db);
  console.log(`${complexSelectTime}ms`);
  
  // Batch Update
  process.stdout.write('5. Batch Update (5,000 records)... ');
  const batchUpdateTime = await benchmarkBatchUpdate(db, 5_000);
  console.log(`${batchUpdateTime}ms`);
  
  // Batch Delete
  process.stdout.write('6. Batch Delete (5,000 records)... ');
  const batchDeleteTime = await benchmarkBatchDelete(db, 5_000);
  console.log(`${batchDeleteTime}ms`);
  
  await db.close();
  
  // Custom Queries Benchmark on existing database
  console.log('\n7. Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ');
  const customQueryTime = await benchmarkCustomQuery('../r18_25_11_04.sqlite', 10);
  console.log(`   Total: ${customQueryTime}ms`);
  
  const totalTime = Date.now() - totalStart;
  
  console.log('\n=== Results ===');
  console.log(`Batch Insert:    ${String(batchInsertTime).padStart(8)}ms`);
  console.log(`Single Inserts:  ${String(singleInsertTime).padStart(8)}ms`);
  console.log(`Simple Select:   ${String(simpleSelectTime).padStart(8)}ms`);
  console.log(`Complex Select:  ${String(complexSelectTime).padStart(8)}ms`);
  console.log(`Batch Update:    ${String(batchUpdateTime).padStart(8)}ms`);
  console.log(`Batch Delete:    ${String(batchDeleteTime).padStart(8)}ms`);
  console.log(`Custom Query:    ${String(customQueryTime).padStart(8)}ms`);
  console.log('─────────────────────────');
  console.log(`Total Time:      ${String(totalTime).padStart(8)}ms`);
}

main().catch(console.error);

