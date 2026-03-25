package postgres

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func toNullString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromNullString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func toNullTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}
