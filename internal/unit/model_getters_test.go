package unit

import (
	"testing"
	"time"
)

func TestUnitGetterFields(t *testing.T) {
	serial := "s1"
	engine := "e1"
	product := "p1"
	hp := 42
	hrs := 12.5
	soldAt := time.Now()
	inv := "  INV123 "
	revenue := 99.5
	cost := 10.1
	suggested := 50.0
	dealer := 45.0
	floorDue := time.Now().Add(24 * time.Hour)
	floorAmt := 11.0
	rentStart := time.Now().Add(-24 * time.Hour)
	rentRev := 5.5
	rentCost := 1.1
	soldTo := "acct1"
	soldBy := "user1"
	soldByName := "User One"
	soldToName := "Buyer"
	soldToAddr := "123 St"
	soldToCity := "City"
	soldToZip := "12345"
	phone := "555-0100"

	u := Unit{
		UnitID:          "u1",
		Year:            "2024",
		Make:            "make",
		Model:           "model",
		Description:     "desc",
		Serial:          &serial,
		Engine:          &engine,
		Status:          "ACTIVE",
		ProductCode:     &product,
		IsNew:           "N",
		Color:           nil,
		HorsePower:      &hp,
		EngineHours:     &hrs,
		Location:        "loc",
		Account:         "acct",
		CreatedAt:       time.Now(),
		SoldAt:          &soldAt,
		InvoiceNumber:   &inv,
		SoldAccount:     &soldTo,
		RevenueAmount:   &revenue,
		Cost:            cost,
		SoldBy:          &soldBy,
		SoldByName:      &soldByName,
		WarrantyCode:    nil,
		TradedOnUnit:    nil,
		GLAccount:       nil,
		SuggestedList:   &suggested,
		DealerList:      &dealer,
		FlooringDueDate: &floorDue,
		FlooringAmount:  &floorAmt,
		RentalStartDate: &rentStart,
		RentalRevenue:   &rentRev,
		RentalCost:      &rentCost,
		SoldToName:      &soldToName,
		SoldToAddress:   &soldToAddr,
		SoldToCity:      &soldToCity,
		SoldToZip:       &soldToZip,
		PhoneNumber:     &phone,
	}

	if !u.GetIsNew() {
		t.Fatalf("expected GetIsNew true when IsNew=='N'")
	}
	if got := *u.GetInvoiceNumber(); got != "INV123" { // whitespace trimmed by SanitizeInvoiceNumber
		t.Fatalf("expected sanitized invoice, got %q", got)
	}
	if u.GetSuggestedList() == nil || *u.GetSuggestedList() != suggested {
		t.Fatalf("suggested list mismatch")
	}
	if u.GetDealerList() == nil || *u.GetDealerList() != dealer {
		t.Fatalf("dealer list mismatch")
	}
	if u.GetFlooringAmount() == nil || *u.GetFlooringAmount() != floorAmt {
		t.Fatalf("floor amount mismatch")
	}
	if u.GetRentalRevenue() == nil || *u.GetRentalRevenue() != rentRev {
		t.Fatalf("rental revenue mismatch")
	}
	if u.GetSoldToAddress() == nil || *u.GetSoldToAddress() != soldToAddr {
		t.Fatalf("sold-to address mismatch")
	}
	if u.GetPhoneNumber() == nil || *u.GetPhoneNumber() != phone {
		t.Fatalf("phone mismatch")
	}
}
