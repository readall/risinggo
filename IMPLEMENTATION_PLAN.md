# RisingWave MCP Server (Go) — Implementation Plan

**Version**: 1.0  
**Date**: May 2026  
**Status**: Ready for Execution

---

## 1. Overview

This document provides a **detailed, step-by-step implementation plan** for building the Go-based RisingWave MCP Server according to the specifications defined in the Gherkin feature files and supporting Markdown documents.

The plan is broken down into small, verifiable units called **"Beads"**. Each bead represents an atomic unit of work with clear tasks, acceptance criteria, and validation methods.

### Goals

- Deliver a **strictly read-only**, high-performance MCP server
- Achieve **p99 latency < 20ms** (with 10% variance tolerance)
- Ensure strong **safety and error handling**
- Maintain high **traceability** to Gherkin specifications
- Enable **production-ready observability**

---

## 2. Guiding Principles

| Principle          | Description |
|--------------------|-------------|
| **Safety First**   | Strictly read-only behavior must be enforced from day one |
| **Test-Driven**    | Every bead must include corresponding tests |
| **Specification-Driven** | All work must be traceable to Gherkin scenarios |
| **Incremental**    | Deliver working, testable functionality in small steps |
| **Observability by Default** | Metrics, logs, and tracing must be built in early |

---

## 3. Implementation Phases

### Phase 0: Foundations & Safety Core (Week 1)

**Objective**: Establish a secure and solid foundation.

| Bead | Task | Validation Criteria | Gherkin / Design Reference | Priority |
|------|------|---------------------|----------------------------|----------|
| 0.1 | Project initialization & structure | Follows recommended layout | Design doc | High |
| 0.2 | Configuration system (env + validation) | Fails fast on invalid config | `technical.features` | High |
| 0.3 | `pgxpool` database access layer | Proper pooling, health checks, context support | ADR-003 | **Critical** |
| 0.4 | Multi-layer Safety & Validation system | Rejects dangerous queries at multiple layers | `invariants_and_non_functional.features` | **Critical** |
| 0.5 | `validateReadOnlyQuery()` + resource limits | Enforces query type, timeout, and row limits | Safety scenarios | **Critical** |
| 0.6 | Basic MCP server bootstrap (stdio + HTTP) | Responds to `tools/list` | MCP Protocol | High |
| 0.7 | Tool registration with metadata (`ReadOnlyHint`) | All tools declare `ReadOnlyHint: true` | Design + MCP best practices | High |
| 0.8 | Unit tests for Safety Layer | All dangerous queries rejected in tests | - | High |

**Phase 0 Validation Gate**: Server starts and **provably cannot** execute any mutating query.

---

### Phase 1: Core Read-Only Functionality (Weeks 2–3)

**Objective**: Deliver the most frequently used read-only tools.

| Bead | Task | Validation Criteria | Gherkin Reference | Priority |
|------|------|---------------------|-------------------|----------|
| 1.1 | `execute_safe_read_query` tool | Executes valid queries, rejects invalid ones | `public_api.features` | **Critical** |
| 1.2 | `show_tables` tool | Returns accurate table list | Schema scenarios | High |
| 1.3 | `describe_table` tool | Returns correct column information | Schema scenarios | High |
| 1.4 | `list_materialized_views` tool | Returns correct materialized views | Schema scenarios | Medium |
| 1.5 | Basic Explain tools | Returns query execution plans | Explain scenarios | Medium |
| 1.6 | Structured result formatting | Returns clean JSON/text results | Agent UX | High |
| 1.7 | Input validation & identifier sanitization | Rejects invalid identifiers | Safety scenarios | High |
| 1.8 | Integration tests with testcontainers | Tools work against real RisingWave | - | High |

**Phase 1 Validation Gate**: Core query and schema tools are functional and remain strictly read-only.

---

### Phase 2: Monitoring & Advanced Tools (Weeks 4–5)

**Objective**: Implement streaming monitoring and storage analysis capabilities.

| Bead | Task | Validation Criteria | Gherkin Reference | Priority |
|------|------|---------------------|-------------------|----------|
| 2.1 | `list_streaming_jobs` tool | Returns active streaming jobs | Monitoring scenarios | High |
| 2.2 | Backfill progress monitoring | Returns accurate backfill status | Monitoring scenarios | High |
| 2.3 | Fragment & Actor monitoring tools | Returns fragment and actor details | Technical scenarios | Medium |
| 2.4 | Storage / Hummock tools | Returns compaction and storage statistics | Storage scenarios | Medium |
| 2.5 | Cluster & version info tools | Returns cluster metadata | Cluster scenarios | Medium |
| 2.6 | Additional schema tools | `show_create_table`, `get_table_stats` | Schema scenarios | Medium |
| 2.7 | Expand integration test coverage | Covers tools from Phase 1 & 2 | - | High |

**Phase 2 Validation Gate**: Monitoring and storage tools function correctly.

---

### Phase 3: Production Readiness & Observability (Weeks 6–7)

