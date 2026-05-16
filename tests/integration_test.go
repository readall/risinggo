package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestMCPIntegrationWithRisingWave is a skeleton for full integration testing.
// It spins up a real RisingWave instance using testcontainers.
func TestMCPIntegrationWithRisingWave(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start RisingWave container
	rwReq := testcontainers.ContainerRequest{
		Image:        "risingwavelabs/risingwave:latest",
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForListeningPort("4566/tcp").WithStartupTimeout(60 * time.Second),
		Cmd:          []string{"playground"},
	}

	rwContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: rwReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer rwContainer.Terminate(ctx)

	// TODO: Start your MCP server container or connect to a running one
	// TODO: Perform actual MCP tool calls against the server connected to this RisingWave

	t.Log("RisingWave container started successfully. Add MCP server + tool call assertions here.")
}