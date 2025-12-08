package types

import (
	"testing"
	"time"
)

func TestUnitListParserDefaultsAndMappings(t *testing.T) {
	parser := UnitListParamParser{}
	if parser.GetDefaultSortBy() != "unit" {
		t.Fatalf("expected default sort to be unit")
	}
	if _, ok := parser.GetAllowedSortByColumns()["amount"]; !ok {
		t.Fatalf("expected amount sort mapping")
	}
	if _, ok := parser.GetAllowedFilterColumns()["productCode"]; !ok {
		t.Fatalf("expected productCode filter mapping")
	}
}

func TestUnitListParserParseValueDateAndNumber(t *testing.T) {
	parser := UnitListParamParser{}
	val, err := parser.ParseValue("2025-01-02", "indate")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := val.(time.Time); !ok {
		t.Fatalf("expected time.Time, got %T", val)
	}
	val, err = parser.ParseValue("12.50", "amount")
	if err != nil || val.(float64) != 12.50 {
		t.Fatalf("expected parsed float, got %v err=%v", val, err)
	}
}

func TestUnitListParserParseValueErrors(t *testing.T) {
	parser := UnitListParamParser{}
	if _, err := parser.ParseValue("bad", "year"); err == nil {
		t.Fatalf("expected error for invalid year")
	}
	if _, err := parser.ParseValue("", "model"); err == nil {
		t.Fatalf("expected error for empty string field")
	}
}
