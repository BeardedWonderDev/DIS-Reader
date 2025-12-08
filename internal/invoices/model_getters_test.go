package invoices

import (
	"testing"
	"time"
)

func TestInvoiceItemGetters(t *testing.T) {
	now := time.Now()
	item := InvoiceItem{
		Division:               "01",
		OrderSource:            "WEB",
		CustomerNumber:         "C123",
		VendorCode:             "V1",
		PartNumber:             "P1",
		PostingDate:            now,
		DocumentNumber:         "DOC1",
		LineID:                 "L1",
		FormatType:             "FT",
		ExceptionCode:          "EX",
		Classification:         "CL",
		Description:            "Desc",
		PriceCode:              "PC",
		Cost:                   10.5,
		Location:               "LOC",
		Quantity:               2,
		Price:                  5.5,
		TaxCode:                "TX",
		DiscountRate:           1.2,
		Memo:                   " memo ",
		RebillFlag:             "Y",
		PostedFlag:             "N",
		StockedFlag:            "yes",
		WarrantyFlag:           "T",
		SerializedWarrantyFlag: "Y",
		SerialNumber:           "SN1",
		Division1:              "D1",
		VendorCode1:            "VV",
		PostingDate1:           now,
		Division2:              "D2",
		PostingDate2:           now,
		DiscountIndicator:      "D",
		ExtendedDescription:    "desc",
		ProcessCode:            "PROC",
		DemandCode:             "DM",
		SeasonalCode:           "SEAS",
		FreightTaxCode:         "FTX",
		OriginFlag:             "ORIG",
		UserField1:             "UF1",
		UserField2:             "UF2",
	}

	if item.GetDivision() != "01" || item.GetOrderSource() != "WEB" || item.GetCustomerNumber() != "C123" {
		t.Fatalf("basic getters mismatch")
	}
	if item.GetVendorCode() != "V1" || item.GetPartNumber() != "P1" || item.GetDocumentNumber() != "DOC1" || item.GetLineID() != "L1" {
		t.Fatalf("doc/line getters mismatch")
	}
	if item.GetFormatType() != "FT" || item.GetExceptionCode() != "EX" || item.GetClassification() != "CL" || item.GetDescription() != "Desc" {
		t.Fatalf("format/exception/classification/description mismatch")
	}
	if item.GetPriceCode() != "PC" || item.GetCost() != 10.5 || item.GetLocation() != "LOC" || item.GetQuantity() != 2 || item.GetTaxCode() != "TX" {
		t.Fatalf("price/location/tax mismatch")
	}
	if item.GetDiscountRate() != 1.2 || item.GetMemo() != " memo " {
		t.Fatalf("discount/memo mismatch")
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
	if item.GetProcessCode() != "PROC" || item.GetDemandCode() != "DM" || item.GetSeasonalCode() != "SEAS" || item.GetFreightTaxCode() != "FTX" || item.GetOriginFlag() != "ORIG" {
		t.Fatalf("expected process/demand/seasonal/freight/origin getters")
	}
	if item.GetUserField1() != "UF1" || item.GetUserField2() != "UF2" {
		t.Fatalf("expected user fields")
	}

	spec := item.ToInvoiceItemSpec()
	if !spec.Rebill || spec.Posted || !spec.Stocked || !spec.Warranty || spec.Memo == nil || *spec.Memo != "memo" {
		t.Fatalf("flag/optional mapping mismatch in spec: %+v", spec)
	}
}

func TestOptionalHelpersAndFlagToBool(t *testing.T) {
	zero := time.Time{}
	if optionalTime(zero) != nil {
		t.Fatalf("expected nil for zero time")
	}
	ts := time.Unix(10, 0)
	if got := optionalTime(ts); got == nil || !got.Equal(ts) {
		t.Fatalf("expected optional time to return pointer")
	}
	if optionalString("   ") != nil {
		t.Fatalf("expected nil for blank string")
	}
	if got := optionalString(" x "); got == nil || *got != "x" {
		t.Fatalf("expected trimmed optional string")
	}
	if got := optionalIntFromFloat(3.9); got == nil || *got != 3 {
		t.Fatalf("expected int floor conversion, got %v", got)
	}
	if got := optionalIntFromFloat(0); got != nil {
		t.Fatalf("expected nil for zero value")
	}
	if !flagToBool("yes") || !flagToBool(" 1") || flagToBool("n") {
		t.Fatalf("flagToBool did not normalize values correctly")
	}
}
