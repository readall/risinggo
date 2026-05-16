# RisingWave MCP Server — Go Implementation
## Detailed Design Document & Implementation Plan

**Version**: 1.0 (Draft)  
**Date**: May 2026  
**Inspired by**: Original concepts from RisingWave MCP tools  
**Target**: Strictly read-only, high-performance MCP server (p99 < 20ms) for 200+ simultaneous users/sessions with zero mutation capability  
**Status**: Design complete; ready for implementation kickoff

---

## 1. Executive Summary

This document outlines the design and phased implementation plan for a high-performance, production-grade **Model Context Protocol (MCP)** server for RisingWave written in Go.

**Note**: This is an independent implementation. It draws inspiration from existing MCP tooling patterns for RisingWave but is not affiliated with or a derivative of any specific repository under risingwavelabs.

**The Go implementation** addresses these by:
- Using the **official `modelcontextprotocol/go-sdk`** for typed, schema-driven tool registration and dual transports (stdio + Streamable HTTP).
- Replacing the single-connection model with **`pgx` + `pgxpool`** for efficient, concurrent, pooled access to RisingWave (Postgres wire protocol).
- Embracing Go’s lightweight goroutines for true high-concurrency handling of 200+ simultaneous MCP sessions.
- Enforcing a **strictly read-only** posture: **no DDL, DML, or any mutating operations are possible** through the MCP server or agents. All tools are read-only by design.
- Targeting aggressive performance: **p99 tool latency < 20ms** under realistic load.
- Providing **very detailed observability** (per-tool metrics, query tracing, pool stats, validation overhead, etc.).
- Adding comprehensive production features from day one: structured observability, security model, resilience patterns, and multi-layer safeguards around the powerful generic **read-only** query executor.

**Expected outcomes**:
- Extremely low and predictable latency suitable for interactive AI agents.
- Native support for high concurrency (200+ sessions) with minimal resource usage.
- Single static binary, tiny Docker images, Kubernetes-native deployment.
- Maximum safety: agents cannot accidentally or maliciously modify schema or data.
- Rich observability for debugging agent behavior and performance tuning.
- Easier long-term maintenance and extension as RisingWave evolves.

---

## 2. Goals & Non-Goals

### Primary Goals
- Deliver a **strictly read-only** MCP server: **zero possibility of destruction or mutation** via MCP tools or AI agents (no DDL, DML, or any write operations).
- Provide a **powerful generic read-only query executor** with multiple layers of safeguards.
- Achieve **aggressive performance**: p99 end-to-end tool latency **< 20ms** under realistic concurrent load for 200 sessions.
- Deliver **very detailed observability** across every layer.
- Achieve **production readiness** for environments with up to 200 simultaneous MCP clients/sessions.

### Non-Goals (for v1)
- Full re-implementation of every edge-case from other MCP tools.
- Support for every experimental RisingWave feature on day one.

### Success Metrics (v1)
- **p99 end-to-end tool latency < 20ms** for common read operations under load from 200 concurrent sessions.
- Handles 200 concurrent long-lived MCP sessions with stable performance and no mutation capability.
- Memory footprint < 80MB idle + connection pool.

---

## 3. High-Level Architecture

(Architecture diagram remains the same)

---

## 4. Technology Stack

(Technology stack remains the same — using official Go MCP SDK and pgx)

---

## 5. Recommended Project Structure

(Structure remains the same)

---

## 6. Detailed Component Designs

(Design sections remain the same)

---

## 7. Production Readiness Features

(Observability, security sections remain the same)

---

## 8. Scalability Considerations for 200+ Simultaneous Users

(Scalability section remains the same)

---

## 9. Implementation Roadmap (Phased)

(Roadmap remains the same)

---

## 10. Testing Strategy

(Testing strategy remains the same)

---

## 11. Risks, Assumptions & Mitigations

(Risks section remains the same)

---

## 12. Future Roadmap (Post v1)

(Future roadmap remains the same)

---

## 13. Open Questions for the Team (Socratic Prompts)

(Open questions remain the same)

---

**This is an independent implementation** created under the `readall/risinggo` repository. It is not affiliated with risingwavelabs.