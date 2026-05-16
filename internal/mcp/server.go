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
			return nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}
		rowMap := make(map[string]any)
		for i, col := range columns {
			rowMap[col] = values[i]
		}
		resultRows = append(resultRows, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("row iteration error: %w", err)
	}

	result, _ := json.Marshal(map[string]any{
		"columns": columns,
		"rows":    resultRows,
		"count":   len(resultRows),
	})

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(result)},
		},
	}, nil, nil
}

func (s *Server) Start() error {
	fmt.Println("Starting MCP server with stdio transport")
	return s.mcpServer.Run(context.Background(), &mcp.StdioTransport{})
}