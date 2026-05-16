package config

import (
	"fmt"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL        string        `mapstructure:"DATABASE_URL"`
	ReadOnly           bool          `mapstructure:"READ_ONLY"`
	ReadOnlyMode       bool          `mapstructure:"READ_ONLY_MODE"`
	MaxConns           int32         `mapstructure:"DB_MAX_CONNS"`
	MinConns           int32         `mapstructure:"DB_MIN_CONNS"`
	ConnTimeout        time.Duration `mapstructure:"DB_CONN_TIMEOUT"`
	QueryTimeout       time.Duration `mapstructure:"QUERY_TIMEOUT"`
	MaxRows            int           `mapstructure:"MAX_ROWS"`
	Transport          string        `mapstructure:"TRANSPORT"`
	HttpPort           int           `mapstructure:"HTTP_PORT"`
	LogLevel           string        `mapstructure:"LOG_LEVEL"`
	MetricsPort        int           `mapstructure:"METRICS_PORT"`
	EnableMetrics      bool          `mapstructure:"ENABLE_METRICS"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetConfigType("env")

	v.AutomaticEnv()

	v.SetDefault("READ_ONLY", true)
	v.SetDefault("READ_ONLY_MODE", true)
	v.SetDefault("DB_MAX_CONNS", 20)
	v.SetDefault("DB_MIN_CONNS", 5)
	v.SetDefault("DB_CONN_TIMEOUT", "5s")
	v.SetDefault("QUERY_TIMEOUT", "30s")
	v.SetDefault("MAX_ROWS", 10000)
	v.SetDefault("TRANSPORT", "stdio")
	v.SetDefault("HTTP_PORT", 8080)
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("METRICS_PORT", 9090)
	v.SetDefault("ENABLE_METRICS", true)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.Unmarshal(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to reload config: %v\n", err)
		}
	})

	return &cfg, nil
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