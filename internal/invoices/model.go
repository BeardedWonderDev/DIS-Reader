package invoices

import (
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Model represents the minimal behavior every invoice item must expose.
type Model interface {
	GetDivision() string
	GetOrderSource() string
	GetCustomerNumber() string
	GetVendorCode() string
	GetPartNumber() string
	GetPostingDate() time.Time
	GetDocumentNumber() string
	GetLineID() string
	GetFormatType() string
	GetExceptionCode() string
	GetClassification() string
	GetDescription() string
	GetPriceCode() string
	GetCost() float64
	GetLocation() string
	GetQuantity() float64
	GetPrice() float64
	GetTaxCode() string
	GetDiscountRate() float64
	GetMemo() string
	GetRebillFlag() string
	GetPostedFlag() string
	GetStockedFlag() string
	GetWarrantyFlag() string
	GetSerializedWarrantyFlag() string
	GetSerialNumber() string
	GetDivision1() string
	GetVendorCode1() string
	GetPostingDate1() *time.Time
	GetDivision2() string
	GetPostingDate2() *time.Time
	GetDiscountIndicator() string
	GetExtendedDescription() string
	GetProcessCode() string
	GetDemandCode() string
	GetSeasonalCode() string
	GetFreightTaxCode() string
	GetOriginFlag() string
	GetUserField1() string
	GetUserField2() string
	ToInvoiceItemSpec() *types.InvoiceItemSpec
}

// InvoiceItem maps the FILEC.IAH table into Go types via mapstructure tags.
type InvoiceItem struct {
	Division               string    `mapstructure:"AHDIV"`
	OrderSource            string    `mapstructure:"AHAORS"`
	CustomerNumber         string    `mapstructure:"AHAS#"`
	VendorCode             string    `mapstructure:"AHVEND"`
	PartNumber             string    `mapstructure:"AHPART"`
	PostingDate            time.Time `mapstructure:"AHPDTE"`
	DocumentNumber         string    `mapstructure:"AHDOC"`
	LineID                 string    `mapstructure:"AHLINE"`
	FormatType             string    `mapstructure:"AHFRMT"`
	ExceptionCode          string    `mapstructure:"AHECDE"`
	Classification         string    `mapstructure:"AHCLSS"`
	Description            string    `mapstructure:"AHDESC"`
	PriceCode              string    `mapstructure:"AHPCDE"`
	Cost                   float64   `mapstructure:"AHCOST"`
	Location               string    `mapstructure:"AHLOC"`
	Quantity               float64   `mapstructure:"AHQTY"`
	Price                  float64   `mapstructure:"AHPRCE"`
	TaxCode                string    `mapstructure:"AHTAX"`
	DiscountRate           float64   `mapstructure:"AHDRTE"`
	Memo                   string    `mapstructure:"AHMEMO"`
	RebillFlag             string    `mapstructure:"AHRBY"`
	PostedFlag             string    `mapstructure:"AHPOST"`
	StockedFlag            string    `mapstructure:"AHSTOK"`
	WarrantyFlag           string    `mapstructure:"AHWARR"`
	SerializedWarrantyFlag string    `mapstructure:"AHSWAR"`
	SerialNumber           string    `mapstructure:"AHSERL"`
	Division1              string    `mapstructure:"AHDIV1"`
	VendorCode1            string    `mapstructure:"AHVND1"`
	PostingDate1           time.Time `mapstructure:"AHPDT1"`
	Division2              string    `mapstructure:"AHDIV2"`
	PostingDate2           time.Time `mapstructure:"AHPDT2"`
	DiscountIndicator      string    `mapstructure:"AHDISC"`
	ExtendedDescription    string    `mapstructure:"AHSPRT"`
	ProcessCode            string    `mapstructure:"AHPROC"`
	DemandCode             string    `mapstructure:"AHDMND"`
	SeasonalCode           string    `mapstructure:"AHSEA"`
	FreightTaxCode         string    `mapstructure:"AHFTAX"`
	OriginFlag             string    `mapstructure:"AHWORI"`
	UserField1             string    `mapstructure:"AHFIL1"`
	UserField2             string    `mapstructure:"AHFIL2"`
}

func (i InvoiceItem) GetDivision() string               { return i.Division }
func (i InvoiceItem) GetOrderSource() string            { return i.OrderSource }
func (i InvoiceItem) GetCustomerNumber() string         { return i.CustomerNumber }
func (i InvoiceItem) GetVendorCode() string             { return i.VendorCode }
func (i InvoiceItem) GetPartNumber() string             { return i.PartNumber }
func (i InvoiceItem) GetPostingDate() time.Time         { return i.PostingDate }
func (i InvoiceItem) GetDocumentNumber() string         { return i.DocumentNumber }
func (i InvoiceItem) GetLineID() string                 { return i.LineID }
func (i InvoiceItem) GetFormatType() string             { return i.FormatType }
func (i InvoiceItem) GetExceptionCode() string          { return i.ExceptionCode }
func (i InvoiceItem) GetClassification() string         { return i.Classification }
func (i InvoiceItem) GetDescription() string            { return i.Description }
func (i InvoiceItem) GetPriceCode() string              { return i.PriceCode }
func (i InvoiceItem) GetCost() float64                  { return i.Cost }
func (i InvoiceItem) GetLocation() string               { return i.Location }
func (i InvoiceItem) GetQuantity() float64              { return i.Quantity }
func (i InvoiceItem) GetPrice() float64                 { return i.Price }
func (i InvoiceItem) GetTaxCode() string                { return i.TaxCode }
func (i InvoiceItem) GetDiscountRate() float64          { return i.DiscountRate }
func (i InvoiceItem) GetMemo() string                   { return i.Memo }
func (i InvoiceItem) GetRebillFlag() string             { return i.RebillFlag }
func (i InvoiceItem) GetPostedFlag() string             { return i.PostedFlag }
func (i InvoiceItem) GetStockedFlag() string            { return i.StockedFlag }
func (i InvoiceItem) GetWarrantyFlag() string           { return i.WarrantyFlag }
func (i InvoiceItem) GetSerializedWarrantyFlag() string { return i.SerializedWarrantyFlag }
func (i InvoiceItem) GetSerialNumber() string           { return i.SerialNumber }
func (i InvoiceItem) GetDivision1() string              { return i.Division1 }
func (i InvoiceItem) GetVendorCode1() string            { return i.VendorCode1 }
func (i InvoiceItem) GetPostingDate1() *time.Time       { return optionalTime(i.PostingDate1) }
func (i InvoiceItem) GetDivision2() string              { return i.Division2 }
func (i InvoiceItem) GetPostingDate2() *time.Time       { return optionalTime(i.PostingDate2) }
func (i InvoiceItem) GetDiscountIndicator() string      { return i.DiscountIndicator }
func (i InvoiceItem) GetExtendedDescription() string    { return i.ExtendedDescription }
func (i InvoiceItem) GetProcessCode() string            { return i.ProcessCode }
func (i InvoiceItem) GetDemandCode() string             { return i.DemandCode }
func (i InvoiceItem) GetSeasonalCode() string           { return i.SeasonalCode }
func (i InvoiceItem) GetFreightTaxCode() string         { return i.FreightTaxCode }
func (i InvoiceItem) GetOriginFlag() string             { return i.OriginFlag }
func (i InvoiceItem) GetUserField1() string             { return i.UserField1 }
func (i InvoiceItem) GetUserField2() string             { return i.UserField2 }

// ToInvoiceItemSpec converts the DB model into the exported API spec.
func (i InvoiceItem) ToInvoiceItemSpec() *types.InvoiceItemSpec {
	return &types.InvoiceItemSpec{
		DocumentNumber:       strings.TrimSpace(i.DocumentNumber),
		LineID:               strings.TrimSpace(i.LineID),
		Division:             strings.TrimSpace(i.Division),
		OrderSource:          strings.TrimSpace(i.OrderSource),
		CustomerNumber:       strings.TrimSpace(i.CustomerNumber),
		VendorCode:           strings.TrimSpace(i.VendorCode),
		PartNumber:           strings.TrimSpace(i.PartNumber),
		PostingDate:          i.PostingDate,
		FormatType:           strings.TrimSpace(i.FormatType),
		ExceptionCode:        strings.TrimSpace(i.ExceptionCode),
		Classification:       strings.TrimSpace(i.Classification),
		Description:          strings.TrimSpace(i.Description),
		PriceCode:            strings.TrimSpace(i.PriceCode),
		Cost:                 i.Cost,
		Location:             strings.TrimSpace(i.Location),
		Quantity:             i.Quantity,
		Price:                i.Price,
		TaxCode:              strings.TrimSpace(i.TaxCode),
		DiscountRate:         i.DiscountRate,
		Memo:                 optionalString(i.Memo),
		Rebill:               flagToBool(i.RebillFlag),
		Posted:               flagToBool(i.PostedFlag),
		Stocked:              flagToBool(i.StockedFlag),
		Warranty:             flagToBool(i.WarrantyFlag),
		SerializedWarranty:   flagToBool(i.SerializedWarrantyFlag),
		SerialNumber:         optionalString(i.SerialNumber),
		AuditDivision:        optionalString(i.Division1),
		AuditVendorCode:      optionalString(i.VendorCode1),
		AuditPostingDate:     optionalTime(i.PostingDate1),
		AlternateDivision:    optionalString(i.Division2),
		AlternatePostingDate: optionalTime(i.PostingDate2),
		DiscountIndicator:    optionalString(i.DiscountIndicator),
		ExtendedDescription:  optionalString(i.ExtendedDescription),
		ProcessCode:          optionalString(i.ProcessCode),
		DemandCode:           optionalString(i.DemandCode),
		SeasonalCode:         optionalString(i.SeasonalCode),
		FreightTaxCode:       optionalString(i.FreightTaxCode),
		OriginFlag:           optionalString(i.OriginFlag),
		UserField1:           optionalString(i.UserField1),
		UserField2:           optionalString(i.UserField2),
	}
}

func flagToBool(val string) bool {
	switch strings.TrimSpace(strings.ToUpper(val)) {
	case "Y", "YES", "1", "TRUE", "T":
		return true
	default:
		return false
	}
}

func optionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func optionalString(val string) *string {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
