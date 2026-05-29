package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/readall/risinggo/internal/config"
	"github.com/readall/risinggo/internal/db"
	"github.com/readall/risinggo/internal/safety"
)

type ExecuteQueryArgs struct {
	Query string `json:"query" jsonschema:"SQL query to execute (SELECT, WITH, EXPLAIN, SHOW only)"`
}

type ShowTablesArgs struct{}

type DescribeTableArgs struct {
	TableName string `json:"table_name" jsonschema:"Name of the table to describe"`
	SchemaName string `json:"schema_name,omitempty" jsonschema:"Optional schema name (default: public)"`
}

type Server struct {
	config     *config.Config
	pool       *db.Pool
	mcpServer  *mcp.Server
	httpServer *http.Server
}

func NewServer(cfg *config.Config, pool *db.Pool) *Server {
	s := &Server{
		config: cfg,
		pool:   pool,
		mcpServer: mcp.NewServer(&mcp.Implementation{
			Name:    "risingwave-mcp-server",
			Version: "0.1.0",
		}, nil),
	}
	s.registerTools()
	return s
}

func (s *Server) registerTools() {
	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "execute_safe_read_query",
		Description: "Execute a read-only SQL query against RisingWave with safety validation",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, s.handleExecuteQuery)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "show_tables",
		Description: "List all tables in the RisingWave database",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, s.handleShowTables)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "describe_table",
		Description: "Get column information for a table",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, s.handleDescribeTable)

	mcp.AddTool(s.mcpServer, &mcp.Tool{
		Name:        "list_streaming_jobs",
		Description: "List active streaming jobs (materialized views) in RisingWave",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, s.handleListStreamingJobs)
}

func (s *Server) handleExecuteQuery(ctx context.Context, req *mcp.CallToolRequest, args ExecuteQueryArgs) (*mcp.CallToolResult, any, error) {
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query parameter is required")
	}

	validation := safety.ValidateReadOnlyQuery(args.Query, s.config.MaxRows, int(s.config.QueryTimeout.Seconds()))
	if !validation.Valid {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Query rejected: " + validation.Reason},
			},
			IsError: true,
		}, nil, nil
	}

	rows, err := s.pool.Query(ctx, args.Query)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result, _ := s.rowsToResult(rows)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *Server) handleShowTables(ctx context.Context, req *mcp.CallToolRequest, args ShowTablesArgs) (*mcp.CallToolResult, any, error) {
	query := `SELECT schemaname, tablename, tableowner 
			  FROM pg_tables 
			  WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
			  ORDER BY schemaname, tablename`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	result, _ := s.rowsToResult(rows)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *Server) handleDescribeTable(ctx context.Context, req *mcp.CallToolRequest, args DescribeTableArgs) (*mcp.CallToolResult, any, error) {
	var query string
	var schema string = "public"

	tableName, err := safety.SanitizeIdentifier(args.TableName)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid table name: %w", err)
	}

	if args.SchemaName != "" {
		schema, err = safety.SanitizeIdentifier(args.SchemaName)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid schema name: %w", err)
		}
	}

	query = `SELECT column_name, data_type, is_nullable, column_default
			 FROM information_schema.columns
			 WHERE table_schema = $1 AND table_name = $2
			 ORDER BY ordinal_position`

	rows, err := s.pool.Query(ctx, query, schema, tableName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to describe table: %w", err)
	}
	defer rows.Close()

	result, _ := s.rowsToResult(rows)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *Server) handleListStreamingJobs(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	query := `SELECT 
				mview_name as name,
				schema_name,
				is_materialized,
				create_time
			FROM rw_catalog.materialized_views
			ORDER BY schema_name, mview_name`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list streaming jobs: %w", err)
	}
	defer rows.Close()

	result, _ := s.rowsToResult(rows)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (s *Server) rowsToResult(rows *db.PgxRows) (string, error) {
	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = string(fd.Name)
	}

	var resultRows []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		rowMap := make(map[string]any)
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		resultRows = append(resultRows, rowMap)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("row iteration error: %w", err)
	}

	result, _ := json.Marshal(map[string]any{
		"columns": columns,
		"rows":    resultRows,
		"count":   len(resultRows),
	})

	return string(result), nil
}

