package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("DATABASE_URL:", os.Getenv("DATABASE_URL"))
	fmt.Println("READ_ONLY:", os.Getenv("READ_ONLY"))
	fmt.Println("READ_ONLY_MODE:", os.Getenv("READ_ONLY_MODE"))
	fmt.Println("DB_MAX_CONNS:", os.Getenv("DB_MAX_CONNS"))
	fmt.Println("DB_MIN_CONNS:", os.Getenv("DB_MIN_CONNS"))
}