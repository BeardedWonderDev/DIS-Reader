package types

import (
	"testing"
	"time"
)

func TestInvoiceListParamParserMetadata(t *testing.T) {
	parser := InvoiceListParamParser{}
	if got := parser.GetDefaultSortBy(); got != "postingDate" {
		t.Fatalf("default sort mismatch: got %q", got)
	}
	sorts := parser.GetAllowedSortByColumns()
	if sorts["postingDate"] != "AHPDTE" || sorts["customer"] != "AHAS#" {
		t.Fatalf("unexpected sort map entries: %v", sorts)
	}
	filters := parser.GetAllowedFilterColumns()
	if filters["orderSource"] != "AHAORS" || filters["priceCode"] != "AHPCDE" {
		t.Fatalf("unexpected filter map entries: %v", filters)
	}
}

func TestInvoiceListParamParserParseValue(t *testing.T) {
	parser := InvoiceListParamParser{}
	date, err := parser.ParseValue("2025-01-02", "postingDate")
	if err != nil {
		t.Fatalf("unexpected error parsing date: %v", err)
	}
	if _, ok := date.(time.Time); !ok {
		t.Fatalf("expected time.Time for postingDate")
	}
	num, err := parser.ParseValue("12.5", "price")
	if err != nil || num.(float64) != 12.5 {
		t.Fatalf("expected parsed float 12.5, got %v (err=%v)", num, err)
	}
	if _, err := parser.ParseValue("", "part"); err == nil {
		t.Fatalf("expected error for empty string value")
	}
}

func TestWholeGoodsInvoiceListParamParserMetadata(t *testing.T) {
	parser := WholeGoodsInvoiceListParamParser{}
	if got := parser.GetDefaultSortBy(); got != "DSHVA" {
		t.Fatalf("default sort mismatch: got %q", got)
	}
	sorts := parser.GetAllowedSortByColumns()
	if sorts["invoiceDate"] != "DSHVA" || sorts["lineItem"] != "DS1XQA" {
		t.Fatalf("unexpected sort map entries: %v", sorts)
	}
	filters := parser.GetAllowedFilterColumns()
	if filters["description"] != "DSHXA" || filters["soldBy"] != "DS22UA" {
		t.Fatalf("unexpected filter map entries: %v", filters)
	}
}

func TestWholeGoodsInvoiceListParamParserParseValue(t *testing.T) {
	parser := WholeGoodsInvoiceListParamParser{}
	date, err := parser.ParseValue("2025-02-03", "invoiceDate")
	if err != nil {
		t.Fatalf("unexpected error parsing date: %v", err)
	}
	if _, ok := date.(time.Time); !ok {
		t.Fatalf("expected time.Time for invoiceDate")
	}
	intVal, err := parser.ParseValue("42", "lineItem")
	if err != nil || intVal.(int) != 42 {
		t.Fatalf("expected parsed int 42, got %v (err=%v)", intVal, err)
	}
	if _, err := parser.ParseValue("", "division"); err == nil {
		t.Fatalf("expected error for empty division value")
	}
}
