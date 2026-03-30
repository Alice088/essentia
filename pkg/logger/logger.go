package xlogger

import (
	"Alice088/essentia/internal/config"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func New(cfg config.Config) *(slog.Logger) {
	logRotator := &lumberjack.Logger{
		Filename:   "./logs/logs.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	var loggerLever slog.Level
	if cfg.Env == "dev" {
		loggerLever = slog.LevelDebug
	} else {
		loggerLever = slog.LevelInfo
	}

	mw := slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, logRotator),
		&slog.HandlerOptions{
			Level: loggerLever,
		},
	)

	return slog.New(mw)
}
