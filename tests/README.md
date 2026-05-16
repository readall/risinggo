# Go Test Skeleton for RisingWave MCP Server

This folder contains a **test-first skeleton** aligned with the Gherkin specifications in `specs/`.

## Goals

- Map Gherkin scenarios to Go tests
- Use `testcontainers-go` for realistic integration testing against RisingWave
- Focus on safety, read-only behavior, and basic functionality
- Easy to expand as the implementation progresses

## Recommended Structure

```
tests/
├── safety_test.go          # Mutation rejection tests
├── schema_test.go          # Schema inspection tests
├── query_test.go           # Safe query execution
├── integration_test.go     # Full MCP server integration (testcontainers)
└── README.md
```

## How Tests Map to Gherkin

| Gherkin Feature                        | Primary Test File       | Key Test Functions |
|----------------------------------------|-------------------------|--------------------|
| `public_api.features`                  | `schema_test.go`, `query_test.go` | Test tool discovery and safe calls |
| `technical.features`                   | `integration_test.go`   | Validation pipeline, DB layer |
| `invariants_and_non_functional.features` | `safety_test.go`      | Read-only invariant tests |
| Rejection scenarios                    | `safety_test.go`        | `TestGenericExecutorRejectsMutations` |

## Running the Tests

```bash
go test ./tests -v
```

For integration tests that spin up RisingWave:

```bash
go test ./tests -run TestIntegration -v
```

## Dependencies (add to go.mod)

```go
require (
    github.com/testcontainers/testcontainers-go v0.33.0
    github.com/jackc/pgx/v5 v5.7.0
)
```

## Next Steps

1. Implement the actual MCP server.
2. Wire the test helpers to call your MCP server (HTTP or stdio).
3. Expand each `_test.go` file with more scenarios from the traceability matrix.