**Objective**: Make the server production-grade.

| Bead | Task | Validation Criteria | Gherkin Reference | Priority |
|------|------|---------------------|-------------------|----------|
| 3.1 | Structured logging + correlation IDs | Every request is traceable | Observability scenarios | High |
| 3.2 | Prometheus metrics (latency, rejections, pool stats) | Metrics endpoint exposed and useful | Observability scenarios | High |
| 3.3 | OpenTelemetry tracing | Full request lifecycle spans available | Observability scenarios | Medium |
| 3.4 | Health & Readiness endpoints | `/healthz` and `/readyz` implemented | Technical scenarios | High |
| 3.5 | Graceful shutdown handling | Handles in-flight requests cleanly | Recovery scenarios | High |
| 3.6 | Authentication middleware (configurable) | Supports API Key (optional) | Security scenarios | Medium |
| 3.7 | Rate limiting | Protects against abuse | Security scenarios | Medium |
| 3.8 | Error classification & rich error responses | Clear, actionable errors returned | Error handling scenarios | High |

**Phase 3 Validation Gate**: Server has production-grade observability and reliability.

---

### Phase 4: Performance, Resilience & Validation (Week 8)

**Objective**: Validate performance targets and system resilience.

| Bead | Task | Validation Criteria | Gherkin Reference | Priority |
|------|------|---------------------|-------------------|----------|
| 4.1 | Latency optimization of hot paths | Measurable p99 improvement | Performance scenarios | High |
| 4.2 | Baseline k6 load test (Normal mode) | Establishes performance baseline | `invariants_and_non_functional.features` | **Critical** |
| 4.3 | k6 load test with 200 concurrent users | Maintains stability and performance | Performance scenarios | **Critical** |
| 4.4 | k6 test in **Chaos Mode** | Safety layers remain effective | Chaos + Safety scenarios | High |
| 4.5 | Latency breakdown instrumentation | Can measure validation vs DB execution time | Observability + Performance | High |
| 4.6 | Validate performance targets | p99 < 20ms (with 10% variance) achieved | Design doc | **Critical** |
| 4.7 | Expand Go test coverage | Maintain high unit + integration coverage | - | High |

**Phase 4 Validation Gate**: Performance targets are met and resilience is validated under chaos conditions.

---

### Phase 5: Documentation, Compliance & Release (Weeks 9–10)

**Objective**: Finalize, document, and release.

| Bead | Task | Validation Criteria | Priority |
|------|------|---------------------|----------|
| 5.1 | Maintain traceability matrix | All implemented features mapped to Gherkin | High |
| 5.2 | Achieve target test coverage | ≥ 80% coverage on core packages | High |
| 5.3 | Final review against all Gherkin specs | Critical scenarios pass | **Critical** |
| 5.4 | Docker multi-stage build + hardening | Small, secure, non-root image | High |
| 5.5 | Update all project documentation | README, design doc, and specs aligned | Medium |
| 5.6 | Create usage examples for AI agents | Example prompts and workflows documented | Medium |
| 5.7 | Prepare v0.1.0 release | Tagged release with changelog | High |

---

## 4. Cross-Cutting Concerns (Ongoing)

| Area              | Responsibility                              | Validation Method                     |
|-------------------|---------------------------------------------|---------------------------------------|
| **Safety**        | Continuous review of rejection logic        | All dangerous queries blocked in tests |
| **Observability** | Metrics & tracing for every new feature     | Dashboards reflect new capabilities   |
| **Traceability**  | Keep `original-behavior-coverage.md` updated | Matrix reflects current implementation |
| **Performance**   | Regular k6 runs during development          | No regression in p99 latency          |
| **Testing**       | Maintain alignment with Gherkin             | Tests map to specification scenarios  |

---

## 5. Recommended Execution Strategy

| Phase     | Focus                        | Suggested Duration | Risk if Delayed          |
|-----------|------------------------------|--------------------|--------------------------|
| **Phase 0** | Safety + Foundations        | 1 week            | High (Security debt)    |
| **Phase 1** | Core Functionality          | 2 weeks           | Medium                  |
| **Phase 3** | Observability               | Parallel with Phase 2 | Medium               |
| **Phase 2** | Monitoring Tools            | 2 weeks           | Low                     |
| **Phase 4** | Performance & Chaos         | 1 week            | **High** (Target not validated) |
| **Phase 5** | Release & Documentation     | 1–2 weeks         | Medium                  |

---

## 6. Summary

This plan provides a structured, traceable path to deliver a secure, high-performance, and observable RisingWave MCP Server. By following the bead-by-bead approach, the team can maintain high confidence in both correctness and alignment with the specifications.

**Key Success Factors**:
- Strict adherence to the read-only invariant
- Early and continuous investment in observability
- Regular validation against Gherkin specifications
- Use of the k6 harness (including Chaos Mode) for performance and resilience testing

---

*This document should be treated as a living plan and updated as implementation progresses.*