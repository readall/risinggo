# risinggo

**Independent Go implementation of a RisingWave MCP Server**

> **Disclaimer**: This project (`readall/risinggo`) is an **independent implementation** created for educational and production use. It is **not affiliated with, endorsed by, or a derivative work** of any repository under the `risingwavelabs` organization.

It draws high-level inspiration from MCP tooling patterns and the official Model Context Protocol, but all code, design, and specifications in this repository are original.

**Key Characteristics**:
- Strictly read-only (zero mutation capability)
- p99 latency target < 20ms (with 10% variance)
- Powerful generic safe read-only query executor
- Very detailed observability
- Comprehensive Gherkin specification suite
- k6 load testing harness for 200 concurrent users

## Directory Structure

- `EXAMPLES.md` - Concrete usage examples for AI agents (MCP clients)
- `design/` - Detailed design document
- `specs/` - Executable Gherkin specifications
- `IMPLEMENTATION_PLAN.md` - Bead-by-bead execution plan + current status
- `k6-loadtest/` - Load testing harness (validated)
- `tests/` - Go tests (unit + integration with testcontainers)

## Official RisingWave

This project integrates with the official [RisingWave](https://github.com/risingwavelabs/risingwave) database (retained as the upstream project). All references to `risingwavelabs/risingwave` are for compatibility with the official project only.

## For AI Agents

See [EXAMPLES.md](EXAMPLES.md) for concrete prompts, tool usage patterns, safety guarantees, and end-to-end workflows that work great with Claude, Cursor, and other MCP clients.

All tool responses are returned as clean human-readable text tables for the best possible agent experience.

## Current Status (as of 2026-05-30)

**Phase 1 Complete** (Core read-only functionality + safety + agent UX)

Implemented and production-validated:
- `execute_safe_read_query` — full SQL power with strict read-only enforcement
- `show_tables`, `describe_table` — schema discovery
- `list_streaming_jobs` — RisingWave-specific monitoring
- Structured result formatting (beautiful text tables for agents + JSON for perf clients)
- Strong identifier sanitization + multi-layer safety
- Integration test infrastructure (testcontainers + real DB)
- `/mcp-raw` direct shim for load testing
- Comprehensive agent usage examples

**Current Focus**: Phase 2 — Monitoring & Advanced Tools (P2 beads)

See `IMPLEMENTATION_PLAN.md` for the full bead-by-bead breakdown and `bd ready` for live task list.

**Quality Gates Passed**:
- k6 load test (P0) validated at high concurrency
- All P1 tasks closed
- Strict bd (beads) tracking + mandatory push discipline

The server is already usable and safe for real AI agent workloads against RisingWave.