# Bead 4.4: k6 Chaos Mode Testing

**Phase**: 4
**Priority**: High

## Description
Run the k6 harness in Chaos Mode to validate safety under error injection.

## Acceptance Criteria
- Chaos mode runs without crashing the server
- Dangerous queries are rejected
- Server remains stable

## Validation
- Review `validation_rejections_total` metric
- Confirm no successful mutations occurred