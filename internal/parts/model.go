package parts

import (
	"time"
)

// InvoiceItemModel defines the Model interface methods for InvoiceItem
type InvoiceItemModel interface {
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
	GetGrossTaxRate() float64
	GetRebillFlag() string
	GetPostedFlag() string
	GetStockedFlag() string
	GetWarrantyFlag() string
	GetSerializedWarrantyFlag() string
	GetSerialNumber() string
	GetDivision1() string
	GetVendorCode1() string
	GetPostingDate1() time.Time
	GetDivision2() string
	GetPostingDate2() time.Time
}

// InvoiceItem represents a single line item from FILEC.IAH
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
	GrossTaxRate           float64   `mapstructure:"AHGSTR"`
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
func (i InvoiceItem) GetGrossTaxRate() float64          { return i.GrossTaxRate }
func (i InvoiceItem) GetRebillFlag() string             { return i.RebillFlag }
func (i InvoiceItem) GetPostedFlag() string             { return i.PostedFlag }
func (i InvoiceItem) GetStockedFlag() string            { return i.StockedFlag }
func (i InvoiceItem) GetWarrantyFlag() string           { return i.WarrantyFlag }
func (i InvoiceItem) GetSerializedWarrantyFlag() string { return i.SerializedWarrantyFlag }
func (i InvoiceItem) GetSerialNumber() string           { return i.SerialNumber }
func (i InvoiceItem) GetDivision1() string              { return i.Division1 }
func (i InvoiceItem) GetVendorCode1() string            { return i.VendorCode1 }
func (i InvoiceItem) GetPostingDate1() time.Time        { return i.PostingDate1 }
func (i InvoiceItem) GetDivision2() string              { return i.Division2 }
func (i InvoiceItem) GetPostingDate2() time.Time        { return i.PostingDate2 }
