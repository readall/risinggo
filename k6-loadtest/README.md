# k6 Load Testing Harness for RisingWave MCP Server

This is a **production-oriented load testing harness** designed to validate the RisingWave MCP Server against the Gherkin specifications.

## Goals

- Validate **p99 latency target** (< 20ms with 10% variance)
- Test under **200+ concurrent users**
- Exercise **functional coverage** of the public API
- Stress **safety and rejection layers**
- Support **chaos / resilience testing**

## Current Coverage

| Area                              | Coverage | Notes |
|-----------------------------------|----------|-------|
| Generic Safe Query Executor       | High     | Primary focus |
| Schema Tools (`show_tables`, `describe_table`, etc.) | Good | Multiple tools |
| Monitoring Tools                  | Good     | Streaming jobs, backfill, storage |
| Mutation / Safety Rejection       | High     | Expanded dangerous queries + malformed requests |
| Tool Discovery                    | Medium   | Regularly exercised |
| Concurrency (200 VUs)             | High     | Core load scenario |
| p99 Latency Measurement           | High     | Dedicated thresholds |
| Chaos / Error Injection           | High     | Enabled via `CHAOS_MODE=true` |

The harness aims for **broad functional + performance coverage** while keeping performance-critical paths clean.

## Usage

### Normal Mode (Performance + Functional)

```bash
k6 run mcp-loadtest.js
```

### Chaos / Resilience Mode

```bash
k6 run -e CHAOS_MODE=true mcp-loadtest.js
```

**Chaos mode** aggressively injects:
- Dangerous/mutating queries
- Malformed JSON-RPC payloads
- Calls to non-existent tools

Use this mode to test safety layers and error handling robustness.

### Custom Load

```bash
k6 run -e VUS=300 -e DURATION=10m mcp-loadtest.js
```

## Metrics & Thresholds

| Metric                        | Description |
|-------------------------------|-------------|
| `tool_latency{tool:xxx}`      | Per-tool latency (p99 tracked) |
| `validation_rejections_total` | Counts rejected dangerous queries |
| `chaos_injections_total`      | Number of chaos actions (only in chaos mode) |
| `successful_calls`            | Success rate of tool calls |

**Key Thresholds**:
- `execute_safe_read_query` p99 < 22ms
- Overall check success rate > 85%

## How to Add / Enhance Coverage

### 1. Add New Tools

Edit the `default` function and add calls:

```js
if (Math.random() < 0.10) {
  callTool('your_new_tool', { param: 'value' });
}
```

### 2. Expand Rejection Testing

Add more dangerous queries to the `dangerousQueries` array:

```js
const dangerousQueries = [
  'DROP TABLE ...',
  'Your new dangerous query here',
];
```

### 3. Add Malformed Payloads

Extend `malformedPayloads` for chaos mode:

```js
const malformedPayloads = [
  { jsonrpc: "2.0", method: "tools/call" }, // missing params
  // Add more invalid structures
];
```

### 4. Create Specialized Scenarios

You can create new functions for specific coverage areas:

```js
function testLargeResultQueries() {
  // Add logic for large result testing
}
```

### 5. Adjust Load Profile

Modify the `options.stages` or use environment variables:

```bash
k6 run -e VUS=500 -e DURATION=15m mcp-loadtest.js
```

## Recommended Workflow

1. Run normal mode first to establish baseline performance.
2. Run chaos mode to validate safety and error handling.
3. Monitor `validation_rejections_total` and p99 latency.
4. Extend coverage by adding new tools or rejection cases as the server implementation grows.

## Files

- `mcp-loadtest.js` — Main test script
- `docker-compose.example.yml` — Example environment with RisingWave

---

**This harness is designed to evolve alongside the Gherkin specifications and implementation.**