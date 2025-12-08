package invoices

import (
	"testing"
	"time"
)

func TestInvoiceItemGetters(t *testing.T) {
	now := time.Now()
	item := InvoiceItem{
		SerializedWarrantyFlag: "Y",
		PostingDate1:           now,
		PostingDate2:           now,
		DiscountIndicator:      "D",
		ExtendedDescription:    "desc",
		SerialNumber:           "SN1",
	}

	if item.GetSerializedWarrantyFlag() != "Y" {
		t.Fatalf("expected serialized warranty flag Y")
	}
	if item.GetSerialNumber() != "SN1" {
		t.Fatalf("expected serial number SN1")
	}
	if ts := item.GetPostingDate1(); ts == nil || !ts.Equal(now) {
		t.Fatalf("expected posting date1 to match")
	}
	if ts := item.GetPostingDate2(); ts == nil || !ts.Equal(now) {
		t.Fatalf("expected posting date2 to match")
	}
	if item.GetDiscountIndicator() != "D" {
		t.Fatalf("expected discount indicator D")
	}
	if item.GetExtendedDescription() != "desc" {
		t.Fatalf("expected extended description")
	}
}
