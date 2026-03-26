package dependencies

import (
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"Alice088/essentia/pkg/env"
	"Alice088/essentia/pkg/s3"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AppDeps struct {
	Logger  *slog.Logger
	Queries *queries.Queries
	S3      s3.S3
	Config  env.Config
	DB      *pgxpool.Pool
}
