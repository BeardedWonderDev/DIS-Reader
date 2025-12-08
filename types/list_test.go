package types

import (
	"errors"
	"strconv"
	"testing"
)

type demoParser struct{}

func (demoParser) GetDefaultSortBy() string { return "name" }
func (demoParser) GetAllowedSortByColumns() map[string]string {
	return map[string]string{"name": "name_col", "price": "price_col"}
}
func (demoParser) GetAllowedFilterColumns() map[string]string {
	return map[string]string{"status": "status_col"}
}
func (demoParser) ParseValue(val string, sortBy string) (interface{}, error) {
	if sortBy == "price_col" {
		return strconv.ParseFloat(val, 64)
	}
	return val, nil
}

func TestBuildListParamsSuccess(t *testing.T) {
	raw := RawListParams{
		PageParam:      "2",
		LimitParam:     "50",
		QueryParam:     "findme",
		SortByParam:    "price",
		SortOrderParam: "DESC",
	}
	lp, err := BuildListParams[demoParser](raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lp.Page != 2 || lp.Limit != 50 || lp.SortBy != "price_col" || lp.SortOrder != SortOrderDesc {
		t.Fatalf("unexpected list params: %+v", lp)
	}
}

func TestBuildListParamsUnsupportedSort(t *testing.T) {
	raw := RawListParams{SortByParam: "bad"}
	_, err := BuildListParams[demoParser](raw)
	var ip *InvalidParameterError
	if !errors.As(err, &ip) || ip.Parameter != paramNameSortBy {
		t.Fatalf("expected InvalidParameterError for sortBy, got %#v", err)
	}
}

func TestBuildListParamsCursorMissingAfterValue(t *testing.T) {
	raw := RawListParams{SortByParam: "price", AfterIdParam: "123"}
	_, err := BuildListParams[demoParser](raw)
	var mp *MissingRequiredParameterError
	if !errors.As(err, &mp) || mp.Parameter != paramNameAfterValue {
		t.Fatalf("expected MissingRequiredParameterError for afterValue, got %#v", err)
	}
}

func TestParseSortOrderInvalid(t *testing.T) {
	if _, err := ParseSortOrder("WRONG"); err == nil {
		t.Fatalf("expected error for invalid sort order")
	}
}
