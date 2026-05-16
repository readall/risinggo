package main

import (
	"fmt"
	"log"

	"github.com/readall/risinggo/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Config loaded: %+v\n", cfg)
}