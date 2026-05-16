package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	DatabaseURL   string        `mapstructure:"DATABASE_URL"`
	ReadOnly      bool          `mapstructure:"READ_ONLY"`
	ReadOnlyMode  bool          `mapstructure:"READ_ONLY_MODE"`
	MaxConns      int32         `mapstructure:"DB_MAX_CONNS"`
	MinConns      int32         `mapstructure:"DB_MIN_CONNS"`
	ConnTimeout   time.Duration `mapstructure:"DB_CONN_TIMEOUT"`
	QueryTimeout  time.Duration `mapstructure:"QUERY_TIMEOUT"`
	MaxRows       int           `mapstructure:"MAX_ROWS"`
	Transport     string        `mapstructure:"TRANSPORT"`
	HttpPort      int           `mapstructure:"HTTP_PORT"`
	LogLevel      string        `mapstructure:"LOG_LEVEL"`
	MetricsPort   int           `mapstructure:"METRICS_PORT"`
	EnableMetrics bool          `mapstructure:"ENABLE_METRICS"`
}

func Load() (*Config, error) {
	// Read configuration from environment variables
	cfg := Config{
		DatabaseURL:   parseString(os.Getenv("DATABASE_URL"), "postgresql://root:root@localhost:4566/dev"),
		ReadOnly:      parseBool(os.Getenv("READ_ONLY"), true),
		ReadOnlyMode:  parseBool(os.Getenv("READ_ONLY_MODE"), true),
		MaxConns:      parseInt32(os.Getenv("DB_MAX_CONNS"), 20),
		MinConns:      parseInt32(os.Getenv("DB_MIN_CONNS"), 5),
		ConnTimeout:   parseDuration(os.Getenv("DB_CONN_TIMEOUT"), 5*time.Second),
		QueryTimeout:  parseDuration(os.Getenv("QUERY_TIMEOUT"), 30*time.Second),
		MaxRows:       parseInt(os.Getenv("MAX_ROWS"), 10000),
		Transport:     parseString(os.Getenv("TRANSPORT"), "stdio"),
		HttpPort:      parseInt(os.Getenv("HTTP_PORT"), 8080),
		LogLevel:      parseString(os.Getenv("LOG_LEVEL"), "info"),
		MetricsPort:   parseInt(os.Getenv("METRICS_PORT"), 9090),
		EnableMetrics: parseBool(os.Getenv("ENABLE_METRICS"), true),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func parseBool(val string, defaultVal bool) bool {
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1" || val == "t" || val == "T" || val == "TRUE" || val == "True"
}

func parseInt32(val string, defaultVal int32) int32 {
	if val == "" {
		return defaultVal
	}
	var result int32
	fmt.Sscanf(val, "%d", &result)
	return result
}

func parseInt(val string, defaultVal int) int {
	if val == "" {
		return defaultVal
	}
	var result int
	fmt.Sscanf(val, "%d", &result)
	return result
}

func parseDuration(val string, defaultVal time.Duration) time.Duration {
	if val == "" {
		return defaultVal
	}
	result, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return result
}

func parseString(val string, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.MaxConns < c.MinConns {
		return fmt.Errorf("DB_MAX_CONNS must be >= DB_MIN_CONNS")
	}

	if c.ConnTimeout <= 0 {
		return fmt.Errorf("DB_CONN_TIMEOUT must be positive")
	}

	if c.QueryTimeout <= 0 {
		return fmt.Errorf("QUERY_TIMEOUT must be positive")
	}

	if c.MaxRows <= 0 {
		return fmt.Errorf("MAX_ROWS must be positive")
	}

	if c.ReadOnly || c.ReadOnlyMode {
		fmt.Println("WARNING: Server running in READ-ONLY mode - no mutations allowed")
	} else {
		return fmt.Errorf("READ_ONLY or READ_ONLY_MODE must be true - server cannot start in write mode")
	}

	validTransports := map[string]bool{"stdio": true, "streamable-http": true}
	if !validTransports[c.Transport] {
		return fmt.Errorf("invalid TRANSPORT: %s (must be 'stdio' or 'streamable-http')", c.Transport)
	}

	return nil
}
