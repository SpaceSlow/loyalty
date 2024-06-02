package config

import (
	"errors"
	"github.com/caarlos0/env"
)

type Config struct {
	DSN               string `env:"DATABASE_URI"`
	SecretKey         string `env:"SECRET_KEY"`
	PasswordIteration int
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

	return cfg, nil
}
