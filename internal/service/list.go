package service

import (
	"fmt"
	"strconv"
)

const (
	paramNameLimit       = "limit"
	paramNamePage        = "page"
	paramNameQuery       = "q"
	paramNameSortBy      = "sortBy"
	paramNameSortOrder   = "sortOrder"
	paramNameAfterId     = "afterId"
	paramNameBeforeId    = "beforeId"
	paramNameAfterValue  = "afterValue"
	paramNameBeforeValue = "beforeValue"
	defaultLimit         = 200
	defaultPage          = 1

	SortOrderAsc  SortOrder = iota
	SortOrderDesc SortOrder = iota
)

type SortOrder int

func (so SortOrder) String() string {
	if so == SortOrderAsc {
		return "ASC"
	}

	if so == SortOrderDesc {
		return "DESC"
	}

	return ""
}

type ListParamParser interface {
	GetDefaultSortBy() string
	GetSupportedSortBys() []string
	ParseValue(val string, sortBy string) (interface{}, error)
}

// Filter represents a dynamic WHERE clause filter.
type Filter struct {
	Field    string      // lower-case logical field name
	Operator string      // SQL operator, e.g. "=", "LIKE", ">", "<", "IN", "NOT IN"
	Value    interface{} // the value to compare against
}

type ListParams struct {
	Page        int
	Limit       int
	Query       string
	SortBy      string
	SortOrder   SortOrder
	AfterId     string
	BeforeId    string
	AfterValue  interface{}
	BeforeValue interface{}
	// dynamic filters to apply in WHERE clauses
	Filters []Filter
}

func (lp ListParams) UseCursorPagination() bool {
	return lp.AfterId != "" || lp.BeforeId != "" || lp.AfterValue != nil || lp.BeforeValue != nil
}

type GetDefaultSortByFunc func() string
type GetSupportedSortBys func() []string

func ParsePage(val string) (int, error) {
	if val == "" {
		return defaultPage, nil
	}

	page, err := strconv.Atoi(val)
	if err != nil || page < 1 {
		return 0, fmt.Errorf("must be an integer greater than 0")
	}

	return page, nil
}

func ParseLimit(val string) (int, error) {
	if val == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(val)
	if err != nil || limit < 1 || limit > 10000 {
		return 0, fmt.Errorf("must be an integer greater than 0 and less than or equal to 10000")
	}

	return limit, nil
}

func ParseSortBy(val string, listParamParser ListParamParser) (string, error) {
	sortBy := val
	if sortBy == "" {
		sortBy = listParamParser.GetDefaultSortBy()
	}

	for _, supportedSortBy := range listParamParser.GetSupportedSortBys() {
		if sortBy == supportedSortBy {
			return sortBy, nil
		}
	}

	return "", fmt.Errorf("unsupported sortBy")
}

func ParseSortOrder(val string) (SortOrder, error) {
	switch val {
	case "ASC":
		return SortOrderAsc, nil
	case "DESC":
		return SortOrderDesc, nil
	case "":
		return SortOrderAsc, nil
	default:
		return SortOrderAsc, fmt.Errorf("must be ASC or DESC")
	}
}

func ParseId(val string) (string, error) {
	return val, nil
}

func ParseValue(val string, sortBy string, listParamParser ListParamParser) (interface{}, error) {
	value, err := listParamParser.ParseValue(val, sortBy)
	if err != nil {
		return nil, err
	}

	return value, nil
}
