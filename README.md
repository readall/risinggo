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

- `design/` - Detailed design document
- `specs/` - Executable Gherkin specifications
- `k6-loadtest/` - Load testing harness
- `tests/` - Go test skeleton

## Official RisingWave

This project integrates with the official [RisingWave](https://github.com/risingwavelabs/risingwave) database (retained as the upstream project). All references to `risingwavelabs/risingwave` are for compatibility with the official project only.

## For AI Agents

See [EXAMPLES.md](EXAMPLES.md) for concrete prompts, tool usage patterns, safety guarantees, and end-to-end workflows that work great with Claude, Cursor, and other MCP clients.

All tool responses are returned as clean human-readable text tables for the best possible agent experience.