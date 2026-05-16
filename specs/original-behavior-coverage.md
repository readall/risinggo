# Coverage Analysis: Our Specs vs Original risingwave-mcp (Python)

**Date**: May 2026  
**Purpose**: Validate how well our Gherkin specifications cover the behavior of the original [risingwavelabs/risingwave-mcp](https://github.com/risingwavelabs/risingwave-mcp).

## Summary

Our specifications provide **strong coverage of the read-only and monitoring behaviors** of the original project, while **intentionally and explicitly excluding** all mutating capabilities (DDL/DML) as per our core design decision (ADR-001).

### Overall Coverage Assessment

| Category                        | Original Has | Our Specs Coverage | Notes |
|--------------------------------|--------------|--------------------|-------|
| Query Execution (SELECT)       | Yes          | Excellent          | Strong coverage via dedicated + generic tool |
| Explain Plans                  | Yes          | Good               | Covered in technical + public API |
| Schema Inspection              | Yes          | Excellent          | Many scenarios |
| Streaming Job Monitoring       | Yes          | Excellent          | Well represented |
| Storage / Hummock Analysis     | Yes          | Good               | Covered |
| Cluster & Management (read)    | Yes          | Good               | Covered |
| **DDL Tools**                  | Yes          | **Intentionally Excluded** | Core design decision |
| **DML Tools**                  | Yes          | **Intentionally Excluded** | Core design decision |

## Detailed Gap Analysis

### 1. Behaviors Well Covered
- Safe query execution (`run_select_query` equivalent + powerful generic `execute_safe_read_query`)
- Schema tools (`show_tables`, `describe_table`, `list_materialized_views`, etc.)
- Streaming monitoring (`list_streaming_jobs`, backfill progress, fragments/actors)
- Storage analysis
- Explain tools
- Basic cluster/version info
- Concurrent usage and stability

### 2. Behaviors Intentionally Not Covered (By Design)
- All `create_*`, `drop_*`, `alter_*`, `insert/update/delete` tools
- Any tool that modifies schema, data, parallelism, rate limits, sources, or sinks
- User/secret management that involves writes

These are **correctly excluded** per our strictly read-only architecture.

### 3. Minor Gaps / Recommendations for Future Enhancement
- More specific scenarios for Iceberg maintenance metadata tools
- Deeper catalog query variations
- Specific error messages matching the original Python implementation (nice-to-have for agent compatibility)
- Testing of very large result sets and pagination behavior (if applicable)

## Traceability Matrix: Original Tools → Our Specifications

This matrix maps key tools/categories from the original `risingwavelabs/risingwave-mcp` to scenarios in our Gherkin specifications.

**Legend**:
- **Covered** — Explicit scenario(s) exist
- **Partially Covered** — High-level coverage, more detail possible
- **Excluded by Design** — Intentionally removed (strictly read-only)
- **Gap** — Not yet covered in current specs

### Query & Explain Tools

| Original Tool              | Category     | Our Coverage          | Relevant Gherkin Scenarios                          | Notes |
|---------------------------|--------------|-----------------------|-----------------------------------------------------|-------|
| `run_select_query`        | Query        | Covered              | `public_api.features`: "Agent executes a simple safe SELECT query" | Strong coverage via both dedicated and generic tool |
| `table_row_count`         | Query        | Partially Covered    | Covered under generic + schema tools                | Can be achieved via `execute_safe_read_query` |
| `get_table_stats`         | Query        | Covered              | Schema inspection scenarios                         | - |
| `explain_query`           | Explain      | Covered              | `public_api.features` + `technical.features`        | - |
| `explain_analyze`         | Explain      | Covered              | Same as above                                       | - |
| `explain_distsql`         | Explain      | Partially Covered    | General explain coverage                            | Distributed plan scenarios can be expanded |

### Schema Tools

| Original Tool                     | Our Coverage     | Relevant Scenarios                              | Notes |
|-----------------------------------|------------------|--------------------------------------------------|-------|
| `show_tables`                     | Covered         | `public_api.features`                            | Frequently used in k6 |
| `describe_table`                  | Covered         | `public_api.features` + k6 functional coverage   | - |
| `list_materialized_views`         | Covered         | Schema inspection scenarios                      | - |
| `show_create_table` / `show_create_mv` | Partially Covered | General schema tools                            | Can use generic query |
| `get_table_columns`, `list_schemas` | Covered        | Schema category scenarios                        | - |
| Most other schema listing tools   | Covered         | Grouped under "Schema Inspection"                | High coverage |

### Streaming & Monitoring Tools

| Original Tool                  | Our Coverage | Relevant Scenarios                              | Notes |
|--------------------------------|--------------|--------------------------------------------------|-------|
| `list_streaming_jobs`          | Covered     | `public_api.features` + k6                       | Frequently tested |
| Backfill progress / fragments  | Covered     | Streaming monitoring scenarios                   | - |
| Actor / fragment monitoring    | Covered     | Technical + public API scenarios                 | - |
| `get_hummock_stats`            | Covered     | Storage analysis scenarios                       | - |
| Compaction / storage tools     | Partially Covered | Storage category                              | Good but can be expanded |

### Cluster & Management (Read-only)

| Original Tool             | Our Coverage | Notes |
|---------------------------|--------------|-------|
| `get_cluster_info`, `get_version` | Covered | Covered in public API and k6 |
| Session variables         | Partially Covered | Basic coverage |
| Most read-only management | Covered | Grouped under Cluster & Management |

### Intentionally Excluded Categories (By Design)

| Original Category     | Tools Included                  | Our Status              | Rationale |
|-----------------------|---------------------------------|-------------------------|---------|
| **DDL Tools**         | create/drop/alter table, MV, column, source, sink, etc. | **Excluded by Design** | ADR-001: Strictly read-only |
| **DML Tools**         | insert, update, delete          | **Excluded by Design** | ADR-001 |
| Source/Sink mutation  | alter_source, create_kafka_table, etc. | **Excluded by Design** | Safety |
| User/Secret management (write) | create user, grant, etc.     | **Excluded by Design** | Security model |

### New Capabilities Added (Not in Original)

| Our Addition                        | Coverage Type     | Benefit |
|-------------------------------------|-------------------|--------|
| Multi-layer safety validation       | Excellent        | Major improvement |
| `execute_safe_read_query` with rejection testing | Excellent | Powerful yet safe generic tool |
| p99 < 20ms performance targets      | Excellent        | New non-functional requirement |
| Very detailed observability         | Excellent        | Significantly stronger than original |
| Explicit 200 concurrent user testing| Excellent        | Scale validation |
| Mutation rejection scenarios        | Excellent        | Strong safety testing |

## Conclusion & Recommendations

- **Read-only surface**: Very well covered.
- **Mutating surface**: Correctly and completely excluded.
- **New quality attributes** (performance, observability, safety): Significantly stronger than original.

**Recommended Next Steps**:
1. Expand specific scenarios for Iceberg and advanced catalog queries if needed.
2. Add more explicit large-result and timeout scenarios.
3. Use this matrix during implementation to ensure test alignment.

This traceability matrix can be maintained as implementation progresses.