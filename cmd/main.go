package main

import (
	"Alice088/ai-greed/pkg/env"
	"Alice088/ai-greed/pkg/logging"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	slogchi "github.com/samber/slog-chi"
)

func main() {
	cfg := env.Load("./.env")
	file := logging.CreateLogFile()
	logger := slog.New(slog.NewJSONHandler(file, nil))

	r := chi.NewRouter()
	r.Use(slogchi.New(logger))
	r.Use(middleware.Recoverer)
	// r.Use(middleware.Compress())
	r.Use(middleware.Throttle(cfg.MAXUSER))
	r.Use(middleware.Timeout(cfg.TIMEOUT))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(":3000", r)
}
