#!/usr/bin/env node

/**
 * SQLite benchmark for Node.js using better-sqlite3
 */

import Database from 'better-sqlite3';
import { unlink } from 'fs/promises';
import { existsSync } from 'fs';

/**
 * Creates the users table in the database
 */
function setupDatabase(db) {
  db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY,
      name TEXT NOT NULL,
      email TEXT NOT NULL,
      age INTEGER NOT NULL
    )
  `);
}

/**
 * Performs batch insert within a transaction
 */
function benchmarkBatchInsert(db, count) {
  const start = Date.now();
  
  const insert = db.prepare('INSERT INTO users (name, email, age) VALUES (?, ?, ?)');
  
  const insertMany = db.transaction((records) => {
    for (const record of records) {
      insert.run(record.name, record.email, record.age);
    }
  });
  
  const records = [];
  for (let i = 0; i < count; i++) {
    records.push({
      name: `User${i}`,
      email: `user${i}@example.com`,
      age: 20 + (i % 50)
    });
  }
  
  insertMany(records);
  
  return Date.now() - start;
}

/**
 * Performs single inserts without explicit transaction
 */
function benchmarkSingleInserts(db, count) {
  const start = Date.now();
  
  const insert = db.prepare('INSERT INTO users (name, email, age) VALUES (?, ?, ?)');
  
  for (let i = 0; i < count; i++) {
    insert.run(
      `SingleUser${i}`,
      `single${i}@example.com`,
      25 + (i % 40)
    );
  }
  
  return Date.now() - start;
}

/**
 * Performs simple SELECT query with WHERE clause
 */
function benchmarkSimpleSelect(db) {
  const start = Date.now();
  
  const stmt = db.prepare('SELECT * FROM users WHERE age > ?');
  const rows = stmt.all(30);
  
  const duration = Date.now() - start;
  console.log(`  → Found ${rows.length} records`);
  
  return duration;
}

/**
 * Performs complex SELECT query with aggregation
 */
function benchmarkComplexSelect(db) {
  const start = Date.now();
  
  const stmt = db.prepare(`
    SELECT age, COUNT(*) as count, AVG(age) as avg_age 
    FROM users 
    WHERE age BETWEEN ? AND ? 
    GROUP BY age 
    ORDER BY count DESC 
    LIMIT 10
  `);
  
  const rows = stmt.all(25, 50);
  
  const duration = Date.now() - start;
  console.log(`  → Aggregated ${rows.length} groups`);
  
  return duration;
}

/**
 * Performs batch update within a transaction
 */
function benchmarkBatchUpdate(db, count) {
  const start = Date.now();
  
  const update = db.prepare('UPDATE users SET age = ? WHERE id = ?');
  
  const updateMany = db.transaction(() => {
    for (let i = 0; i < count; i++) {
      update.run(30 + (i % 30), i + 1);
    }
  });
  
  updateMany();
  
  return Date.now() - start;
}

/**
 * Performs batch delete within a transaction
 */
function benchmarkBatchDelete(db, count) {
  const start = Date.now();
  
  const deleteStmt = db.prepare('DELETE FROM users WHERE id <= ?');
  
  const deleteMany = db.transaction(() => {
    deleteStmt.run(count);
  });
  
  deleteMany();
  
  return Date.now() - start;
}

/**
 * Performs custom queries benchmark on existing database
 * Tests 3 different query patterns: index page, DVD detail, and DVD relationships
 */
function benchmarkCustomQuery(dbPath, iterations) {
  const db = new Database(dbPath, { readonly: true });
  
  // Query 1: Index page query (listing with filters)
  const indexQuery = `
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
  `;
  
  // Query 2: DVD detail page query
  const detailQuery = `
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
  `;
  
  // Query 3: DVD relationships query (categories and actresses)
  const relationshipsQuery = `
    SELECT derived_video.content_id, derived_category.id AS cat_id, derived_category.name_en AS cat_name_en, 
           derived_category.name_ja AS cat_name_ja, derived_actress.id AS act_id, derived_actress.name_romaji, 
           derived_actress.name_kana, derived_actress.name_kanji, derived_actress.image_url AS act_image_url 
    FROM derived_video 
    LEFT OUTER JOIN derived_video_category ON derived_video_category.content_id = derived_video.content_id 
    LEFT OUTER JOIN derived_category ON derived_category.id = derived_video_category.category_id 
    LEFT OUTER JOIN derived_video_actress ON derived_video_actress.content_id = derived_video.content_id 
    LEFT OUTER JOIN derived_actress ON derived_actress.id = derived_video_actress.actress_id 
    WHERE derived_video.dvd_id = ?
  `;
  
  // Query 4: Similar DVDs (same year, random order)
  const similarQuery = `
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
  `;
  
  const stmt1 = db.prepare(indexQuery);
  const stmt2 = db.prepare(detailQuery);
  const stmt3 = db.prepare(relationshipsQuery);
  const stmt4 = db.prepare(similarQuery);
  
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
    
    const rows1 = stmt1.all(randomDate, limit, offset);
    totalRows1 += rows1.length;
    
    // Query 2, 3, 4: Use a random dvd_id from Query 1 results
    if (rows1.length > 0) {
      const randomDvdId = rows1[Math.floor(Math.random() * rows1.length)].dvd_id;
      
      const rows2 = stmt2.all(randomDvdId);
      totalRows2 += rows2.length;
      
      const rows3 = stmt3.all(randomDvdId);
      totalRows3 += rows3.length;
      
      const rows4 = stmt4.all(randomDvdId);
      totalRows4 += rows4.length;
    }
  }
  
  const duration = Date.now() - start;
  
  db.close();
  
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
    console.log('=== JavaScript (Node.js) SQLite Benchmark - Custom Queries Only ===\n');
    
    const totalStart = Date.now();
    
    // Custom Queries Benchmark on existing database
    console.log('Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ');
    const customQueryTime = benchmarkCustomQuery('../r18_25_11_04.sqlite', 10);
    console.log(`   Total: ${customQueryTime}ms`);
    
    const totalTime = Date.now() - totalStart;
    
    console.log('\n=== Results ===');
    console.log(`Custom Query:    ${String(customQueryTime).padStart(8)}ms`);
    console.log('─────────────────────────');
    console.log(`Total Time:      ${String(totalTime).padStart(8)}ms`);
    return;
  }
  
  console.log('=== JavaScript (Node.js) SQLite Benchmark ===\n');
  
  // Remove old database file if exists
  if (existsSync('benchmark.db')) {
    await unlink('benchmark.db');
  }
  
  const db = new Database('benchmark.db');
  setupDatabase(db);
  
  const totalStart = Date.now();
  
  // Batch Insert
  process.stdout.write('1. Batch Insert (10,000 records)... ');
  const batchInsertTime = benchmarkBatchInsert(db, 10_000);
  console.log(`${batchInsertTime}ms`);
  
  // Single Inserts
  process.stdout.write('2. Single Inserts (1,000 records)... ');
  const singleInsertTime = benchmarkSingleInserts(db, 1_000);
  console.log(`${singleInsertTime}ms`);
  
  // Simple Select
  process.stdout.write('3. Simple Select (age > 30)... ');
  const simpleSelectTime = benchmarkSimpleSelect(db);
  console.log(`${simpleSelectTime}ms`);
  
  // Complex Select
  process.stdout.write('4. Complex Select (aggregation)... ');
  const complexSelectTime = benchmarkComplexSelect(db);
  console.log(`${complexSelectTime}ms`);
  
  // Batch Update
  process.stdout.write('5. Batch Update (5,000 records)... ');
  const batchUpdateTime = benchmarkBatchUpdate(db, 5_000);
  console.log(`${batchUpdateTime}ms`);
  
  // Batch Delete
  process.stdout.write('6. Batch Delete (5,000 records)... ');
  const batchDeleteTime = benchmarkBatchDelete(db, 5_000);
  console.log(`${batchDeleteTime}ms`);
  
  db.close();
  
  // Custom Queries Benchmark on existing database
  console.log('\n7. Custom Queries (4 queries × 10 iterations on r18_25_11_04.sqlite)... ');
  const customQueryTime = benchmarkCustomQuery('../r18_25_11_04.sqlite', 10);
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

