# Bead 1.3: describe_table Tool

**Phase**: 1
**Priority**: High

## Description
Implement the `describe_table` tool.

## Tasks
- Accept `table_name` and optional `schema_name`
- Query column information from catalog
- Return structured column metadata
- Validate input identifiers

## Acceptance Criteria
- Returns accurate column names, types, and nullability
- Handles tables in different schemas
- Input validation prevents injection

## Validation
- Unit + integration tests
- Cross-check with `psql` or direct queries