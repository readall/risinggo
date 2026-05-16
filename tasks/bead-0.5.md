# Bead 0.5: validateReadOnlyQuery and Resource Limits

**Phase**: 0
**Priority**: Critical

## Description
Implement the core function that validates whether a query is safe to execute.

## Tasks
- Create strict allow-list for query types (SELECT, WITH, EXPLAIN, SHOW, etc.)
- Implement keyword blocklist for mutations
- Add configurable resource limits (timeout, max rows)
- Return structured rejection reasons

## Acceptance Criteria
- Valid read queries pass
- All mutation keywords are blocked
- Timeouts and row limits are enforced
- Clear rejection messages are provided

## Validation
- Comprehensive unit tests with dangerous query list
- Chaos mode testing in k6

## Related Specs
- `invariants_and_non_functional.features`
- Safety scenarios in `public_api.features`