#!/bin/bash

# Script to add indexes to the SQLite database for query optimization

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

DB_FILE="r18_25_11_04.sqlite"
SQL_FILE="optimize_indexes.sql"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  SQLite Index Optimization${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if database file exists
if [ ! -f "$DB_FILE" ]; then
    echo -e "${YELLOW}Error: Database file '$DB_FILE' not found!${NC}"
    exit 1
fi

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    echo -e "${YELLOW}Error: SQL file '$SQL_FILE' not found!${NC}"
    exit 1
fi

# Get database size before optimization
DB_SIZE_BEFORE=$(du -h "$DB_FILE" | cut -f1)
echo -e "Database: ${GREEN}$DB_FILE${NC}"
echo -e "Size before: ${GREEN}$DB_SIZE_BEFORE${NC}"
echo ""

# Show current indexes
echo -e "${BLUE}Current indexes:${NC}"
sqlite3 "$DB_FILE" "SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%';" | wc -l | xargs echo "Existing optimized indexes:"
echo ""

# Apply indexes
echo -e "${BLUE}Adding indexes...${NC}"
echo ""
sqlite3 "$DB_FILE" < "$SQL_FILE"
echo ""

# Get database size after optimization
DB_SIZE_AFTER=$(du -h "$DB_FILE" | cut -f1)
echo -e "Size after: ${GREEN}$DB_SIZE_AFTER${NC}"
echo ""

# Show all indexes created
echo -e "${BLUE}All optimized indexes:${NC}"
sqlite3 "$DB_FILE" "SELECT name FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%' ORDER BY name;"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Optimization Complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Tip: Run the benchmarks again to see performance improvements."
echo "     ./run_all.sh"

