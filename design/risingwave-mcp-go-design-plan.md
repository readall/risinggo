# RisingWave MCP Server — Go Implementation
## Detailed Design Document & Implementation Plan

**Version**: 1.0 (Draft)  
**Date**: May 2026  
**Inspired by**: https://github.com/risingwavelabs/risingwave-mcp (Python/FastMCP + risingwave-py)  
**Target**: Strictly read-only, high-performance MCP server (p99 < 20ms) for 200+ simultaneous users/sessions with zero mutation capability  
**Status**: Design complete; ready for implementation kickoff

---

## 1. Executive Summary

This document outlines the design and phased implementation plan for a high-performance, production-grade **Model Context Protocol (MCP)** server for RisingWave written in Go.

The Python reference implementation provides an excellent functional foundation with **100+ tools** for natural-language interaction with RisingWave (queries, DDL/DML, streaming job monitoring, storage analysis, schema management, etc.). However, it has architectural limitations for scale: single global DB connection (no pooling), Python GIL + sync tool execution constraints, high per-process memory in stdio mode, and missing production features (auth, observability, rate limiting, robust error handling).

**The Go implementation** addresses these by:
- Using the **official `modelcontextprotocol/go-sdk`** for typed, schema-driven tool registration and dual transports (stdio + Streamable HTTP).
- Replacing the single-connection model with **`pgx` + `pgxpool`** for efficient, concurrent, pooled access to RisingWave (Postgres wire protocol).
- Embracing Go’s lightweight goroutines for true high-concurrency handling of 200+ simultaneous MCP sessions.
- Enforcing a **strictly read-only** posture: **no DDL, DML, or any mutating operations are possible** through the MCP server or agents. All tools are read-only by design.
- Targeting aggressive performance: **p99 tool latency < 20ms** under realistic load.
- Providing **very detailed observability** (per-tool metrics, query tracing, pool stats, validation overhead, etc.).
- Adding comprehensive production features from day one: structured observability, security model, resilience patterns, and multi-layer safeguards around the powerful generic **read-only** query executor.
- Maintaining (and improving upon) the modular, category-based tool organization of the Python version while removing all mutating capabilities.

**Expected outcomes**:
- Extremely low and predictable latency suitable for interactive AI agents.
- Native support for high concurrency (200+ sessions) with minimal resource usage.
- Single static binary, tiny Docker images, Kubernetes-native deployment.
- Maximum safety: agents cannot accidentally or maliciously modify schema or data.
- Rich observability for debugging agent behavior and performance tuning.
- Easier long-term maintenance and extension as RisingWave evolves.

This design incorporates all prior recommendations plus the new requirements: proper connection pooling, typed MCP tools, production hardening, realistic effort estimates, p99 < 20ms target, powerful generic read-only executor with safeguards, very detailed observability, and a clear path to supporting 200 concurrent users comfortably with zero mutation capability.

---

## 2. Goals & Non-Goals

### Primary Goals
- Deliver a **strictly read-only** MCP server: **zero possibility of destruction or mutation** via MCP tools or AI agents (no DDL, DML, or any write operations).
- Provide a **powerful generic read-only query executor** with multiple layers of safeguards (query type validation, resource limits, AST-level analysis where feasible, logging, rate limiting).
- Achieve **aggressive performance**: p99 end-to-end tool latency **< 20ms** under realistic concurrent load for 200 sessions.
- Deliver **very detailed observability** across every layer (MCP request, validation, DB execution, result serialization, pool behavior).
- Achieve **production readiness** for environments with up to 200 simultaneous MCP clients/sessions.
- Provide **excellent developer and agent experience** (rich tool descriptions, strong typing, clear errors, safe powerful query capabilities).
- Enable **horizontal scalability** and operational excellence (observability, deployment, resilience).
- Maintain close behavioral parity with the read-only portions of the Python reference while adding stronger safety and performance guarantees.

### Non-Goals (for v1)
- Full re-implementation of every edge-case Python helper (some internal utilities can be simplified or improved in Go).
- Support for every experimental RisingWave feature on day one (focus on core + high-value tools first).
- Building a new MCP client — only server.
- Multi-database / multi-tenant isolation beyond connection-level configuration (future phase).

