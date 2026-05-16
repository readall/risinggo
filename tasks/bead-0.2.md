# Bead 0.2: Configuration System

**Phase**: 0
**Priority**: High

## Description
Build a robust configuration system using environment variables with validation.

## Tasks
- Use `viper` or standard library for env var loading
- Define configuration struct with validation tags
- Implement startup validation that fails fast on missing/invalid config
- Support key settings: DB connection, read-only mode, pool settings, timeouts

## Acceptance Criteria
- Server refuses to start with missing critical config
- Configuration is validated at startup
- Read-only mode flag is properly loaded

## Validation
- Unit tests for config parsing and validation
- Integration test: server fails to start with bad config
- Verify `READ_ONLY=true` is respected