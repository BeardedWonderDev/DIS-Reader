package invoices

import (
	"testing"
	"time"
)

func strPtr(s string) *string { return &s }

func TestInvoiceItemGettersAndSpec(t *testing.T) {
	ts := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	alt := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	item := InvoiceItem{
		Division:               "01 ",
		OrderSource:            "WEB",
		CustomerNumber:         "C123",
		VendorCode:             "V1",
		PartNumber:             "P1",
		PostingDate:            ts,
		DocumentNumber:         " DOC1 ",
		LineID:                 " 1 ",
		FormatType:             "F",
		ExceptionCode:          "E",
		Classification:         "CLASS",
		Description:            " Desc ",
		PriceCode:              "PC",
		Cost:                   10.5,
		Location:               " LOC ",
		Quantity:               2,
		Price:                  5.5,
		TaxCode:                "T",
		DiscountRate:           0.1,
		Memo:                   "  note ",
		RebillFlag:             "Y",
		PostedFlag:             "N",
		StockedFlag:            "Y",
		WarrantyFlag:           "Y",
		SerializedWarrantyFlag: "N",
		SerialNumber:           " SN ",
		Division1:              "02",
		VendorCode1:            "V2",
		PostingDate1:           ts,
		Division2:              "03",
		PostingDate2:           alt,
		DiscountIndicator:      "DI",
		ExtendedDescription:    " long ",
		ProcessCode:            "PR",
		DemandCode:             "DM",
		SeasonalCode:           "SC",
		FreightTaxCode:         "FT",
		OriginFlag:             "OR",
		UserField1:             "UF1",
		UserField2:             "UF2",
	}

	if item.GetDivision() != "01 " || item.GetPrice() != 5.5 || item.GetPostingDate() != ts {
		t.Fatalf("basic getters not returning fields")
	}
	if item.GetPostingDate1() == nil || !item.GetPostingDate1().Equal(ts) {
		t.Fatalf("expected PostingDate1 pointer")
	}

	spec := item.ToInvoiceItemSpec()
	if spec.DocumentNumber != "DOC1" || spec.LineID != "1" || spec.Description != "Desc" {
		t.Fatalf("expected trimmed core fields, got %+v", spec)
	}
	if spec.Rebill != true || spec.Posted != false || spec.Stocked != true || spec.Warranty != true || spec.SerializedWarranty != false {
		t.Fatalf("flag conversions incorrect: %+v", spec)
	}
	if spec.Memo == nil || *spec.Memo != "note" {
		t.Fatalf("expected memo optional string trimmed")
	}
	if spec.AuditPostingDate == nil || !spec.AuditPostingDate.Equal(ts) {
		t.Fatalf("expected audit posting date")
	}
	if spec.AlternatePostingDate == nil || !spec.AlternatePostingDate.Equal(alt) {
		t.Fatalf("expected alternate posting date")
	}
	if spec.SerialNumber == nil || *spec.SerialNumber != "SN" {
		t.Fatalf("expected serial number optional string trimmed")
	}
}
