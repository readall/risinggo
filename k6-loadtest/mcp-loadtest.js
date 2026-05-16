import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';

// ============================================================
// Configuration via environment variables
// ============================================================
const MCP_BASE_URL = __ENV.MCP_BASE_URL || 'http://localhost:8000';
const MCP_ENDPOINT = `${MCP_BASE_URL}/mcp`;

const VUS = parseInt(__ENV.VUS || '200');
const DURATION = __ENV.DURATION || '5m';
const RAMP_UP = __ENV.RAMP_UP || '1m';
const RAMP_DOWN = __ENV.RAMP_DOWN || '30s';

// Target p99 with 10% variance (in milliseconds)
const P99_TARGET_MS = 20;
const P99_WITH_VARIANCE_MS = 22;

// ============================================================
// Custom Metrics
// ============================================================
const toolLatency = new Trend('tool_latency', true);
const toolCalls = new Counter('tool_calls_total');
const validationRejections = new Counter('validation_rejections_total');

// ============================================================
// k6 Options - Ramp to 200 VUs
// ============================================================
export const options = {
  stages: [
    { duration: RAMP_UP, target: VUS },
    { duration: DURATION, target: VUS },
    { duration: RAMP_DOWN, target: 0 },
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500', 'p(99)<1000'],
    // === PERFORMANCE & P99 THRESHOLDS (keep separate) ===
    'tool_latency{tool:execute_safe_read_query}': [`p(99)<${P99_WITH_VARIANCE_MS}`],
    'tool_latency{tool:show_tables}': [`p(99)<${P99_WITH_VARIANCE_MS}`],
    'tool_latency{tool:describe_table}': [`p(99)<${P99_WITH_VARIANCE_MS}`],
    'tool_latency{tool:list_streaming_jobs}': [`p(99)<${P99_WITH_VARIANCE_MS}`],
    'checks': ['rate>0.95'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
};

// ============================================================
// Sample safe queries for the generic executor
// ============================================================
const sampleQueries = [
  "SELECT table_name, table_type FROM information_schema.tables LIMIT 20",
  "SELECT * FROM rw_tables LIMIT 10",
  "SELECT actor_id, status, parallelism FROM rw_actors LIMIT 15",
  "SELECT fragment_id, state, parallelism FROM rw_fragments LIMIT 10",
  "SELECT name, schema_name, definition FROM rw_materialized_views LIMIT 10",
  "SELECT * FROM rw_hummock_version LIMIT 5",
];

// ============================================================
// Helper: Call an MCP tool via Streamable HTTP
// ============================================================
function callTool(toolName, args = {}) {
  const payload = {
    jsonrpc: "2.0",
    id: Date.now(),
    method: "tools/call",
    params: {
      name: toolName,
      arguments: args,
    },
  };

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json, text/event-stream',
    },
    tags: { tool: toolName },
  };

  const start = Date.now();
  const res = http.post(MCP_ENDPOINT, JSON.stringify(payload), params);
  const duration = Date.now() - start;

  toolLatency.add(duration, { tool: toolName });
  toolCalls.add(1, { tool: toolName });

  const success = check(res, {
    [`${toolName} status is 200`]: (r) => r.status === 200,
    [`${toolName} has result or error`]: (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.result !== undefined || body.error !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  // Track validation rejections for the generic executor
  if (toolName === 'execute_safe_read_query' && res.status === 200) {
    try {
      const body = JSON.parse(res.body);
      if (body.error && body.error.message && body.error.message.toLowerCase().includes('reject')) {
        validationRejections.add(1);
      }
    } catch (e) {}
  }

  return { res, duration, success };
}

// ============================================================
// Main test logic
// ============================================================
export default function () {
  // ============================================================
  // === PERFORMANCE & P99 LOAD SCENARIOS (keep this section clean) ===
  // ============================================================
  // Heavy focus on generic executor + key monitoring tools for p99 measurement

  const rand = Math.random();

  if (rand < 0.40) {
    // Performance-critical path: Generic safe query executor
    const query = sampleQueries[Math.floor(Math.random() * sampleQueries.length)];
    callTool('execute_safe_read_query', { query: query });
  } 
  else if (rand < 0.60) {
    callTool('show_tables', {});
  } 
  else if (rand < 0.75) {
    callTool('list_streaming_jobs', {});
  } 
  else {
    callTool('get_cluster_info', {});
  }

  // ============================================================
  // === FUNCTIONAL COVERAGE SCENARIOS (separate from pure load) ===
  // ============================================================
  // These add broader test coverage without heavily impacting p99 thresholds

  // --- Schema Variations ---
  if (Math.random() < 0.12) {
    callTool('describe_table', { table_name: 'orders' });
  }
  if (Math.random() < 0.08) {
    callTool('list_materialized_views', {});
  }
  if (Math.random() < 0.06) {
    callTool('show_create_table', { table_name: 'orders' });
  }

  // --- Rejection Test Cases (Safety Coverage) ---
  const dangerousQueries = [
    'DROP TABLE orders',
    'DELETE FROM orders',
    'INSERT INTO orders VALUES (1)',
    'UPDATE orders SET amount = 0',
    'CREATE MATERIALIZED VIEW test AS SELECT 1',
    'ALTER TABLE orders ADD COLUMN new_col INT',
    'GRANT ALL ON orders TO public',
  ];

  if (Math.random() < 0.10) {
    const badQuery = dangerousQueries[Math.floor(Math.random() * dangerousQueries.length)];
    callTool('execute_safe_read_query', { query: badQuery });
  }

  // --- Tool Discovery ---
  if (Math.random() < 0.05) {
    const listPayload = {
      jsonrpc: "2.0",
      id: Date.now(),
      method: "tools/list",
    };
    http.post(MCP_ENDPOINT, JSON.stringify(listPayload), {
      headers: { 'Content-Type': 'application/json' },
      tags: { tool: 'tools_list' },
    });
  }

  sleep(Math.random() * 0.4 + 0.1);
}

// ============================================================
// Custom Summary (optional enhanced reporting)
// ============================================================
export function handleSummary(data) {
  const p99Overall = data.metrics.tool_latency ? data.metrics.tool_latency.values['p(99)'] : 'N/A';

  console.log('\n=== RisingWave MCP Load Test Summary ===');
  console.log(`Target p99: < ${P99_TARGET_MS}ms (acceptable up to ${P99_WITH_VARIANCE_MS}ms with 10% variance)`);
  console.log(`Observed overall p99: ${p99Overall} ms`);
  console.log('Thresholds configured with variance tolerance.');
  console.log('See full k6 summary above for per-tool breakdowns.\n');

  return {
    'stdout': JSON.stringify(data, null, 2),
  };
}