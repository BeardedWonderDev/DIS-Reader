package types

import (
	"time"
)

// Part represents a parts/inventory record with parsed date fields.
type Part struct {
	SrcTable     string    `json:"SRC_TABLE"`
	Division     string    `json:"PIMDIV"`
	Vendor       string    `json:"PIMVEN"`
	PartNumber   string    `json:"PIMPRT"`
	Description  string    `json:"PIMDES"`
	Bin          string    `json:"PIMBIN"`
	Class        string    `json:"PIMCLS"`
	Active       string    `json:"PIMACT"`
	Taxable      string    `json:"PIMTAX"`
	DateAdded    time.Time `json:"PIMDAD"`
	DateLastChg  time.Time `json:"PIMDLC"`
	DateLastSale time.Time `json:"PIMDLS"`
	Weight       float64   `json:"PIMWHT"`
	PkgQty       int       `json:"PIMPKG"`
	QtyRequired  int       `json:"PIMJOB"`
	OnHand       int       `json:"PIMONH"`
	ReservedRO   int       `json:"PIMRSA"`
	ReservedCT   int       `json:"PIMRWO"`
	Available    int       `json:"PIMAVL"`
	LastPurchase time.Time `json:"PIMPUP"`
	AvgCost      float64   `json:"PIMAVG"`
	DealerList   float64   `json:"PIMCBS"`
	DlrNetPrice  float64   `json:"PIMCVA"`
	IntPrice     float64   `json:"PIMIVA"`
	SuggList     float64   `json:"PIMSVA"`
	ListPrice    float64   `json:"PIMLVA"`
	DeliveryCode string    `json:"PIMDLV"`
	OrderCode    string    `json:"PIMORC"`
	MinQty       int       `json:"PIMMIN"`
	MaxQty       int       `json:"PIMMAX"`
	OrderQty     int       `json:"PIMODQ"`
	LastRcvDate  time.Time `json:"PIMLRD"`
	LastRcvQty   int       `json:"PIMLRQ"`
	DealerOrder  string    `json:"PIMPSO"`
	APRCode      string    `json:"PIMAPR"`
	APRSource    string    `json:"PIMAPC"`
	Bulk         int       `json:"PIM9BL"`
	Comments     string    `json:"PIMCMT"`
}

type PartsFilter struct {
	PartNumber string
	Vendor     string
	Division   string
	// Add other filter options as needed
}
