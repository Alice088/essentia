package main

import (
	"Alice088/essentia/internal/app"
	"Alice088/essentia/pkg/env"
)

func main() {
	app.Run(new(env.Load("./.env")))
}
