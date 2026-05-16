import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// ============================================================
// Configuration
// ============================================================
const MCP_BASE_URL = __ENV.MCP_BASE_URL || 'http://localhost:8000';
const MCP_ENDPOINT = `${MCP_BASE_URL}/mcp`;

const VUS = parseInt(__ENV.VUS || '200');
const DURATION = __ENV.DURATION || '5m';
const RAMP_UP = __ENV.RAMP_UP || '1m';

// ============================================================
// Custom Metrics
// ============================================================
const toolLatency = new Trend('tool_latency', true);
const toolCalls = new Counter('tool_calls_total');
const validationRejections = new Counter('validation_rejections_total');
const successfulCalls = new Rate('successful_calls');

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
    'http_req_duration': ['p(95)<500'],
    'tool_latency{tool:execute_safe_read_query}': ['p(99)<22'],
    'tool_latency{tool:show_tables}': ['p(99)<22'],
    'checks': ['rate>0.90'],
  },
  summaryTrendStats: ['avg', 'p(95)', 'p(99)'],
};

// ============================================================
// Sample Data
// ============================================================
const sampleQueries = [
  "SELECT table_name, table_type FROM information_schema.tables LIMIT 20",
  "SELECT * FROM rw_tables LIMIT 10",
  "SELECT actor_id, status FROM rw_actors LIMIT 15",
  "SELECT fragment_id, state FROM rw_fragments LIMIT 10",
  "SELECT name FROM rw_materialized_views LIMIT 10",
];

const dangerousQueries = [
  'DROP TABLE orders',
  'DELETE FROM orders WHERE id = 1',
  'INSERT INTO orders VALUES (1, 100)',
  'UPDATE orders SET amount = 0',
  'CREATE MATERIALIZED VIEW malicious AS SELECT 1',
  'ALTER TABLE orders ADD COLUMN hack INT',
  'GRANT ALL PRIVILEGES ON orders TO public',
  'TRUNCATE TABLE orders',
];

// ============================================================
// Helper Functions
// ============================================================
function callTool(toolName, args = {}) {
  const payload = {
    jsonrpc: "2.0",
    id: Date.now(),
    method: "tools/call",
    params: { name: toolName, arguments: args },
  };

  const res = http.post(MCP_ENDPOINT, JSON.stringify(payload), {
    headers: { 'Content-Type': 'application/json' },
    tags: { tool: toolName },
  });

  const duration = res.timings.duration;
  toolLatency.add(duration, { tool: toolName });
  toolCalls.add(1, { tool: toolName });

  const success = check(res, {
    [`${toolName} returns 200`]: (r) => r.status === 200,
  });

  successfulCalls.add(success);

  // Track rejections for safety coverage
  if (toolName === 'execute_safe_read_query' && res.status === 200) {
    try {
      const body = JSON.parse(res.body);
      if (body.error) {
        validationRejections.add(1);
      }
    } catch (e) {}
  }

  return res;
}

// ============================================================
// Main Test Function
// ============================================================
export default function () {
  // === PERFORMANCE & P99 SCENARIOS ===
  const rand = Math.random();

  if (rand < 0.35) {
    const query = sampleQueries[Math.floor(Math.random() * sampleQueries.length)];
    callTool('execute_safe_read_query', { query });
  } 
  else if (rand < 0.50) {
    callTool('show_tables', {});
  } 
  else if (rand < 0.65) {
    callTool('list_streaming_jobs', {});
  } 
  else {
    callTool('get_cluster_info', {});
  }

  // === FUNCTIONAL COVERAGE SCENARIOS (Expanded) ===

  // Schema Tools Coverage
  if (Math.random() < 0.15) callTool('describe_table', { table_name: 'orders' });
  if (Math.random() < 0.10) callTool('list_materialized_views', {});
  if (Math.random() < 0.08) callTool('show_create_table', { table_name: 'orders' });

  // Monitoring & Storage Tools
  if (Math.random() < 0.10) callTool('get_backfill_progress', {});
  if (Math.random() < 0.08) callTool('get_hummock_stats', {});

  // Rejection Testing (Safety Coverage)
  if (Math.random() < 0.12) {
    const badQuery = dangerousQueries[Math.floor(Math.random() * dangerousQueries.length)];
    callTool('execute_safe_read_query', { query: badQuery });
  }

  // Tool Discovery
  if (Math.random() < 0.06) {
    http.post(MCP_ENDPOINT, JSON.stringify({
      jsonrpc: "2.0", id: Date.now(), method: "tools/list"
    }), {
      headers: { 'Content-Type': 'application/json' },
      tags: { tool: 'tools_list' },
    });
  }

  sleep(Math.random() * 0.3 + 0.1);
}

// ============================================================
// Summary
// ============================================================
export function handleSummary(data) {
  console.log('\n=== Enhanced k6 Load Test Summary ===');
  console.log('Performance + Functional Coverage scenarios executed.');
  console.log('See thresholds and metrics for p99 and rejection rates.\n');
}