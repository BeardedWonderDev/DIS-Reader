package types

import "testing"

func TestPartListParserDefaults(t *testing.T) {
	parser := PartListParamParser{}
	if parser.GetDefaultSortBy() != "partNumber" {
		t.Fatalf("expected default sort to be partNumber")
	}
	if _, ok := parser.GetAllowedSortByColumns()["price"]; !ok {
		t.Fatalf("expected price sort mapping to exist")
	}
	if _, ok := parser.GetAllowedFilterColumns()["vendor"]; !ok {
		t.Fatalf("expected vendor filter mapping to exist")
	}
}

func TestPartListParserParseValueNumericAndErrors(t *testing.T) {
	parser := PartListParamParser{}
	val, err := parser.ParseValue("12345", "PIMDAD")
	if err != nil || val.(int) != 12345 {
		t.Fatalf("expected numeric parse, got val=%v err=%v", val, err)
	}
	if _, err := parser.ParseValue("", "PIMDAD"); err == nil {
		t.Fatalf("expected error for empty numeric field")
	}
	if _, err := parser.ParseValue("x.y", "PIMAVG"); err == nil {
		t.Fatalf("expected error for invalid decimal")
	}
}

func TestPartListParserParseValueFallsBackToRaw(t *testing.T) {
	parser := PartListParamParser{}
	val, err := parser.ParseValue("abc", "PIMVEN")
	if err != nil || val.(string) != "abc" {
		t.Fatalf("expected raw string, got val=%v err=%v", val, err)
	}
}
