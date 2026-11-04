package common

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func TestUUIDConversions(t *testing.T) {
	rawUUID := uuid.New()
	pgUUID := pgtype.UUID{Bytes: rawUUID, Valid: true}

	str := UUIDToString(pgUUID)
	if str != rawUUID.String() {
		t.Fatalf("expected %s, got %s", rawUUID.String(), str)
	}

	parsed, err := StringToUUID(str)
	if err != nil {
		t.Fatalf("unexpected error parsing UUID: %v", err)
	}
	if parsed != pgUUID {
		t.Fatalf("expected %v, got %v", pgUUID, parsed)
	}

	empty, err := StringToUUID("")
	if err != nil || empty.Valid {
		t.Fatalf("expected empty valid=false UUID, got %v", empty)
	}

	_, err = StringToUUID("not-a-uuid")
	if err == nil {
		t.Fatalf("expected error for invalid UUID string")
	}
}

func TestNumericAndFinancialPrecision(t *testing.T) {
	// Test that decimal values never lose precision through conversion
	cases := []string{
		"0.00",
		"100.50",
		"-45.99",
		"12345678.1234",
		"-0.01",
	}

	for _, c := range cases {
		num, err := StringToNumeric(c)
		if err != nil {
			t.Fatalf("failed to convert %q to numeric: %v", c, err)
		}
		str := NumericToString(num)
		d1, _ := decimal.NewFromString(c)
		d2, _ := decimal.NewFromString(str)
		if !d1.Equal(d2) {
			t.Fatalf("precision mismatch: input %s became %s", c, str)
		}
	}
}

func TestDateConversions(t *testing.T) {
	dateStr := "2026-11-04"
	pgDate, err := StringToDate(dateStr)
	if err != nil {
		t.Fatalf("failed to convert date string: %v", err)
	}
	outStr := DateToString(pgDate)
	if outStr != dateStr {
		t.Fatalf("expected %s, got %s", dateStr, outStr)
	}

	_, err = StringToDate("invalid-date")
	if err == nil {
		t.Fatalf("expected error for invalid date")
	}

	now := time.Now()
	tz := TimeToTimestamptz(now)
	if !tz.Valid || !tz.Time.Equal(now) {
		t.Fatalf("timestamptz conversion failed")
	}
}