### Success Metrics (v1)
- **p99 end-to-end tool latency < 20ms** for common read operations under load from 200 concurrent sessions.
- Handles 200 concurrent long-lived MCP sessions with stable performance and no mutation capability.
- Memory footprint < 80MB idle + connection pool.
- All read-oriented tool categories ported + one powerful generic safe read-query tool with multi-layer safeguards.
- Very detailed observability: per-tool + per-query metrics, tracing spans, validation overhead, DB pool detailed stats, and query sampling.
- Strictly enforced read-only posture with multiple independent enforcement layers (config, middleware, query validator, DB user privileges if desired).
- Comprehensive metrics, structured logs, and health checks.

---

## 3. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Clients (Claude, Copilot, etc.)       │
│   (stdio per session  OR  Streamable HTTP / SSE to shared server)│
└───────────────────────────────┬─────────────────────────────────┘
                                │
                ┌───────────────▼───────────────┐
                │     Go MCP Server (single binary) │
                │  • modelcontextprotocol/go-sdk   │
                │  • Dual transport support        │
                │  • Typed tool handlers           │
                │  • Middleware stack (auth, rate, │
                │    logging, metrics, timeout)    │
                └───────────────┬───────────────┘
                                │
                ┌───────────────▼───────────────┐
                │   Database Access Layer        │
                │   • pgx + pgxpool (pooled)     │
                │   • Connection validation      │
                │   • Query timeouts & retries   │
                │   • Context propagation        │
                └───────────────┬───────────────┘
                                │
                ┌───────────────▼───────────────┐
                │      RisingWave Cluster        │
                │   (distributed streaming DB)   │
                └────────────────────────────────┘
```

**Key improvements over Python**:
- **Shared pooled connections** instead of single global instance.
- **Goroutine-per-request** model (cheap and scalable).
- **Typed tool definitions** with automatic JSON schema generation.
- **Middleware pipeline** for cross-cutting concerns.
- **Stateless core** (easy horizontal scaling behind load balancer).

---

## 4. Technology Stack

| Layer                  | Choice                                      | Rationale |
|------------------------|---------------------------------------------|---------|
| Language               | Go 1.23+                                    | Performance, concurrency, static binary, excellent Postgres support |
| MCP SDK                | `github.com/modelcontextprotocol/go-sdk` (official) | Typed tools, schema generation, stdio + Streamable HTTP, actively maintained with Google involvement |
| Database Driver        | `github.com/jackc/pgx/v5` + `pgxpool`       | Best-in-class Postgres driver; native connection pooling, context support, high performance |
| Configuration          | `github.com/spf13/viper` + env + struct tags | Flexible, validation-friendly |
| Logging                | `log/slog` (structured)                     | Standard library, high performance |
| Metrics / Observability| `prometheus/client_golang` + OpenTelemetry (optional) | Industry standard |
| HTTP Server (for Streamable HTTP) | Standard `net/http` or lightweight router if needed | SDK provides handlers |
| Validation             | `github.com/go-playground/validator/v10`    | Struct tag validation for tool inputs |
| Testing                | `testing` + `testcontainers-go` + `pgx` mocks | Realistic integration tests |
| Containerization       | Multi-stage Docker (distroless or alpine)   | Tiny images (~30-50MB) |
| Deployment             | Kubernetes + HPA, docker-compose for local  | Production & dev parity |

**Why not other options?**
- Community MCP libs (mark3labs/mcp-go, etc.) are good but official SDK is preferred for long-term alignment.
- `database/sql` is viable but `pgx` offers superior performance and features for this use case.

---

## 5. Recommended Project Structure

```
risingwave-mcp-go/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point, config loading, server bootstrap
├── internal/
│   ├── config/                     # Viper + env config + validation
│   ├── db/                         # pgxpool wrapper, query helpers, health checks
│   │   ├── pool.go
│   │   ├── queries.go              # Common/reusable queries
│   │   └── health.go
│   ├── mcp/                        # MCP server setup, middleware, transports
│   │   ├── server.go
│   │   ├── middleware/
│   │   └── transports/
│   ├── tools/                      # All tool categories (modular like Python)
│   │   ├── query/
│   │   ├── ddl/
│   │   ├── dml/
│   │   ├── schema/
│   │   ├── streaming/
│   │   ├── storage/
│   │   ├── source/
│   │   ├── sink/
│   │   ├── cluster/
│   │   ├── management/
│   │   ├── user/
│   │   ├── iceberg/
│   │   └── catalog/
│       └── register.go         # Central registration function
│   ├── safety/                     # Identifier validation, SQL escaping, read-only mode
│   ├── observability/              # Metrics, logging helpers, tracing
│   └── server/                     # HTTP server setup if needed beyond SDK
├── pkg/                            # Public packages (if any reusable components)
├── tests/
│   ├── integration/                # testcontainers + RisingWave
│   └── e2e/
├── Dockerfile
├── docker-compose.yml              # Local RisingWave + MCP server
├── Makefile
├── go.mod
├── README.md
├── DESIGN.md                       # This document
└── .github/workflows/              # CI (build, test, lint, Docker)
```

This structure mirrors the Python modularity while leveraging Go packages and strong typing.

---

## 6. Detailed Component Designs

### 6.1 Database Access Layer (`internal/db`)

**Core principles**:
- One shared `*pgxpool.Pool` per server instance.
- All queries go through context-aware methods with timeouts.
- Automatic connection health checks and background maintenance (pgxpool handles most of this).
- Support for both simple queries and prepared statements.
- Read-only mode flag that can reject mutating statements at this layer or higher.

**Key types**:
```go
type DB struct {
    pool   *pgxpool.Pool
    config Config
    // metrics, logger, etc.
}

