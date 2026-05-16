# RisingWave MCP Server (Go) — Specification Suite Overview

**Version**: 1.0  
**Date**: May 2026  
**Purpose**: Exhaustive, verifiable BDD-style executable specifications for the strictly read-only, high-performance Go MCP server for RisingWave.  
**Target Characteristics**: p99 latency < 20ms, zero mutation capability, very detailed observability, powerful safe generic read-only query executor, support for 200+ concurrent sessions.

## 1. Scope of This Specification Suite

This suite covers the **complete expected behavior** of the Go implementation of the RisingWave MCP Server as defined in the design document.

It is intentionally **stricter** than the original Python reference:
- **Zero tolerance for mutation** — no DDL, DML, or destructive operations are possible.
- **Aggressive performance target** — p99 tool latency < 20ms.
- **Multi-layer safety** around the generic read-only query tool.
- **Production-grade observability** by default.

## 2. Specification Structure

| File | Focus | Primary Audience | Verifiability Approach |
|------|-------|------------------|------------------------|
| `public_api.features` | External MCP tool surface, tool discovery, calling behavior from AI agents | Agent developers, product | MCP Inspector + custom test client |
| `technical.features` | Internal mechanics (pooling, validation pipeline, query execution, transports) | Engineers | Unit + integration tests with testcontainers |
| `persistence_and_recovery.features` | Connection recovery, graceful shutdown, config handling, failure resilience | SRE / Platform | Chaos + integration tests |
| `invariants_and_non_functional.features` | Invariants (read-only forever), performance (p99), scalability, security, observability | Architects + QA | Performance tests, property-based tests, static analysis |
| `implementation-constraints.md` | Technology choices, ADRs, what must/must-not be implemented | Implementers | Code review + automated checks (linters, dependency pins) |

## 3. Key Architectural Decision Records (ADRs) Embedded

All major decisions are captured as lightweight ADRs in `implementation-constraints.md` and referenced from scenarios where relevant.

**Core ADRs**:
- **ADR-001**: Strictly read-only by architectural design (no mutations ever)
- **ADR-002**: Use official `modelcontextprotocol/go-sdk`
- **ADR-003**: `pgx/v5` + `pgxpool` as the sole database access layer
- **ADR-004**: Multi-layer defense-in-depth for the generic safe read-only query executor
- **ADR-005**: p99 < 20ms as primary non-functional performance target with specific optimization techniques
- **ADR-006**: Very detailed observability (tracing + per-layer metrics + structured logging) enabled by default
- **ADR-007**: Dual transport support (stdio + Streamable HTTP) with preference for HTTP in scaled deployments
- **ADR-008**: Stateless core + connection pooling for horizontal scalability to 200+ sessions

## 4. How to Verify These Specifications

1. **Unit & Integration Tests** — Map `Given/When/Then` to Go test functions + `testcontainers-go`.
2. **MCP Protocol Compliance** — Use official MCP Inspector or custom Go client.
3. **Performance Verification** — Dedicated load tests measuring p99 latency under 200 concurrent sessions.
4. **Safety Verification** — Attempt mutation queries through the generic executor and assert rejection at every layer.
5. **Observability Verification** — Assert presence of required metrics, spans, and log fields.

All scenarios are written to be **executable** with minimal custom step definitions.

## 5. Coverage Goals

- **Functional**: 100% of public tools + generic executor behavior.
- **Safety**: Every rejection path in the validation pipeline has a scenario.
- **Performance**: Explicit scenarios for latency target and breakdown measurement.
- **Invariants**: Read-only posture is asserted in multiple independent ways.
- **Observability**: Every major component emits the required signals.

## 6. Relationship to Design Document

These specifications are the **verifiable contract** derived from `risingwave-mcp-go-design-plan.md`. Any implementation must pass this suite to be considered compliant.

---

**This specification suite is designed to be living documentation.** Update scenarios as the implementation or requirements evolve.