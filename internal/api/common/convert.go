package common

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func UUIDToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

func StringToUUID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	parsed, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID %q: %w", s, err)
	}
	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

func MustStringToUUID(s string) pgtype.UUID {
	u, _ := StringToUUID(s)
	return u
}

func NumericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	v, err := n.Value()
	if err != nil || v == nil {
		return "0.00"
	}
	return fmt.Sprintf("%v", v)
}

func StringToNumeric(s string) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if s == "" {
		s = "0"
	}
	err := n.Scan(s)
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("invalid numeric %q: %w", s, err)
	}
	return n, nil
}

func DateToString(d pgtype.Date) string {
	if !d.Valid {
		return ""
	}
	return d.Time.Format("2006-01-02")
}

func StringToDate(s string) (pgtype.Date, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, fmt.Errorf("invalid date %q (expected YYYY-MM-DD): %w", s, err)
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

func TimeToTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
