package invoices

import (
	"testing"
	"time"
)

func TestWholeGoodsInvoiceToSpec(t *testing.T) {
	date := time.Unix(1000, 0)
	wg := WholeGoodsInvoice{
		Division:        "01",
		SequenceCode:    1,
		LineItemNumber:  2,
		InvoiceNumber:   "INV1",
		StatusCode:      "S",
		InvoiceDate:     date,
		LineItemPrice:   9.5,
		LineDescription: "desc",
		ReferenceValue:  3,
		SoldBy:          "Bob",
	}

	spec := wg.ToWholeGoodsInvoiceSpec()
	if spec.InvoiceNumber != "INV1" || spec.LineItemNumber != 2 || spec.Division != "01" {
		t.Fatalf("expected core fields mapped, got %+v", spec)
	}
	if !spec.InvoiceDate.Equal(date) {
		t.Fatalf("expected invoice date set")
	}
	if spec.LegacyDS9YA == nil || *spec.LegacyDS9YA != 1 {
		t.Fatalf("expected legacy sequence code set")
	}
	if spec.LegacyDSYGA == nil || *spec.LegacyDSYGA != "S" {
		t.Fatalf("expected legacy status set")
	}
	if spec.LegacyDSWVA == nil || *spec.LegacyDSWVA != 3 {
		t.Fatalf("expected legacy reference set")
	}
	if spec.SoldBy != "Bob" {
		t.Fatalf("expected SoldBy populated")
	}
}
