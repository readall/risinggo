# Bead 0.3: pgxpool Database Access Layer

**Phase**: 0 - Foundations & Safety Core
**Priority**: Critical

## Description
Implement the database access layer using `pgxpool` for efficient connection pooling to RisingWave.

## Tasks
- Create `internal/db/pool.go`
- Configure connection pool with `MinConns`, `MaxConns`, timeouts
- Add health check on startup
- Support context propagation for all queries
- Expose pool statistics for observability

## Acceptance Criteria
- Pool initializes successfully with valid connection string
- Health checks pass on startup
- Queries can be executed with context
- Pool metrics are available

## Validation
- Unit tests for pool creation and basic queries
- Integration test with real RisingWave using testcontainers
- Verify `pgxpool` stats are exposed

## Related Gherkin
- `technical.features`: "Server initializes connection pool correctly"
- ADR-003 in `implementation-constraints.md`