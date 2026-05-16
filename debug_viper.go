package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func main() {
	// Set environment variable explicitly
	os.Setenv("DATABASE_URL", "postgresql://root:root@localhost:4566/dev")

	v := viper.New()
	v.SetConfigType("env")
	v.AutomaticEnv()

	var cfg struct {
		DatabaseURL string `mapstructure:"DATABASE_URL"`
	}

	if err := v.Unmarshal(&cfg); err != nil {
		fmt.Printf("Unmarshal error: %v\n", err)
		return
	}

	fmt.Printf("Config from viper: %+v\n", cfg)
	fmt.Printf("DATABASE_URL from os.Getenv: %s\n", os.Getenv("DATABASE_URL"))
}