package main

import (
	"Alice088/pdf-summarize/internal/app"
	"Alice088/pdf-summarize/pkg/env"
)

func main() {
	app.Run(new(env.Load("./.env")))
}
