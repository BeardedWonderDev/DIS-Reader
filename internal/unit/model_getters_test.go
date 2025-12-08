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

func TestUnitSimpleGetters(t *testing.T) {
	serial := "S2"
	engine := "E2"
	product := "P2"
	color := "Blue"
	account := "ACCT2"
	location := "LOC2"
	status := "INACTIVE"
	soldAcct := "CUST2"
	soldBy := "EMP2"
	soldByName := "Bob"
	warranty := "W2"
	traded := "OLDUNIT"
	gl := "GL2"
	city := "BigCity"
	zip := "11111"
	soldToName := "Buyer2"
	addr := "456 Ave"
	phone := "555-2222"
	hp := 55
	engHours := 7.5
	rev := 22.2
	cost := 3.3
	floorAmt := 9.9
	rentCost := 4.4
	created := time.Unix(1000, 0)
	soldAt := time.Unix(2000, 0)
	floorDue := time.Unix(3000, 0)
	rentStart := time.Unix(4000, 0)

	u := Unit{
		UnitID:          "U2",
		Year:            "2020",
		Make:            "Make2",
		Model:           "Model2",
		Description:     "Desc2",
		Serial:          &serial,
		Engine:          &engine,
		Status:          status,
		ProductCode:     &product,
		Color:           &color,
		HorsePower:      &hp,
		EngineHours:     &engHours,
		Location:        location,
		Account:         account,
		CreatedAt:       created,
		SoldAt:          &soldAt,
		SoldAccount:     &soldAcct,
		RevenueAmount:   &rev,
		Cost:            cost,
		SoldBy:          &soldBy,
		SoldByName:      &soldByName,
		WarrantyCode:    &warranty,
		TradedOnUnit:    &traded,
		GLAccount:       &gl,
		FlooringDueDate: &floorDue,
		FlooringAmount:  &floorAmt,
		RentalStartDate: &rentStart,
		RentalCost:      &rentCost,
		SoldToName:      &soldToName,
		SoldToAddress:   &addr,
		SoldToCity:      &city,
		SoldToZip:       &zip,
		PhoneNumber:     &phone,
	}

	if u.GetUnitID() != "U2" || u.GetYear() != "2020" || u.GetMake() != "Make2" || u.GetModel() != "Model2" || u.GetDescription() != "Desc2" {
		t.Fatalf("basic string getters mismatch")
	}
	if u.GetSerial() != &serial || u.GetEngine() != &engine || u.GetProductCode() != &product || u.GetColor() != &color {
		t.Fatalf("pointer getters mismatch")
	}
	if u.GetStatus() != status || u.GetLocation() != location || u.GetAccount() != account {
		t.Fatalf("status/location/account mismatch")
	}
	if !u.GetCreatedAt().Equal(created) || u.GetSoldAt() == nil || !u.GetSoldAt().Equal(soldAt) {
		t.Fatalf("created/sold at mismatch")
	}
	if u.GetSoldAccount() != &soldAcct || u.GetSoldTo() != &soldAcct {
		t.Fatalf("sold account getter mismatch")
	}
	if u.GetRevenueAmount() != &rev || u.GetCost() != cost {
		t.Fatalf("revenue/cost mismatch")
	}
	if u.GetSoldBy() != &soldBy || u.GetSoldByName() != &soldByName {
		t.Fatalf("sold by getters mismatch")
	}
	if u.GetWarrantyCode() != &warranty || u.GetTradedOnUnit() != &traded || u.GetGLAccount() != &gl {
		t.Fatalf("warranty/traded/gl mismatch")
	}
	if u.GetFlooringDueDate() != &floorDue || u.GetFlooringAmount() == nil || *u.GetFlooringAmount() != floorAmt {
		t.Fatalf("flooring getters mismatch")
	}
	if u.GetRentalStartDate() != &rentStart || u.GetRentalCost() == nil || *u.GetRentalCost() != rentCost {
		t.Fatalf("rental getters mismatch")
	}
	if u.GetSoldToName() != &soldToName || u.GetSoldToCity() != &city || u.GetSoldToZip() != &zip || u.GetSoldToAddress() != &addr {
		t.Fatalf("sold-to getters mismatch")
	}
	if u.GetPhoneNumber() != &phone {
		t.Fatalf("phone getter mismatch")
	}
	if u.GetHorsePower() == nil || *u.GetHorsePower() != hp || u.GetEngineHours() == nil || *u.GetEngineHours() != engHours {
		t.Fatalf("power/engine hours mismatch")
	}
	if inv := u.GetInvoiceNumber(); inv != nil {
		t.Fatalf("expected nil invoice when not set, got %v", inv)
	}
	if u.GetRentalRevenue() != nil {
		t.Fatalf("expected nil rental revenue when unset")
	}
}
