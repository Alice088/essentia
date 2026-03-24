package dependencies

import (
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"Alice088/essentia/pkg/env"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
)

type AppDeps struct {
	Logger  *slog.Logger
	Queries *queries.Queries
	MinIO   *minio.Client
	Config  *env.Config
	DB      *pgxpool.Pool
}
