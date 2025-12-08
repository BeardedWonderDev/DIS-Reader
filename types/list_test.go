package types

import (
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestListMiddlewarePopulatesContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?limit=5&page=2&sortBy=name&sortOrder=ASC&filters=status:eq:active", nil)
	rw := httptest.NewRecorder()
	called := false
	h := ListMiddleware[demoParser](http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		lp := GetListParamsFromContext(r.Context())
		if lp.Page != 2 || lp.Limit != 5 || lp.SortBy != "name_col" || lp.SortOrder != SortOrderAsc {
			t.Fatalf("unexpected list params: %+v", lp)
		}
		if len(lp.Filters) != 1 || lp.Filters[0].Column != "status_col" || lp.Filters[0].Value != "active" {
			t.Fatalf("unexpected filters: %+v", lp.Filters)
		}
	}))

	h.ServeHTTP(rw, req)
	if !called {
		t.Fatalf("handler was not called")
	}
}

func TestListMiddlewareRejectsBadFilterFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?filters=badfilter", nil)
	rw := httptest.NewRecorder()
	h := ListMiddleware[demoParser](http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("handler should not be reached on error")
	}))
	h.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad filter format, got %d", rw.Code)
	}
}

func TestParseFilterHandlesInAndEscaping(t *testing.T) {
	filter := Filter{Field: "status", Operator: "IN", Value: []string{"a", "b'b"}}
	parsed, err := ParseFilter(filter, demoParser{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Column != "status_col" {
		t.Fatalf("expected column to be mapped, got %s", parsed.Column)
	}
	expected := "('a','b''b')"
	if parsed.Value != expected {
		t.Fatalf("expected escaped IN list %s, got %v", expected, parsed.Value)
	}
}

func TestParseFilterRejectsUnknownField(t *testing.T) {
	filter := Filter{Field: "unknown", Operator: "=", Value: "x"}
	if _, err := ParseFilter(filter, demoParser{}); err == nil {
		t.Fatalf("expected error for unsupported filter field")
	}
}

func TestParseValuePropagatesParserError(t *testing.T) {
	if _, err := ParseValue("oops", "price_col", demoParser{}); err == nil {
		t.Fatalf("expected parser error for invalid float")
	}
}

func TestBuildListParamsWithFiltersAndCursorValue(t *testing.T) {
	raw := RawListParams{
		SortByParam:     "price",
		SortOrderParam:  "ASC",
		AfterIdParam:    "cursor-1",
		AfterValueParam: "12.5",
		FilterObjects: []Filter{{
			Field:    "status",
			Operator: "=",
			Value:    "active",
		}},
	}
	lp, err := BuildListParams[demoParser](raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lp.AfterId != "cursor-1" {
		t.Fatalf("expected afterId to be preserved, got %s", lp.AfterId)
	}
	if fv, ok := lp.AfterValue.(float64); !ok || fv != 12.5 {
		t.Fatalf("expected parsed afterValue 12.5 float, got %#v", lp.AfterValue)
	}
	if len(lp.Filters) != 1 || lp.Filters[0].Column != "status_col" || lp.Filters[0].Value != "active" {
		t.Fatalf("expected mapped filters, got %+v", lp.Filters)
	}
}
