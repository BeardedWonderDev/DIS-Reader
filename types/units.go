package types

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type UnitService interface {
	GetByUnitNumber(ctx context.Context, unitNum string) (*UnitSpec, error)
	ListUnits(ctx context.Context, lp ListParams) ([]*UnitSpec, error)
}

type UnitSpec struct {
	UnitID        string    `json:"unitId" validate:"required"`
	Year          string    `json:"year"`
	Make          string    `json:"make"`
	Model         string    `json:"model"`
	Description   string    `json:"description"`
	Serial        *string   `json:"serial,omitempty"`
	Engine        *string   `json:"engine,omitempty"`
	Status        string    `json:"status"`
	ProductCode   *string   `json:"productCode,omitempty"`
	New           bool      `json:"new"`
	Color         *string   `json:"color,omitempty"`
	HorsePower    *int      `json:"horsePower,omitempty"`
	EngineHours   *float64  `json:"engineHours,omitempty"`
	Location      string    `json:"location"`
	Account       string    `json:"account"`
	CreatedAt     time.Time `json:"createdAt"`
	SoldAt        time.Time `json:"soldAt,omitempty"`
	InvoiceNumber *string   `json:"invoiceNumber,omitempty"`
	SoldTo        *string   `json:"soldTo,omitempty"`
	RevenueAmount *float64  `json:"revenueAmount,omitempty"`
	Cost          float64   `json:"cost,omitempty"`
	SoldBy        *string   `json:"soldBy,omitempty"`
	SoldByName    *string   `json:"soldByName,omitempty"`
}

// UnitListParamParser satisfies service.ListParamParser for units.
type UnitListParamParser struct{}

func (UnitListParamParser) GetDefaultSortBy() string {
	return "unit"
}

func (UnitListParamParser) GetAllowedSortByColumns() map[string]string {
	return map[string]string{
		"unit":     "UNIT",
		"year":     "YEAR",
		"make":     "MAKE",
		"model":    "MODEL",
		"location": "LOCATION",
		"indate":   "INDATE",
		"soldat":   "SOLDAT",
		"soldto":   "SOLDTO",
		"amount":   "AMOUNT",
		"cost":     "COST",
		"soldby":   "SOLDBY",
	}
}

func (UnitListParamParser) GetAllowedFilterColumns() map[string]string {
	return map[string]string{
		"unit":        "UNIT",
		"year":        "YEAR",
		"make":        "MAKE",
		"model":       "MODEL",
		"productCode": "PRODCT",
		"location":    "LOCATION",
		"indate":      "INDATE",
		"soldat":      "SOLDAT",
		"soldto":      "SOLDTO",
		"amount":      "AMOUNT",
		"cost":        "COST",
		"soldby":      "SOLDBY",
	}
}

func (UnitListParamParser) ParseValue(val string, sortBy string) (interface{}, error) {
	switch sortBy {
	case "indate", "soldat":
		// Expect dates as YYYY-MM-DD
		t, err := time.Parse("2006-01-02", val)
		if err != nil {
			return nil, fmt.Errorf("invalid date for %s: %w", sortBy, err)
		}
		return t, nil
	case "amount", "cost":
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number for %s: %w", sortBy, err)
		}
		return f, nil
	case "year":
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("invalid integer for year: %w", err)
		}
		return i, nil
	default:
		if val == "" {
			return nil, fmt.Errorf("%s must not be empty", sortBy)
		}
		return val, nil
	}
}
