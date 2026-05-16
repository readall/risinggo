package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/readall/risinggo/internal/config"
)

func main() {
	// Print all environment variables that start with DATABASE, READ, DB_, etc.
	fmt.Println("=== Environment Variables ===")
	for _, e := range os.Environ() {
		if containsAny(e, []string{"DATABASE", "READ", "DB_", "MAX_CONNS", "MIN_CONNS", "CONN_", "QUERY_", "MAX_ROWS", "TRANSPORT", "HTTP_", "LOG_", "METRICS_", "ENABLE_"}) {
			fmt.Println(e)
		}
	}
	fmt.Println("============================")
	
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Config loaded: %+v\n", cfg)
}

func containsAny(s string, list []string) bool {
	for _, v := range list {
		if strings.HasPrefix(s, v) {
			return true
		}
	}
	return false
}