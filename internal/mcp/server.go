package mcp

import (
	"context"
	"encoding/json"
	"fmt"

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
	config    *config.Config
	pool      *db.Pool
	mcpServer *mcp.Server
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

func (s *Server) Start() error {
	fmt.Println("Starting MCP server with stdio transport")
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}