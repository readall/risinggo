# Bead 0.1: Project Initialization & Structure

**Phase**: 0 - Foundations & Safety Core
**Priority**: High

## Description
Set up the Go project with a clean, maintainable structure aligned with the recommended layout in the design document.

## Tasks
- Initialize Go module (`go mod init`)
- Create recommended directory structure (`cmd/`, `internal/`, `pkg/`, `tests/`)
- Add `.gitignore`, `Makefile`, and basic CI workflow
- Create initial `README.md` with project overview

## Acceptance Criteria
- Project compiles successfully
- Directory structure matches the one defined in `IMPLEMENTATION_PLAN.md` and design doc
- Basic build and test commands work via Makefile

## Validation
- `go build ./...` succeeds
- `go test ./...` runs without errors
- Folder structure is reviewed and approved