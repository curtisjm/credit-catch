package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port         int           `envconfig:"PORT" default:"8080"`
	DatabaseURL  string        `envconfig:"DATABASE_URL" required:"true"`
	JWTSecret    string        `envconfig:"JWT_SECRET" required:"true"`
	JWTExpiry    time.Duration `envconfig:"JWT_EXPIRY" default:"24h"`
	Environment  string        `envconfig:"ENVIRONMENT" default:"development"`
	LogLevel     string        `envconfig:"LOG_LEVEL" default:"info"`
	PlaidEnv     string        `envconfig:"PLAID_ENV" default:"sandbox"`
	PlaidClient  string        `envconfig:"PLAID_CLIENT_ID"`
	PlaidSecret  string        `envconfig:"PLAID_SECRET"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	return &cfg, nil
}
