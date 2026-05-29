import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// ============================================================
// Configuration
// ============================================================
const MCP_BASE_URL = __ENV.MCP_BASE_URL || 'http://localhost:8000';
const MCP_ENDPOINT = __ENV.MCP_ENDPOINT || `${MCP_BASE_URL}/mcp`;

const VUS = parseInt(__ENV.VUS || '200');
const DURATION = __ENV.DURATION || '5m';
const RAMP_UP = __ENV.RAMP_UP || '1m';

const CHAOS_MODE = __ENV.CHAOS_MODE === 'true'; // Enable with CHAOS_MODE=true

// ============================================================
// Custom Metrics
// ============================================================
const toolLatency = new Trend('tool_latency', true);
const toolCalls = new Counter('tool_calls_total');
const validationRejections = new Counter('validation_rejections_total');
const successfulCalls = new Rate('successful_calls');
const chaosInjections = new Counter('chaos_injections_total');

// ============================================================
// k6 Options
// ============================================================
export const options = {
  stages: [
    { duration: RAMP_UP, target: VUS },
    { duration: DURATION, target: VUS },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    'http_req_duration': ['p(95)<800'],
    'tool_latency{tool:execute_safe_read_query}': ['p(99)<25'],
    'checks': ['rate>0.85'],
  },
};

// ============================================================
// Data
// ============================================================
const sampleQueries = [
  "SELECT table_name FROM information_schema.tables LIMIT 10",
  "SELECT * FROM rw_tables LIMIT 5",
  "SELECT actor_id FROM rw_actors LIMIT 10",
];

const dangerousQueries = [
  'DROP TABLE orders', 'DELETE FROM orders', 'INSERT INTO orders VALUES (1)',
  'UPDATE orders SET x=1', 'CREATE MATERIALIZED VIEW bad AS SELECT 1',
  'ALTER TABLE orders ADD hack INT', 'TRUNCATE TABLE orders',
];

const malformedPayloads = [
  { jsonrpc: "2.0", method: "tools/call" }, // missing params
  { jsonrpc: "1.0", id: 1, method: "tools/call", params: {} }, // wrong version
  { method: "tools/call" }, // invalid JSON-RPC
];

// ============================================================
// Helpers
// ============================================================
function callTool(toolName, args = {}) {
  const payload = { jsonrpc: "2.0", id: Date.now(), method: "tools/call", params: { name: toolName, arguments: args } };
  const res = http.post(MCP_ENDPOINT, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
    tags: { tool: toolName },
  });

  toolLatency.add(res.timings.duration, { tool: toolName });
  toolCalls.add(1, { tool: toolName });

  const success = check(res, { [`${toolName} status 200`]: r => r.status === 200 });
  successfulCalls.add(success);

  return res;
}

function injectChaos() {
  chaosInjections.add(1);

  // 1. Send dangerous query
  if (Math.random() < 0.6) {
    const badQuery = dangerousQueries[Math.floor(Math.random() * dangerousQueries.length)];
    callTool('execute_safe_read_query', { query: badQuery });
  }

  // 2. Send malformed JSON-RPC
  if (Math.random() < 0.3) {
    const badPayload = malformedPayloads[Math.floor(Math.random() * malformedPayloads.length)];
    http.post(MCP_ENDPOINT, JSON.stringify(badPayload), {
      headers: { 'Content-Type': 'application/json' },
      tags: { tool: 'chaos_malformed' },
    });
  }

  // 3. Rapid invalid tool call
  if (Math.random() < 0.2) {
    callTool('non_existent_tool', { foo: 'bar' });
  }
}

// ============================================================
// Main
// ============================================================
export default function () {
  const rand = Math.random();

  // === Normal Performance Path ===
  if (!CHAOS_MODE) {
    if (rand < 0.4) {
      callTool('execute_safe_read_query', { query: sampleQueries[Math.floor(Math.random() * sampleQueries.length)] });
    } else if (rand < 0.6) {
      callTool('show_tables', {});
    } else {
      callTool('list_streaming_jobs', {});
    }
  }

  // === Chaos Injection Mode ===
  if (CHAOS_MODE && Math.random() < 0.7) {
    injectChaos();
  }

  // === Functional Coverage (always run) ===
  if (Math.random() < 0.12) callTool('describe_table', { table_name: 'orders' });
  if (Math.random() < 0.08) callTool('list_materialized_views', {});

  if (!CHAOS_MODE && Math.random() < 0.10) {
    const bad = dangerousQueries[Math.floor(Math.random() * dangerousQueries.length)];
    callTool('execute_safe_read_query', { query: bad });
  }

  if (Math.random() < 0.05) {
    http.post(MCP_ENDPOINT, JSON.stringify({ jsonrpc: "2.0", id: Date.now(), method: "tools/list" }), {
      headers: { 'Content-Type': 'application/json' }, tags: { tool: 'tools_list' }
    });
  }

  sleep(CHAOS_MODE ? 0.1 : 0.3);
}

// ============================================================
// Summary
// ============================================================
export function handleSummary(data) {
  console.log('\n=== k6 Load Test Summary ===');
  if (CHAOS_MODE) {
    console.log('CHAOS MODE ENABLED — Error injection active');
    console.log(`Chaos injections: ${data.metrics.chaos_injections_total?.values.count || 0}`);
  }
  console.log('Run with CHAOS_MODE=true for resilience testing.');
}