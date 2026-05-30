# Usage Examples for AI Agents

This document shows how AI agents (Claude, Cursor, Continue, etc.) can use the RisingWave MCP Server (`risinggo`) to safely explore and query a RisingWave cluster.

All examples return **clean, human-readable text tables** (thanks to `rowsToText`).

## Running the Server

```bash
# From source
go run ./cmd/mcp

# With Docker (after building the image)
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgresql://root:root@host.docker.internal:4566/dev" \
  ghcr.io/readall/risinggo:latest
```

The server listens on `/mcp` (streamable HTTP) and `/mcp-raw` (direct JSON for load testing).

## Connecting an Agent

### Claude Desktop (example config)
```json
{
  "mcpServers": {
    "risingwave": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/inspector", "http://localhost:8080/mcp"]
    }
  }
}
```

Or use any MCP client that supports streamable HTTP.

## Available Tools

### 1. `execute_safe_read_query`
Run any read-only query (SELECT, WITH, EXPLAIN, SHOW, DESC).

**Example prompt for agent:**
> "Show me the top 5 materialized views by size in the dev database."

**What the agent actually calls:**
```json
{
  "name": "execute_safe_read_query",
  "arguments": {
    "query": "SELECT * FROM rw_catalog.materialized_views ORDER BY create_time DESC LIMIT 5"
  }
}
```

**Typical response (clean text table):**
```
name | schema_name | is_materialized | create_time
-----|-------------|-----------------|------------
mv_foo | public | true | 2026-05-20 10:11:22
...
(5 rows)
```

### 2. `show_tables`
List all user tables (excludes pg_catalog and information_schema).

**Example:**
> "What tables exist in the public schema?"

### 3. `describe_table`
Get column information for a specific table.

**Example:**
> "Describe the columns in the 'users' table"

**Arguments:**
```json
{
  "table_name": "users",
  "schema_name": "public"   // optional, defaults to public
}
```

### 4. `list_streaming_jobs`
List all materialized views and their status.

## Safety Guarantees (Important for Agents)

The server **rejects** any attempt to mutate data:

- INSERT, UPDATE, DELETE, DROP, CREATE, ALTER, etc. → rejected with clear reason
- Multiple statements (`; DROP ...`) → rejected
- SQL comments (`--` or `/*`) → rejected (prevents hidden mutations)
- Reserved schema prefixes (`pg_`, `information_schema`) in identifiers → rejected

Example rejection:
> "Query rejected: mutation keyword detected: DROP"

## Best Practices for Agents

1. **Always prefer `execute_safe_read_query`** for complex analysis — it supports full SQL power (CTEs, EXPLAIN, etc.).
2. Use `show_tables` + `describe_table` for discovery before writing queries.
3. When exploring large clusters, add `LIMIT` and `ORDER BY` in your queries.
4. The server enforces `MaxRows` (default 1000) — ask for more only if you really need it.
5. All results are truncated gracefully with row counts.

## Example End-to-End Workflow

Agent prompt:
> "I want to understand the streaming jobs in this RisingWave cluster. First show me the tables, then describe the most important one, then run a query that shows the 10 largest materialized views with their row counts."

The agent will naturally chain:
1. `show_tables`
2. `describe_table` on a promising table
3. `execute_safe_read_query` with a sophisticated SELECT against `rw_catalog`

## Raw Endpoint (for load testing / direct clients)

Use `/mcp-raw` when you want structured JSON instead of text (primarily for the k6 harness):

```bash
curl -X POST http://localhost:8080/mcp-raw \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"show_tables","arguments":{}}}'
```

## Related Documentation

- [README.md](README.md) — Project overview
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) — Current phase status
- [AGENTS.md](AGENTS.md) — Instructions for AI coding agents working on this repo
- [k6-loadtest/README.md](k6-loadtest/README.md) — Performance validation

This server is designed so agents can safely and effectively work with RisingWave without any risk of data mutation.
