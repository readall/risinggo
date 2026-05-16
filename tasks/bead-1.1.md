# Bead 1.1: execute_safe_read_query Tool

**Phase**: 1 - Core Read-Only Functionality
**Priority**: Critical

## Description
Implement the main generic read-only query tool.

## Tasks
- Register `execute_safe_read_query` using typed MCP tool
- Integrate with Safety Layer before execution
- Execute query via pooled connection with timeout
- Return results in clean JSON format
- Include basic execution metadata

## Acceptance Criteria
- Valid SELECT queries execute successfully
- Dangerous queries are rejected before reaching DB
- Results are returned in structured format
- Timeouts are respected

## Validation
- Integration tests with real queries
- k6 functional coverage tests
- Rejection testing with dangerous queries

## Related Gherkin
- `public_api.features`: Agent uses the powerful generic safe read-only executor