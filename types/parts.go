package types

import (
	"context"
	"fmt"
	"strconv"
)

// PartService exposes accessors for DIS parts/inventory records.
type PartService interface {
	GetPart(ctx context.Context, division, partNumber string) (*PartInventorySpec, error)
	ListParts(ctx context.Context, lp ListParams) ([]*PartInventorySpec, error)
}

// PartInventorySpec represents a fully decoded parts master record.
type PartInventorySpec struct {
	Division            string  `json:"division"`
	VendorCode          string  `json:"vendorCode"`
	PartNumber          string  `json:"partNumber"`
	Description         string  `json:"description"`
	BinLocation         string  `json:"binLocation"`
	QuickCode           string  `json:"quickCode"`
	ItemClass           string  `json:"itemClass"`
	ActiveFlag          string  `json:"activeFlag"`
	TaxCode             string  `json:"taxCode"`
	DateAddedYYYYMM     int32   `json:"dateAddedYYYYMM"`
	DateLastCountYYYYMM int32   `json:"dateLastCountYYYYMM"`
	DateLastSaleYYYYMM  int32   `json:"dateLastSaleYYYYMM"`
	UnitWeight          float64 `json:"unitWeight"`
	BackorderFlag       string  `json:"backorderFlag"`
	PackQty             int32   `json:"packQty"`
	OrderMultiplier     int32   `json:"orderMultiplier"`
	JobCode             int32   `json:"jobCode"`
	OnHandQty           int32   `json:"onHandQty"`
	ReservedSA          int32   `json:"reservedSA"`
	ReservedWO          int32   `json:"reservedWO"`
	AvailableQty        int32   `json:"availableQty"`
	ReturnableFlag      string  `json:"returnableFlag"`
	DateLastPriceUpdate int32   `json:"dateLastPriceUpdate"`
	AvgCost             float64 `json:"avgCost"`
	Price1Basis         string  `json:"price1Basis"`
	Price1Value         float64 `json:"price1Value"`
	Price2Basis         string  `json:"price2Basis"`
	Price2Value         float64 `json:"price2Value"`
	Price3Basis         string  `json:"price3Basis"`
	Price3Value         float64 `json:"price3Value"`
	Price4Basis         string  `json:"price4Basis"`
	Price4Value         float64 `json:"price4Value"`
	DeliverCode         string  `json:"deliverCode"`
	OrderCode           string  `json:"orderCode"`
	MinQty              int32   `json:"minQty"`
	MaxQty              int32   `json:"maxQty"`
	OrderQty            int32   `json:"orderQty"`
	LastReceiptDate     int32   `json:"lastReceiptDate"`
	LastReceiptQty      int32   `json:"lastReceiptQty"`
	StockFlag           string  `json:"stockFlag"`
	OpenOnSales         int32   `json:"openOnSales"`
	OpenOnPO            int32   `json:"openOnPO"`
	OpenOnRO            int32   `json:"openOnRO"`
	OpenOnReq           int32   `json:"openOnReq"`
	Peak12MoUsage       int32   `json:"peak12MoUsage"`
	MonthlyUsage        []int32 `json:"monthlyUsage"`
	MonthlyPurchases    []int32 `json:"monthlyPurchases"`
	YearUsageQty        []int32 `json:"yearUsageQty"`
	YearUsageValue      []int64 `json:"yearUsageValue"`
	Comment             string  `json:"comment"`
	CycleCountCode      string  `json:"cycleCountCode"`
	CorePartNumber      string  `json:"corePartNumber"`
	CoreQty             int32   `json:"coreQty"`
	CorePrice           float64 `json:"corePrice"`
	Sup1Vendor          string  `json:"sup1Vendor"`
	Sup1Part            string  `json:"sup1Part"`
	Sup1Price           float64 `json:"sup1Price"`
	Sup2Vendor          string  `json:"sup2Vendor"`
	Sup2Part            string  `json:"sup2Part"`
	Sup2Price           float64 `json:"sup2Price"`
	Sup3Vendor          string  `json:"sup3Vendor"`
	Sup3Part            string  `json:"sup3Part"`
	Sup3Price           float64 `json:"sup3Price"`
	RebuildIndicator    string  `json:"rebuildIndicator"`
	RebuildVendor       string  `json:"rebuildVendor"`
	RebuildPart         string  `json:"rebuildPart"`
	RebuildQtyFactor    int32   `json:"rebuildQtyFactor"`
	PriceScheduleCode   string  `json:"priceScheduleCode"`
	AutoPricingRule     string  `json:"autoPricingRule"`
	AutoPricingClass    string  `json:"autoPricingClass"`
	AutoPricingSource   string  `json:"autoPricingSource"`
	AutoPricingQtyBrk   int32   `json:"autoPricingQtyBrk"`
	AltAvailQty         float64 `json:"altAvailQty"`
	AltOnHandQty        float64 `json:"altOnHandQty"`
	AltUomUnits         int32   `json:"altUomUnits"`
	AltUomCode          string  `json:"altUomCode"`
	AltUomMultiple      int32   `json:"altUomMultiple"`
	AltUomBackorderFlag string  `json:"altUomBackorderFlag"`
	AltUomFreightFlag   string  `json:"altUomFreightFlag"`
	Message1            int32   `json:"message1"`
	Message2            int32   `json:"message2"`
	Message3            int32   `json:"message3"`
}

// PartListParamParser validates filters, sorts, and cursor fields for parts endpoints.
type PartListParamParser struct{}

func (PartListParamParser) GetDefaultSortBy() string {
	return "partNumber"
}

func (PartListParamParser) GetAllowedSortByColumns() map[string]string {
	return map[string]string{
		"partNumber":   "PIMPRT",
		"division":     "PIMDIV",
		"vendor":       "PIMVEN",
		"description":  "PIMDES",
		"available":    "PIMAVL",
		"onHand":       "PIMONH",
		"avgCost":      "PIMAVG",
		"price":        "PIMCVA",
		"dateAdded":    "PIMDAD",
		"dateLastSale": "PIMDLS",
	}
}

func (PartListParamParser) GetAllowedFilterColumns() map[string]string {
	return map[string]string{
		"division":    "PIMDIV",
		"vendor":      "PIMVEN",
		"partNumber":  "PIMPRT",
		"quickCode":   "PIMQKC",
		"itemClass":   "PIMCLS",
		"active":      "PIMACT",
		"bin":         "PIMBIN",
		"taxCode":     "PIMTAX",
		"orderCode":   "PIMORC",
		"stockFlag":   "PIMSTK",
		"deliverCode": "PIMDLV",
	}
}

func (PartListParamParser) ParseValue(val string, column string) (interface{}, error) {
	if val == "" {
		return nil, fmt.Errorf("%s must not be empty", column)
	}

	switch column {
	case "PIMDAD", "PIMDLC", "PIMDLS", "PIMLRD", "PIMPUP":
		num, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("%s must be numeric", column)
		}
		return num, nil
	case "PIMONH", "PIMAVL", "PIMMIN", "PIMMAX", "PIMODQ", "PIMLRQ":
		num, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("%s must be numeric", column)
		}
		return num, nil
	case "PIMAVG", "PIMCVA", "PIMIVA", "PIMSVA", "PIMLVA":
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be a decimal", column)
		}
		return f, nil
	default:
		return val, nil
	}
}
