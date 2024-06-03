package config

import (
	"errors"
	"time"

	"github.com/caarlos0/env"
)

var ServerConfig *Config

type Config struct {
	DSN               string `env:"DATABASE_URI"`
	SecretKey         string `env:"SECRET_KEY"`
	PasswordIteration int
	TokenExpiredAt    time.Duration
}

func GetConfigWithFlags() (*Config, error) {
	parseFlags()
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if cfg.DSN == "" {
		if flagDSN == "" {
			return nil, errors.New("flag error: needed DSN. check specification")
		}
		cfg.DSN = flagDSN
	}

	if cfg.SecretKey == "" {
		cfg.SecretKey = flagSecretKey
	}

	cfg.PasswordIteration = 500000
	cfg.TokenExpiredAt = time.Hour

	return cfg, nil
}