func (db *DB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
func (db *DB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
// Higher-level helpers: FetchRowsAsJSON, ExecuteDDL, etc.
```

**Improvements over Python**:
- True connection pooling with tunable `MinConns`, `MaxConns`, `MaxConnLifetime`, etc.
- Per-query timeouts via context.
- Better concurrency (many goroutines can use the pool simultaneously).

### 6.2 MCP Server & Tool Framework

Use official SDK patterns:

```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "risingwave-mcp",
    Version: "0.1.0",
}, nil)

// Example typed tool registration
mcp.AddTool(server, &mcp.Tool{
    Name:        "run_select_query",
    Description: "Execute a read-only SELECT or WITH query...",
    Annotations: &mcp.ToolAnnotations{ReadOnlyHint: boolPtr(true)},
}, queryHandler)
```

**Tool handler signature** (recommended):
```go
func queryHandler(ctx context.Context, req *mcp.CallToolRequest, input QueryInput) (*mcp.CallToolResult, QueryOutput, error)
```

**Benefits of typed approach**:
- Automatic JSON schema generation from Go structs.
- Compile-time safety.
- Easy input validation with `validator` tags.
- Clear `Output` struct for structured results (SDK can handle it).

**Tool Registration Strategy**:
- One `register_xxx_tools(server *mcp.Server, db *db.DB, safety *safety.Service)` per category.
- Central `RegisterAllTools(...)` in `tools/register.go`.
- Tools return either structured output or `mcp.TextContent` / `mcp.JSONContent`.

### 6.3 Safety, Read-Only Enforcement & Validation Layer (Core to Design)

Because **no destruction or mutation is allowed**, safety is elevated to a first-class, multi-layer architectural concern:

**Enforcement Layers (defense in depth)**:
1. **Configuration / Startup**: Hard `READ_ONLY=true` flag. Server refuses to start or register any mutating tools if not set.
2. **Middleware Layer**: Request-level interceptor that rejects any tool call not explicitly marked as read-only.
3. **Tool-level**: All registered tools declare `ReadOnlyHint: true`. No DDL/DML tools are registered at all.
4. **Generic Query Executor Safeguards** (the powerful generic tool):
   - Query type allow-list: Only `SELECT`, `WITH`, `EXPLAIN`, `SHOW`, `DESCRIBE`, and safe catalog queries.
   - Basic query analysis / keyword scanning to reject `INSERT`, `UPDATE`, `DELETE`, `CREATE`, `ALTER`, `DROP`, `TRUNCATE`, `GRANT`, etc.
   - (Future) Lightweight SQL parser or RisingWave-specific validation for deeper safety.
   - Hard resource limits: `max_rows_returned`, `query_timeout_ms` (e.g., 5s), `max_query_length`.
   - Every executed query is logged (sanitized) with full context for audit.
5. **DB Connection Layer**: Optional connection with a read-only database user/role on the RisingWave side.
6. **Result Layer**: Size limits and streaming for large results to prevent memory exhaustion.

**Identifier & Input Validation** (retained & strengthened):
- Strict `validate_identifier` for all object names.
- Parameterized queries via `pgx` everywhere (injection prevention by design).
- Struct validation with `validator` tags on all tool inputs.

This design makes it **impossible** for an agent to perform destructive actions through this MCP server.

### 6.4 Tool Categories (Strictly Read-Only Focus)

All mutating categories (DDL, DML) are **explicitly excluded** from this implementation. The server is read-only by architecture.

**Supported Categories** (focused on introspection, monitoring, and safe querying):

| Category                  | Approx. Tools | Key Characteristics in Go |
|---------------------------|---------------|---------------------------|
| Query & Generic Safe Executor | 4+         | Parameterized SELECT/WITH + one **powerful generic `execute_safe_read_query`** with multi-layer validation, resource limits, and full audit logging |
| Explain                   | 3             | EXPLAIN, EXPLAIN ANALYZE, distributed plans (read-only) |
| Schema Inspection         | 15+           | describe_*, show_*, list_* — heavy use of information_schema + RisingWave catalog |
| Streaming Monitoring      | 8+            | Fragments, actors, backfill progress, job status |
| Storage / Hummock         | 5+            | Compaction stats, SST analysis |
| Cluster & Management      | 10+           | Version, session vars, cluster health (read-only views) |
| Sources / Sinks (describe)| 8+            | Inspection and metadata only (no create/alter) |
| Catalog & Iceberg         | 8+            | System catalog queries, Iceberg maintenance metadata (read-only) |
| User & Privileges (view)  | 5+            | Read-only user and privilege inspection |

**The Powerful Generic Read-Only Executor** (`execute_safe_read_query`):
- Accepts a query string.
- Runs through the full Safety & Validation Layer (keyword blocklist + type check + limits).
- Executes with strict timeout and row limit.
- Returns results (JSON or structured) + metadata (execution time, rows returned, cache hit?).
- Every invocation is fully logged for audit and debugging.
- This tool gives agents significant power for ad-hoc analysis while remaining safe.

**Example Port Pattern**:
- Input struct with `Query string` + optional `max_rows`, `timeout_ms`.
- Handler runs validation pipeline → executes via pooled `pgx` with context deadline → returns structured result or clear error.
- Reusable helpers: `validateReadOnlyQuery`, `executeQueryToStructured`, `enforceResourceLimits`.

Many tools are thin wrappers around catalog queries; create reusable helpers for common patterns.

### 6.5 Latency Optimization Strategy (Targeting p99 < 20ms)

Achieving **p99 tool latency under 20ms** is an explicit non-negotiable target. The following techniques are designed in from the start:

- **Connection Pool Tuning**: Pre-warm connections, aggressive `MaxConnLifetime` and health checks, low `AcquireTimeout`. Size pool appropriately for expected concurrency.
- **Prepared Statements & Query Caching**: Cache prepared statements for frequent catalog/metadata queries. Consider a small in-memory cache (with short TTL or invalidation on schema changes) for expensive but stable metadata (table schemas, stats).
- **Minimal Allocations & Fast Paths**: Use efficient result handling (`pgx` rows → struct or fast JSON). Avoid heavy reflection in hot paths.
- **Context & Timeout Discipline**: Strict per-query deadlines (e.g., 10–15ms target for DB portion). Early rejection on validation.
- **Result Serialization**: Prefer compact/structured output over pretty-printed JSON when possible. Stream large results if needed.
- **Hot Path Optimization**: Profile early with `pprof`. Focus on validation speed, pool acquire time, and DB round-trip.
- **Deployment Assumptions for Target**: Low-latency network between MCP server and RisingWave (same availability zone or colocated), RisingWave tuned for low query latency on metadata/catalog queries, sufficient CPU cores. Document these requirements clearly.
- **Measurement**: Built-in latency breakdown metrics (as part of detailed observability) so teams can see exactly where time is spent and iterate.

This combination of Go’s efficiency, pooled connections, caching, and observability-driven tuning makes sub-20ms p99 realistic for the majority of read/metadata/monitoring tools.

---

## 7. Production Readiness Features

### 7.1 Very Detailed Observability (First-Class Requirement)

Observability is designed to be extremely rich to support debugging agent behavior, performance tuning for the <20ms p99 target, and operational insight.

**Metrics (Prometheus + custom)**:
- Per-tool: call count, latency (p50/p95/p99), error rate, success rate.
- Per generic query executor: validation time, query parse/scan time, DB execution time, result serialization time, rows returned distribution.
- DB Connection Pool: active/idle conns, wait time, acquire time, total queries, errors (detailed `pgxpool` stats + custom).
- Validation & Safety layer: identifier validation failures, query rejection reasons (by category), resource limit hits.
- MCP layer: transport (stdio vs HTTP), request size, response size, concurrent sessions gauge.
- System: Go runtime (goroutines, GC pauses, memory), CPU.

**Structured Logging (slog)**:
- Every tool invocation: `tool_name`, `session_id` (if available), `duration_ms`, `db_duration_ms`, `validation_duration_ms`, `rows_returned`, `error_code`, `query_hash` (for generic executor, to avoid logging full sensitive queries).
- Sampled full query logging for the generic executor (configurable sample rate).
- Structured error classification (validation_error, db_error, timeout, etc.).
- Startup config dump (redacted) and pool configuration.

**Distributed Tracing (OpenTelemetry recommended)**:
- Spans: `mcp.tool_call` → `safety.validate` → `db.acquire_conn` → `db.execute` → `serialize_result`.
- Attributes on every span: tool name, query type, latency breakdowns.
- Correlation IDs propagated from MCP request to DB query.

**Health & Readiness**:
- `/healthz` — basic liveness.
- `/readyz` — checks DB pool health + recent successful query.
- `/metrics` — full Prometheus exposition.
- Optional internal `/debug` endpoints (pprof, pool stats) protected by auth.

**Dashboards & Alerting (recommended)**:
- Grafana dashboard with panels for: p99 latency by tool, pool saturation, validation rejection rate, generic query patterns, error breakdown.
- Alerts on: p99 > 15ms sustained, pool exhaustion, high validation failure rate, rising error rate.

This level of detail enables precise identification of whether latency comes from validation, connection acquisition, RisingWave execution, or serialization — critical for hitting <20ms p99.

---

## 8. Scalability Considerations for 200+ Simultaneous Users

- **Connection Pool Sizing**: Tune `MaxConns` to 50–200 depending on RisingWave cluster size and workload (pgxpool is very efficient).
- **Goroutine Model**: Each MCP request/session is a goroutine. Thousands are trivial.
- **HTTP Mode Preferred for Scale**: One (or few) server instances behind a load balancer or Kubernetes Service. Session affinity usually not required.
- **Stdio Mode**: Still useful per-developer; not the path for 200 shared users.
- **Horizontal Scaling**: Stateless design allows easy replica scaling. RisingWave handles the heavy lifting.
- **Resource Estimate (rough)**: On modest hardware (4 vCPU, 4–8 GB), one instance should comfortably handle 200+ bursty AI agent sessions.

**Load Testing Plan**: Use custom load generator or `hey`/`k6` or custom load gen simulating AI tool calls.

---

## 9. Implementation Roadmap (Phased)

### Phase 0: Foundations (Week 1)
- Project scaffolding, go.mod, basic config + DB pool.
- Hello-world MCP server with one typed tool.
- Dockerfile + basic CI.
- Local docker-compose with RisingWave.

### Phase 1: Core Infrastructure + Query Tools (Weeks 2–3)
- Full DB layer with pooling, helpers, health.
- Safety/validation package.
- Query, Explain, and basic Schema tools (highest usage).
- Middleware skeleton (logging, metrics, timeout).
- Integration tests with testcontainers.

### Phase 2: DDL/DML + Management Tools (Weeks 4–5)
- All DDL, DML, management, session, cluster tools.
- Read-only mode enforcement.
- Tool annotations for destructive operations.
- Audit logging for mutations.

### Phase 3: Streaming, Storage, Source/Sink, Advanced (Weeks 6–7)
- Streaming monitoring, storage/Hummock, sources, sinks, Iceberg, users, catalog.
- Complete tool registration.
- Rich error handling and result formatting.

### Phase 4: Production Hardening (Weeks 8–9)
- Full observability (Prometheus + structured logs).
- Auth middleware + read-only mode.
- Rate limiting, resilience patterns.
- Comprehensive documentation and examples.
- Performance tuning and load testing.

### Phase 5: Polish, Release & Handover (Week 10)
- End-to-end testing against real RisingWave features.
- README, usage examples for Claude/VS Code, deployment guides.
- Release v0.1.0 (or v1.0.0) Docker image + GitHub release.
- Migration guide from Python version.
- Open issues for future enhancements.

**Total**: ~8–10 weeks for a high-quality v1 by one senior Go engineer (or faster with a small team). Many repetitive tools accelerate after the first 20–30 are done.

---

## 10. Testing Strategy

- **Unit**: Every handler + helper (table-driven tests).
- **Integration**: `testcontainers-go` spinning up RisingWave; run real tool calls.
- **MCP Protocol**: Use MCP inspector or custom client to verify tool listing, calling, errors.
- **Load / Concurrency**: Custom Go test or external tool simulating 200 concurrent sessions.
- **Security**: Fuzzing on inputs, SQL injection attempts (should be blocked by parameterization).
- **Chaos**: Simulate DB connection loss, slow queries.

---

## 11. Risks, Assumptions & Mitigations

| Risk                              | Likelihood | Impact | Mitigation |
|-----------------------------------|------------|--------|----------|
| Behavioral differences in tool output | Medium    | High (agent UX) | Thorough side-by-side testing with Python version; rich error messages |
| RisingWave catalog query changes  | Low–Medium | Medium | Abstract queries; good test coverage; monitor RisingWave releases |
| Overly permissive DDL tools       | High      | High (safety) | Strong validation + read-only mode + annotations + audit logs from day one |
| Performance regression in some tools | Low      | Medium | Benchmark early; use prepared statements and indexes where possible |
| MCP SDK evolution                 | Low       | Low    | Pin version; follow official updates; contribute if needed |

**Key Assumptions**:
- RisingWave remains Postgres-wire compatible for the queries we use.
- Most tools are read-heavy or monitoring; mutating tools are used cautiously.
- Teams will adopt HTTP mode for shared production use.

---

## 12. Future Roadmap (Post v1)

- Structured output schemas for better agent consumption.
- Progress notifications for long-running operations (backfills, etc.).
- OAuth / fine-grained RBAC integration.
- Multi-cluster / RisingWave Cloud support.
- Web UI or admin dashboard for the MCP server itself.
- Agent Skills integration (complementary to MCP).
- Automatic tool description improvement via LLM or static analysis.

---

## 13. Open Questions for the Team (Socratic Prompts)

1. **Strict No-Destruction Policy**: With mutations completely removed, how do we communicate the capabilities and limitations clearly to agent developers and users so they understand what is (and is not) possible through this MCP server?

2. **Powerful Generic Read-Only Executor Safeguards**: What is the right balance between power (allowing complex ad-hoc analytical queries) and safety in the generic query tool? Should we add lightweight SQL parsing/AST analysis in v1, or start with keyword + type validation + strict limits and iterate based on real usage?

3. **p99 < 20ms Feasibility**: Given the target of p99 latency under 20ms, what deployment constraints are acceptable (e.g., network latency to RisingWave, hardware sizing, caching strategy)? How will we validate this target in load testing?

4. **Observability Value**: With very detailed tracing and per-layer timing, which specific breakdowns or custom metrics would be most useful for your team when tuning agent performance or debugging slow tool calls?

5. **Future Mutation Path**: If there is ever a need for controlled mutation in the future, how should we design the architecture now (separate server? capability-based? human-in-the-loop approval) so it doesn’t compromise the current strict safety model?

---

## Appendix: Tool Inventory Summary (from Python reference)

The Go implementation will aim for full parity across:
- Query & Explain tools
- Schema inspection & listing
- Full DDL suite (tables, MVs, columns, parallelism, rate limits, etc.)
- DML (insert/update/delete with batch support)
- Source & Sink management
- Streaming job monitoring (fragments, actors, backfill)
- Storage / Hummock analysis
- Cluster, session, management, user, secret, function, index, connection, catalog, and Iceberg tools

Exact count and naming will be maintained in code and README.

---

**This design is ready for review and implementation.**  
It balances fidelity to the excellent Python reference with the performance, safety, and operational characteristics required for serious production use with AI agents at scale.

**Next step recommendation**: Approve this design, then begin Phase 0 scaffolding in a new repository (or fork/contribution to risingwavelabs if desired).

---

*Document generated with care for clarity, completeness, and actionability. Feedback welcome.*