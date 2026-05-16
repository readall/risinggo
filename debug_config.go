package main

import (
	"fmt"
	"log"
	"os"

	"github.com/readall/risinggo/internal/config"
)

func main() {
	fmt.Println("DATABASE_URL from os.Getenv:", os.Getenv("DATABASE_URL"))
	fmt.Println("READ_ONLY from os.Getenv:", os.Getenv("READ_ONLY"))
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Config loaded: %+v\n", cfg)
}