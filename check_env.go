package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("DATABASE_URL:", os.Getenv("DATABASE_URL"))
}