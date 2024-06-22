package config

import (
	"errors"
	"time"

	"github.com/caarlos0/env"
)

var ServerConfig *Config

type Config struct {
	DSN                   string `env:"DATABASE_URI"`
	SecretKey             string `env:"SECRET_KEY"`
	AccrualSystemAddress  string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	PasswordIterationsNum int
	TokenExpiredAt        time.Duration
	TimeoutOperation      time.Duration
	TimeoutServerShutdown time.Duration
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

	if cfg.AccrualSystemAddress == "" {
		if flagAccrualSystemAddress == "" {
			return nil, errors.New("flag error: needed accrual system address. check specification")
		}
		cfg.AccrualSystemAddress = flagAccrualSystemAddress
	}

	cfg.TimeoutOperation = 3 * time.Second
	cfg.PasswordIterationsNum = 500000
	cfg.TokenExpiredAt = time.Hour
	cfg.TimeoutServerShutdown = 10 * time.Second

	return cfg, nil
}