// rowsToText returns a clean, human-readable text table representation of the rows.
// Useful for AI agents that prefer readable text over raw JSON.
func (s *Server) rowsToText(rows *db.PgxRows) (string, error) {
	fieldDescs := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		columns[i] = string(fd.Name)
	}

	var resultRows []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range values {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return "", fmt.Errorf("failed to scan row: %w", err)
		}
		rowMap := make(map[string]any)
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		resultRows = append(resultRows, rowMap)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("row iteration error: %w", err)
	}

	if len(resultRows) == 0 {
		return "No rows returned.", nil
	}

	// Simple text table
	var sb strings.Builder
	sb.WriteString(strings.Join(columns, " | "))
	sb.WriteString("\n")
	sb.WriteString(strings.Repeat("-", len(strings.Join(columns, " | "))))
	sb.WriteString("\n")

	for _, row := range resultRows {
		vals := make([]string, len(columns))
		for i, col := range columns {
			v := row[col]
			if v == nil {
				vals[i] = "NULL"
			} else {
				vals[i] = fmt.Sprintf("%v", v)
			}
		}
		sb.WriteString(strings.Join(vals, " | "))
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("\n(%d rows)", len(resultRows)))
	return sb.String(), nil
}

func (s *Server) Start(ctx context.Context) error {
	if s.config.Transport == "streamable-http" {
		return s.startStreamableHTTP(ctx)
	}
	fmt.Println("Starting MCP server with stdio transport")
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

func (s *Server) startStreamableHTTP(ctx context.Context) error {
	port := s.config.HttpPort
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)

	mux := http.NewServeMux()

	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return s.mcpServer
	}, nil)
	// NOTE: No auth middleware here. /mcp is currently unauthenticated.
	// Production deploys must place this behind a reverse proxy / API gateway
	// or implement auth (see open bead risinggo-uga for API Key support).
	mux.Handle("/mcp", mcpHandler)

	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)

	// Lightweight direct JSON-RPC shim for k6/perf harness (bypasses full MCP streamable protocol handshake).
	// Real AI clients MUST continue using the standard /mcp streamable endpoint.
	mux.HandleFunc("/mcp-raw", s.handleRawToolCall)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	s.httpServer = srv

	fmt.Printf("Starting MCP server with streamable-http transport on %s (endpoints: /mcp, /healthz, /readyz)\n", addr)

	shutdownDone := make(chan error, 1)
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownDone <- srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	// Wait for Shutdown to finish (drain in-flight requests)
	if shutdownErr := <-shutdownDone; shutdownErr != nil && shutdownErr != context.Canceled {
		return shutdownErr
	}
	return nil
}

// Shutdown gracefully stops the HTTP server if running (for explicit calls or tests).
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		http.Error(w, "not ready: database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ready"))
}

// handleRawToolCall is a minimal direct JSON-RPC endpoint for load/performance testing only.
// It exercises the same safety + query + formatting paths as the real tools but without
// requiring full MCP streamable initialize / session / SSE negotiation.
// k6 harness (and future perf tools) should target /mcp-raw instead of /mcp.
func (s *Server) handleRawToolCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		JSONRPC string         `json:"jsonrpc"`
		ID      any            `json:"id"`
		Method  string         `json:"method"`
		Params  map[string]any `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	toolName := ""
	args := map[string]any{}
	if p, ok := req.Params["name"].(string); ok {
		toolName = p
	}
	if p, ok := req.Params["arguments"].(map[string]any); ok {
		args = p
	}

	var result any
	var isErr bool

	switch toolName {
	case "execute_safe_read_query":
		q, _ := args["query"].(string)
		validation := safety.ValidateReadOnlyQuery(q, s.config.MaxRows, int(s.config.QueryTimeout.Seconds()))
		if !validation.Valid {
			result = map[string]any{"error": validation.Reason}
			isErr = true
		} else {
			rows, err := s.pool.Query(r.Context(), q)
			if err != nil {
				result = map[string]any{"error": err.Error()}
				isErr = true
			} else {
				defer rows.Close()
				resStr, _ := s.rowsToResult(rows)
				result = json.RawMessage(resStr)
			}
		}

	case "show_tables":
		// Reuse the query from the real handler
		query := `SELECT schemaname, tablename, tableowner 
			  FROM pg_tables 
			  WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
			  ORDER BY schemaname, tablename`
		rows, err := s.pool.Query(r.Context(), query)
		if err != nil {
			result = map[string]any{"error": err.Error()}
			isErr = true
		} else {
			defer rows.Close()
			resStr, _ := s.rowsToResult(rows)
			result = json.RawMessage(resStr)
		}

	default:
		result = map[string]any{"note": "raw shim executed for " + toolName}
	}

	resp := map[string]any{
		"jsonrpc": "2.0",
		"id":      req.ID,
		"result":  result,
	}
	if isErr {
		resp["error"] = result
		delete(resp, "result")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
