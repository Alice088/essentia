package logging

import (
	"log/slog"
	"os"
)

func CreateLogFile(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		f, err = os.Create(path)
		if err != nil {
			slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("Failed create log file")
			os.Exit(1)
		}

	}
	return f
}
