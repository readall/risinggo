# RisingWave MCP Server (Go) — Implementation Constraints & Architectural Decision Records (ADRs)

**Version**: 1.0  
**Status**: Binding for v1 implementation

## 1. Technology Constraints (Must Use)

- **Language & Runtime**: Go 1.23 or newer.
- **MCP Framework**: Must use the **official** `github.com/modelcontextprotocol/go-sdk`. No community forks unless the official SDK is abandoned.
- **Database Driver**: Must use `github.com/jackc/pgx/v5` + `pgxpool`. No `database/sql`, no `risingwave-py` or other drivers.
- **Configuration**: Environment variables + Viper (or equivalent). No config files for v1.
- **Logging**: `log/slog` (structured). JSON or text handler as appropriate.
- **Metrics**: `prometheus/client_golang`.
- **Tracing**: OpenTelemetry (optional but recommended for v1; spans must be present in design).
- **Validation**: `github.com/go-playground/validator/v10` for struct inputs.
- **Testing**: `testing` + `github.com/testcontainers/testcontainers-go` for integration tests against real RisingWave.
- **Docker**: Multi-stage build, non-root user, minimal base image.

## 2. Explicit Non-Goals / Forbidden

- **No mutations ever**: Do not implement, register, or expose any DDL, DML, or mutating tools. This is a hard architectural constraint (see ADR-001).
- **No dynamic plugin system** in v1.
- **No persistent local state** (the server must remain stateless).
- **No automatic schema migration** or direct modification of RisingWave objects.
- **Do not bypass the safety layers** for the generic query executor even for "trusted" callers in v1.

## 3. Architectural Decision Records (ADRs)

### ADR-001: Strictly Read-Only by Architectural Design
**Decision**: The MCP server shall make it impossible to perform any mutation (DDL/DML) through any mechanism.  
**Rationale**: Maximizes safety for AI agents. Prevents accidental or malicious damage. Aligns with "powerful but safe" philosophy.  
**Consequences**: All DDL/DML tools from Python reference are removed. Generic executor has strong rejection logic. Read-only mode is not optional.  
**Status**: Accepted.

### ADR-002: Official Go MCP SDK
**Decision**: Use `github.com/modelcontextprotocol/go-sdk`.  
**Rationale**: Official support, typed tools, schema generation, future compatibility, maintained with Google involvement.  
**Alternatives Considered**: mark3labs/mcp-go and others — rejected for long-term alignment.  
**Status**: Accepted.

### ADR-003: pgx + pgxpool as Sole Database Layer
**Decision**: All database access goes through a shared `pgxpool.Pool`.  
**Rationale**: Best performance, native connection pooling, context support, prepared statements. Enables p99 target and high concurrency.  
**Consequences**: Single connection pattern from Python version is forbidden. Pool tuning becomes critical.  
**Status**: Accepted.

### ADR-004: Multi-Layer Defense-in-Depth for Generic Read-Only Executor
**Decision**: The powerful generic query tool must be protected by configuration, middleware, keyword/type validation, resource limits, and parameterized execution.  
**Rationale**: Provides power while maintaining safety. Allows defense even if one layer fails.  
**Status**: Accepted.

### ADR-005: p99 Latency Target < 20ms (Measurable with 10% Variance)
**Decision**: Design and optimize so that p99 end-to-end tool latency **targets under 20ms** for common operations under 200 concurrent sessions, with up to 10% variance considered acceptable. Treat as a measurable continuous improvement target tracked via observability rather than a strict binary gate.  
**Rationale**: Enables responsive AI agent interactions while remaining realistic for production measurement and continuous improvement. Requires strong observability-driven optimization from day one.  
**Consequences**: Latency breakdown instrumentation is mandatory. Performance tests should measure and report against the target + variance. Deployment assumptions must be documented. Caching and pool tuning remain first-class concerns.  
**Status**: Accepted.

### ADR-006: Very Detailed Observability by Default
**Decision**: Tracing spans, per-layer metrics, and structured logging with correlation IDs shall be present for every tool call.  
**Rationale**: Essential for achieving and maintaining the p99 target and for debugging agent-driven workloads.  
**Consequences**: Observability cannot be disabled in production builds. Adds some overhead that must be measured.  
**Status**: Accepted.

### ADR-007: Dual Transport Support (stdio + Streamable HTTP)
**Decision**: Support both stdio (for local AI tools) and Streamable HTTP (for scaled/shared deployments).  
**Rationale**: Matches Python reference usage patterns while enabling production multi-user scenarios.  
**Status**: Accepted.

### ADR-008: Stateless Core for Horizontal Scalability
**Decision**: The server core shall be stateless. All state lives in the RisingWave connection pool and external observability systems.  
**Rationale**: Enables easy horizontal scaling behind load balancers or Kubernetes. Simplifies recovery.  
**Status**: Accepted.

## 4. Additional Implementation Constraints

- Every tool handler must be a typed `mcp.AddTool` registration.
- All database queries must use context with deadlines.
- Every rejection in the safety pipeline must produce a distinct, testable error classification.
- Latency measurement points must be instrumented before any optimization work.
- The generic `execute_safe_read_query` tool must be the primary (and most powerful) ad-hoc query mechanism.
- Read-only enforcement must be verifiable via both code inspection and runtime behavior tests.
- Dependency versions should be pinned in `go.mod` for reproducibility.

## 5. Verification of Constraints

- Static analysis / linters can check for absence of mutation-related code.
- Integration tests must attempt mutation queries and assert rejection.
- Performance tests must measure p99 latency and report against the <20ms target with up to 10% variance tolerance.
- Observability tests must assert presence of required spans, metrics, and log fields.

---

These constraints and ADRs are binding. Any deviation requires a new ADR and update to this document and the Gherkin specifications.