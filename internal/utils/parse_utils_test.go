package utils

import (
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	t.Run("YYYYMM", func(t *testing.T) {
		got, err := ParseDate("202407")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expect := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
		if !got.Equal(expect) {
			t.Fatalf("expected %v, got %v", expect, got)
		}
	})

	t.Run("YYYYMMDD", func(t *testing.T) {
		got, err := ParseDate("20240715")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expect := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
		if !got.Equal(expect) {
			t.Fatalf("expected %v, got %v", expect, got)
		}
	})

	t.Run("invalid length returns zero time", func(t *testing.T) {
		got, err := ParseDate("2024")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if !got.IsZero() {
			t.Fatalf("expected zero time for invalid input, got %v", got)
		}
	})
}

func TestParseIntAndFloat(t *testing.T) {
	if ParseInt("12") != 12 {
		t.Fatalf("expected 12")
	}
	if ParseInt("bogus") != 0 {
		t.Fatalf("expected 0 for invalid int input")
	}
	if ParseFloat("1.5") != 1.5 {
		t.Fatalf("expected 1.5")
	}
	if ParseFloat("bogus") != 0 {
		t.Fatalf("expected 0 for invalid float input")
	}
}
