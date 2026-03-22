package sqlc

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToUUID(target uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: target,
		Valid: true,
	}
}

func ToTEXT(target string) pgtype.Text {
	return pgtype.Text{
		String: target,
		Valid:  true,
	}
}
