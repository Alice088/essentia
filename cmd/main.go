package main

import (
	"log/slog"
	
)

func main() {
	CreateLogFile
	slog.NewJSONHandler()
}
