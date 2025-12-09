package unit

import (
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestUnitGettersAndToSpec(t *testing.T) {
	now := time.Now().UTC()
	horse := 150
	engHours := 12.5
	soldAt := now.AddDate(0, 0, -1)
	invoiceRaw := "INV123 #*"

	u := Unit{
		UnitID:          "U1",
		Year:            "2024",
		Make:            "Kubota",
		Model:           "KX040",
		Description:     "Excavator",
		Serial:          ptr("S123"),
		Engine:          ptr("E123"),
		Status:          "ACTIVE",
		ProductCode:     ptr("PC"),
		IsNew:           "n",
		Color:           ptr("Orange"),
		HorsePower:      &horse,
		EngineHours:     &engHours,
		Location:        "LOC1",
		Account:         "ACCT1",
		CreatedAt:       now,
		SoldAt:          &soldAt,
		InvoiceNumber:   &invoiceRaw,
		SoldAccount:     ptr("CUST1"),
		RevenueAmount:   ptr(1234.5),
		Cost:            999.9,
		SoldBy:          ptr("EMP1"),
		SoldByName:      ptr("Alice"),
		WarrantyCode:    ptr("W1"),
		TradedOnUnit:    ptr("TRADE1"),
		GLAccount:       ptr("GL1"),
		SuggestedList:   ptr(2000.0),
		DealerList:      ptr(1800.0),
		FlooringDueDate: &now,
		FlooringAmount:  ptr(500.0),
		RentalStartDate: &now,
		RentalRevenue:   ptr(50.0),
		RentalCost:      ptr(25.0),
		SoldToName:      ptr("Bob"),
		SoldToAddress:   ptr("123 Main"),
		SoldToCity:      ptr("Town"),
		SoldToZip:       ptr("90210"),
		PhoneNumber:     ptr("555-1212"),
	}

	if !u.GetIsNew() {
		t.Fatalf("expected GetIsNew to treat 'n' as new")
	}
	if got := *u.GetInvoiceNumber(); got != "INV123" {
		t.Fatalf("expected sanitized invoice number, got %s", got)
	}
	if got := u.GetHorsePower(); got == nil || *got != horse {
		t.Fatalf("expected horsepower pointer")
	}
	if got := u.GetEngineHours(); got == nil || *got != engHours {
		t.Fatalf("expected engine hours pointer")
	}

	spec := u.ToUnitSpec()
	if spec.UnitID != "U1" || !spec.New || spec.InvoiceNumber == nil || *spec.InvoiceNumber != "INV123" {
		t.Fatalf("spec fields not mapped correctly: %+v", spec)
	}
	if spec.SoldAt == nil || !spec.SoldAt.Equal(soldAt) {
		t.Fatalf("expected SoldAt propagated")
	}
	if spec.SoldByName == nil || *spec.SoldByName != "Alice" {
		t.Fatalf("expected SoldByName in spec")
	}
}
