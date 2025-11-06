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

async function main() {
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
  
  const totalTime = Date.now() - totalStart;
  
  db.close();
  
  console.log('\n=== Results ===');
  console.log(`Batch Insert:    ${String(batchInsertTime).padStart(8)}ms`);
  console.log(`Single Inserts:  ${String(singleInsertTime).padStart(8)}ms`);
  console.log(`Simple Select:   ${String(simpleSelectTime).padStart(8)}ms`);
  console.log(`Complex Select:  ${String(complexSelectTime).padStart(8)}ms`);
  console.log(`Batch Update:    ${String(batchUpdateTime).padStart(8)}ms`);
  console.log(`Batch Delete:    ${String(batchDeleteTime).padStart(8)}ms`);
  console.log('─────────────────────────');
  console.log(`Total Time:      ${String(totalTime).padStart(8)}ms`);
}

main().catch(console.error);

