package unit

import (
	"fmt"
	"strconv"
	"time"
)

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
