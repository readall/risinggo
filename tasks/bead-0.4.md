# Bead 0.4: Multi-layer Safety & Validation System

**Phase**: 0 - Foundations & Safety Core
**Priority**: Critical

## Description
Build the core safety system that enforces the strictly read-only posture of the server.

## Tasks
- Create `internal/safety/` package
- Implement keyword-based mutation detection
- Add resource limits (timeout, max rows, query length)
- Create validation pipeline (config → middleware → query validator)
- Log all rejection decisions with context

## Acceptance Criteria
- All dangerous queries (DROP, DELETE, INSERT, UPDATE, etc.) are rejected
- Rejection happens at the earliest possible layer
- Clear error messages are returned
- Rejections are logged

## Validation
- Unit tests covering all dangerous query patterns
- Run k6 in Chaos Mode and verify rejections
- Confirm no mutation ever reaches the database

## Related Gherkin
- `invariants_and_non_functional.features`: Strict Read-Only Invariant
- `public_api.features`: Generic executor rejects dangerous queries