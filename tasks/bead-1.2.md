# Bead 1.2: show_tables Tool

**Phase**: 1
**Priority**: High

## Description
Implement the `show_tables` tool to list tables in the database.

## Tasks
- Create handler in `internal/tools/schema/`
- Execute safe query against information_schema or RisingWave catalog
- Return results in clean JSON format
- Add proper error handling

## Acceptance Criteria
- Returns list of tables with schema information
- Works on valid databases
- Rejects invalid input

## Validation
- Integration test with real RisingWave
- Compare output with direct SQL query
- Part of k6 functional coverage