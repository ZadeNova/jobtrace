package config

import (
	"fmt"
	"os"
)

type Config struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresPort     string
	PostgresHost     string
}

func Load() (Config, error) {
	cfg := Config{
		PostgresUser:     os.Getenv("POSTGRES_USER"),
		PostgresPassword: os.Getenv("POSTGRES_PASSWORD"),
		PostgresDB:       os.Getenv("POSTGRES_DB"),
		PostgresPort:     os.Getenv("POSTGRES_PORT"),
		PostgresHost:     os.Getenv("POSTGRES_HOST"),
	}

	if cfg.PostgresUser == "" || cfg.PostgresPassword == "" || cfg.PostgresDB == "" || cfg.PostgresPort == "" || cfg.PostgresHost == "" {
		return Config{}, fmt.Errorf("missing required postgres env vars")
	}

	return cfg, nil
}
