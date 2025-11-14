package types

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// InvoiceService describes operations available for DIS invoice data.
type InvoiceService interface {
	GetItem(ctx context.Context, documentNumber string, lineID string) (*InvoiceItemSpec, error)
	ListItems(ctx context.Context, lp ListParams) ([]*InvoiceItemSpec, error)
}

// InvoiceItemSpec is the exported representation of a FILEC.IAH row.
type InvoiceItemSpec struct {
	DocumentNumber       string     `json:"documentNumber"`
	LineID               string     `json:"lineId"`
	Division             string     `json:"division"`
	OrderSource          string     `json:"orderSource"`
	CustomerNumber       string     `json:"customerNumber"`
	VendorCode           string     `json:"vendorCode"`
	PartNumber           string     `json:"partNumber"`
	PostingDate          time.Time  `json:"postingDate"`
	FormatType           string     `json:"formatType"`
	ExceptionCode        string     `json:"exceptionCode"`
	Classification       string     `json:"classification"`
	Description          string     `json:"description"`
	PriceCode            string     `json:"priceCode"`
	Cost                 float64    `json:"cost"`
	Location             string     `json:"location"`
	Quantity             float64    `json:"quantity"`
	Price                float64    `json:"price"`
	TaxCode              string     `json:"taxCode"`
	DiscountRate         float64    `json:"discountRate"`
	Memo                 *string    `json:"memo,omitempty"`
	Rebill               bool       `json:"rebill"`
	Posted               bool       `json:"posted"`
	Stocked              bool       `json:"stocked"`
	Warranty             bool       `json:"warranty"`
	SerializedWarranty   bool       `json:"serializedWarranty"`
	SerialNumber         *string    `json:"serialNumber,omitempty"`
	AuditDivision        *string    `json:"auditDivision,omitempty"`
	AuditVendorCode      *string    `json:"auditVendorCode,omitempty"`
	AuditPostingDate     *time.Time `json:"auditPostingDate,omitempty"`
	AlternateDivision    *string    `json:"alternateDivision,omitempty"`
	AlternatePostingDate *time.Time `json:"alternatePostingDate,omitempty"`
	DiscountIndicator    *string    `json:"discountIndicator,omitempty"`
	ExtendedDescription  *string    `json:"extendedDescription,omitempty"`
	ProcessCode          *string    `json:"processCode,omitempty"`
	DemandCode           *string    `json:"demandCode,omitempty"`
	SeasonalCode         *string    `json:"seasonalCode,omitempty"`
	FreightTaxCode       *string    `json:"freightTaxCode,omitempty"`
	OriginFlag           *string    `json:"originFlag,omitempty"`
	UserField1           *string    `json:"userField1,omitempty"`
	UserField2           *string    `json:"userField2,omitempty"`
}

// InvoiceListParamParser implements ListParamParser for invoices.
type InvoiceListParamParser struct{}

func (InvoiceListParamParser) GetDefaultSortBy() string {
	return "postingDate"
}

func (InvoiceListParamParser) GetAllowedSortByColumns() map[string]string {
	return map[string]string{
		"document":    "AHDOC",
		"line":        "AHLINE",
		"postingDate": "AHPDTE",
		"customer":    "AHAS#",
		"vendor":      "AHVEND",
		"part":        "AHPART",
		"division":    "AHDIV",
		"cost":        "AHCOST",
		"price":       "AHPRCE",
	}
}

func (InvoiceListParamParser) GetAllowedFilterColumns() map[string]string {
	return map[string]string{
		"document":    "AHDOC",
		"line":        "AHLINE",
		"postingDate": "AHPDTE",
		"customer":    "AHAS#",
		"vendor":      "AHVEND",
		"part":        "AHPART",
		"division":    "AHDIV",
		"orderSource": "AHAORS",
		"formatType":  "AHFRMT",
		"class":       "AHCLSS",
		"priceCode":   "AHPCDE",
	}
}

func (InvoiceListParamParser) ParseValue(val string, field string) (interface{}, error) {
	switch strings.ToLower(field) {
	case "postingdate":
		t, err := time.Parse("2006-01-02", val)
		if err != nil {
			return nil, fmt.Errorf("invalid date for %s: %w", field, err)
		}
		return t, nil
	case "cost", "price", "discountrate", "quantity":
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number for %s: %w", field, err)
		}
		return f, nil
	default:
		if val == "" {
			return nil, fmt.Errorf("%s must not be empty", field)
		}
		return val, nil
	}
}
