package main

import (
	"fmt"
	"log"
	"os"

	"github.com/readall/risinggo/internal/config"
)

func main() {
	// Print the environment variables we're setting
	fmt.Println("Environment variables:")
	fmt.Printf("  DATABASE_URL=%s\n", os.Getenv("DATABASE_URL"))
	fmt.Printf("  MCP_TRANSPORT=%s\n", os.Getenv("MCP_TRANSPORT"))
	fmt.Printf("  HTTP_PORT=%s\n", os.Getenv("HTTP_PORT"))
	fmt.Printf("  READ_ONLY=%s\n", os.Getenv("READ_ONLY"))
	fmt.Printf("  READ_ONLY_MODE=%s\n", os.Getenv("READ_ONLY_MODE"))
	fmt.Printf("  MAX_CONNS=%s\n", os.Getenv("MAX_CONNS"))
	fmt.Printf("  MIN_CONNS=%s\n", os.Getenv("MIN_CONNS"))
	fmt.Printf("  CONN_TIMEOUT=%s\n", os.Getenv("CONN_TIMEOUT"))
	fmt.Printf("  QUERY_TIMEOUT=%s\n", os.Getenv("QUERY_TIMEOUT"))
	fmt.Printf("  MAX_ROWS=%s\n", os.Getenv("MAX_ROWS"))
	fmt.Printf("  LOG_LEVEL=%s\n", os.Getenv("LOG_LEVEL"))
	fmt.Printf("  METRICS_PORT=%s\n", os.Getenv("METRICS_PORT"))
	fmt.Printf("  ENABLE_METRICS=%s\n", os.Getenv("ENABLE_METRICS"))
	
	// Now try to load the config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("\nLoaded config: %+v\n", cfg)
}