# k6 Load Testing Harness for RisingWave MCP Server (Go)

This harness validates the **RisingWave MCP Server** against the specification requirements, with special focus on:

- **p99 latency target < 20ms** with up to **10% variance** (acceptable up to ~22ms)
- Support for **200+ concurrent users/sessions**
- **Strictly read-only** behavior
- Performance of the **powerful generic safe read-only query executor**
- Overall stability under sustained load

## Features of This Harness

- Uses **k6** (Grafana k6) for modern, scriptable load testing.
- Targets the **Streamable HTTP** transport (`/mcp` endpoint).
- Simulates realistic AI agent behavior with a mix of tools.
- Heavy emphasis on `execute_safe_read_query` (the most powerful read-only tool).
- Custom metrics and per-tool p99 tracking.
- Configurable ramp-up to 200 virtual users (VUs).
- Thresholds aligned with the specification (p99 target + 10% variance).
- Easy to extend with more queries or tool calls.

## Prerequisites

- [k6](https://k6.io/docs/getting-started/installation/) installed
- RisingWave MCP Server running with `MCP_TRANSPORT=streamable-http`
- A running RisingWave instance with some sample data (recommended: at least one table like `orders`)

## Quick Start

### 1. Start the MCP Server (HTTP mode)

```bash
docker run -p 8000:8000 \
  -e MCP_TRANSPORT=streamable-http \
  -e MCP_PORT=8000 \
  -e RISINGWAVE_CONNECTION_STR="postgresql://root:root@host.docker.internal:4566/dev" \
  risingwavelabs/risingwave-mcp-server:latest   # or your built Go image
```

### 2. Run the Load Test (default: 200 VUs)

```bash
k6 run mcp-loadtest.js
```

### 3. Run with custom parameters

```bash
k6 run \
  -e MCP_BASE_URL=http://localhost:8000 \
  -e VUS=200 \
  -e DURATION=10m \
  -e RAMP_UP=2m \
  mcp-loadtest.js
```

## Environment Variables

| Variable          | Default     | Description |
|-------------------|-------------|-----------|
| `MCP_BASE_URL`    | `http://localhost:8000` | Base URL of the MCP server |
| `VUS`             | `200`       | Number of concurrent virtual users |
| `DURATION`        | `5m`        | How long to sustain the target load |
| `RAMP_UP`         | `1m`        | Time to ramp up to target VUs |
| `RAMP_DOWN`       | `30s`       | Time to ramp down after test |

## What the Test Does

The script is intentionally split into two sections for clarity:

### 1. Performance & P99 Load Scenarios (Primary Focus)
- Heavy weighting toward `execute_safe_read_query` and key monitoring tools.
- These drive the p99 measurements and thresholds.
- Designed to validate the **< 20ms target with 10% variance**.

### 2. Functional Coverage Scenarios (Complementary)
- Additional calls for broader test coverage:
  - Deeper schema inspection
  - Explicit testing of mutation rejection in the generic executor
  - Tool discovery
- These improve overall specification coverage without distorting core performance metrics.

This separation makes it easy to analyze pure performance vs. functional correctness.

## Interpreting Results Against the Specification

### Success Criteria (aligned with spec)

- **p99 latency** for key tools (especially `execute_safe_read_query`) should **target < 20ms**.
- Up to **10% variance** is acceptable → observed p99 up to **~22ms** is within tolerance.
- The test uses thresholds like:
  ```js
  'tool_latency{tool:execute_safe_read_query}': ['p(99)<22']
  ```
- High check success rate (>95%).
- No unbounded latency growth during sustained load.

### Key Metrics to Watch

- `tool_latency{tool:xxx}` → p(99) per tool
- `http_req_duration` → overall request timing
- Custom counters for tool calls and validation rejections

k6 will automatically fail the test if thresholds are breached.

## Extending the Test

### Add more sample queries

Edit the `sampleQueries` array in `mcp-loadtest.js`.

### Add new tools

Add more `else if` branches in the `default` function and call `callTool('tool_name', args)`.

### Run with Prometheus output (for dashboards)

```bash
k6 run --out prometheus mcp-loadtest.js
```

Then visualize in Grafana using the official k6 dashboard.

## Mapping to Specification

This harness directly supports verification of:

- `invariants_and_non_functional.features` → Performance target scenarios
- `technical.features` → Latency breakdown and pool behavior under load
- `public_api.features` → Concurrent tool usage and read-only behavior

## Recommended Workflow

1. Run short test (1-2 min) to validate setup.
2. Run full test with 200 VUs for 5-10 minutes.
3. Review p99 values per tool.
4. If p99 is significantly above 22ms, investigate using the server's detailed observability (tracing + per-layer metrics) to identify bottlenecks.
5. Iterate on server optimizations (pool tuning, caching, query paths).

## Notes & Limitations

- This test focuses on the **HTTP transport**. stdio mode is harder to load test at this scale.
- Assumes the MCP server is already running and connected to RisingWave.
- For very accurate p99 measurement, run longer tests (10m+) in a stable environment.
- The generic query executor is intentionally stressed the most, as it is the most powerful read-only capability.

---

**This harness is designed to be production-grade and directly tied to the Gherkin specifications.** Use it to continuously validate that your implementation meets the p99 target with acceptable variance under realistic agent-like load.