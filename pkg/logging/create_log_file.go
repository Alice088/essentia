package logging

import (
	"log"
	"log/slog"
	"os"
)

func CreateLogFile() *os.File {
	path := "./logs/"

	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(
		path+"logs.json",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		f, err = os.Create(path + "logs.json")
		if err != nil {
			slog.New(slog.NewTextHandler(os.Stderr, nil)).Error("Failed create log file", "err", err)
			os.Exit(1)
		}

	}
	return f
}
