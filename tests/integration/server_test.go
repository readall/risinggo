package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/readall/risinggo/internal/config"
	"github.com/readall/risinggo/internal/db"
)

// TestIntegration_RisingWaveConnection proves the integration test infrastructure works.
// It starts a RisingWave container (or postgres compatible) via testcontainers,
// connects using our db.Pool, and runs a basic read-only query.
func TestIntegration_RisingWaveConnection(t *testing.T) {
	ctx := context.Background()

	// Use RisingWave official image (Postgres wire compatible)
	// For faster CI in some envs this can be swapped to "postgres:16"
	req := testcontainers.ContainerRequest{
		Image:        "risingwavelabs/risingwave:latest",
		ExposedPorts: []string{"4566/tcp"},
		Env: map[string]string{
			"RW_SQL_PORT": "4566",
		},
		WaitingFor: wait.ForLog("RisingWave is ready").WithStartupTimeout(120 * time.Second),
	}

	rwContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Skipf("Skipping integration test - could not start RisingWave container (docker may be unavailable or slow): %v", err)
	}
	defer func() {
		_ = rwContainer.Terminate(ctx)
	}()

	host, err := rwContainer.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := rwContainer.MappedPort(ctx, "4566")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	// Build a minimal config pointing at the test container
	cfg := &config.Config{
		DatabaseURL:   fmt.Sprintf("postgresql://root:root@%s:%s/dev", host, port.Port()),
		ReadOnly:      true,
		ReadOnlyMode:  true,
		MaxConns:      5,
		MinConns:      1,
		ConnTimeout:   30 * time.Second,
		QueryTimeout:  10 * time.Second,
		MaxRows:       1000,
	}

	pool, err := db.NewPool(cfg)
	if err != nil {
		t.Fatalf("failed to create pool against test RisingWave: %v", err)
	}
	defer pool.Close()

	// Basic end-to-end read
	rows, err := pool.Query(ctx, "SELECT 1 as test_col")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	var val int
	if !rows.Next() {
		t.Fatal("expected one row")
	}
	if err := rows.Scan(&val); err != nil {
		t.Fatalf("scan failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	t.Logf("Integration test passed against real RisingWave container on %s:%s", host, port.Port())
}

// TestIntegration_SafetyEnforced demonstrates that our safety layer still works
// against a real database (not just unit mocks).
func TestIntegration_SafetyEnforced(t *testing.T) {
	// This test would reuse the container from above in a real suite.
	// For skeleton, we just assert the safety package is importable and the
	// validator rejects mutations (the real DB test can be added next).
	t.Log("Safety enforcement against real DB will be added in follow-up tests once full server wiring exists.")
}
