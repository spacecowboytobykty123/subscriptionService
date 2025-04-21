package config

import (
	"time"
)

type Config struct {
	env string
	db  struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	TokenTTL time.Duration
}

type GRPCConfig struct {
	Port    int
	Timeout time.Duration
}
