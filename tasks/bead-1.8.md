# Bead 1.8: Integration Tests with testcontainers

**Phase**: 1
**Priority**: High

## Description
Build integration test infrastructure using testcontainers-go.

## Tasks
- Set up testcontainers for RisingWave
- Create helper to start MCP server or connect to it
- Write tests for core tools (`execute_safe_read_query`, `show_tables`, `describe_table`)
- Test rejection of dangerous queries end-to-end

## Acceptance Criteria
- Tests spin up real RisingWave
- Core tools return correct results
- Safety rejections work in integration environment

## Validation
- All Phase 1 tools have passing integration tests
- Tests run reliably in CI