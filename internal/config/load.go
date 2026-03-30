package config

import (
	"log/slog"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func Load(path ...string) Config {
	err := godotenv.Load(path...)
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("Failed to load env", "err", err)
		os.Exit(1)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("Failed to parse env", "err", err)
		os.Exit(1)
	}

	return cfg
}
