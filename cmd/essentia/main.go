package main

import (
	"Alice088/essentia/internal/app"
	"Alice088/essentia/internal/config"
)

func main() {
	app.Run(config.Load("./.env"))
}
