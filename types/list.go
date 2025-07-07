package types

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	slogctx "github.com/veqryn/slog-context"
)

const (
	paramNameLimit            = "limit"
	paramNamePage             = "page"
	paramNameQuery            = "q"
	paramNameSortBy           = "sortBy"
	paramNameSortOrder        = "sortOrder"
	paramNameAfterId          = "afterId"
	paramNameBeforeId         = "beforeId"
	paramNameAfterValue       = "afterValue"
	paramNameBeforeValue      = "beforeValue"
	paramNameFilters          = "filters"
	defaultLimit              = 200
	defaultPage               = 1
	contextKeyLimit       key = iota
	contextKeyPage        key = iota
	contextKeyQuery       key = iota
	contextKeySortBy      key = iota
	contextKeySortOrder   key = iota
	contextKeyAfterId     key = iota
	contextKeyBeforeId    key = iota
	contextKeyAfterValue  key = iota
	contextKeyBeforeValue key = iota
	contextKeyFilters     key = iota
)

const (
	SortOrderAsc  SortOrder = iota
	SortOrderDesc SortOrder = iota
)

type key int
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
	GetAllowedSortByColumns() map[string]string
	GetAllowedFilterColumns() map[string]string
	ParseValue(val string, sortBy string) (interface{}, error)
}

// Filter represents a dynamic WHERE clause filter.
type Filter struct {
	Field    string      // lower-case logical field name
	Operator string      // SQL operator, e.g. "=", "LIKE", ">", "<", "IN", "NOT IN"
	Value    interface{} // the value to compare against
	Column   string      // the SQL column name - to be added by parser
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

// RawListParams holds raw string inputs for building ListParams in a type-safe manner.
type RawListParams struct {
	PageParam        string
	LimitParam       string
	QueryParam       string
	SortByParam      string
	SortOrderParam   string
	AfterIdParam     string
	BeforeIdParam    string
	AfterValueParam  string
	BeforeValueParam string
	// Pre-parsed filters for programmatic usage
	FilterObjects []Filter
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

	supportedSortBy := listParamParser.GetAllowedSortByColumns()
	col, ok := supportedSortBy[sortBy]
	if !ok {
		return "", fmt.Errorf("unsupported sortBy")
	}

	return col, nil
}

func ParseFilter(filter Filter, listParamParser ListParamParser) (Filter, error) {
	// Verify Field Is Allowed And Get Column
	supportedFilter := listParamParser.GetAllowedFilterColumns()
	col, ok := supportedFilter[filter.Field]
	if !ok {
		return Filter{}, fmt.Errorf("unsupported filter field")
	}

	filter.Column = col

	// Handle slice for IN/NOT IN
	switch v := filter.Value.(type) {
	case []string:
		placeholders := make([]string, len(v))
		for i, s := range v {
			placeholders[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
		}
		joined := strings.Join(placeholders, ",")
		if strings.EqualFold(filter.Operator, "IN") || strings.EqualFold(filter.Operator, "NOT IN") {
			filter.Value = fmt.Sprintf("(%s)", joined)
		} else {
			filter.Value = joined
		}
	default:
		filter.Value = v
	}

	return filter, nil
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

// BuildListParams parses raw string inputs into ListParams.
// T must implement ListParamParser.
func BuildListParams[T ListParamParser](raw RawListParams) (ListParams, error) {
	// Instantiate parser
	parser := ListParamParser(*new(T))

	// Parse filters from raw FilterObjects
	var filters []Filter
	for _, rawFilter := range raw.FilterObjects {
		parsed, err := ParseFilter(rawFilter, parser)
		if err != nil {
			return ListParams{}, NewInvalidParameterError(paramNameFilters, err.Error())
		}
		filters = append(filters, parsed)
	}

	// Extract raw values
	pageParam := raw.PageParam
	limitParam := raw.LimitParam
	query := raw.QueryParam
	sortByParam := raw.SortByParam
	sortOrderParam := raw.SortOrderParam
	afterIdParam := raw.AfterIdParam
	beforeIdParam := raw.BeforeIdParam
	afterValueParam := raw.AfterValueParam
	beforeValueParam := raw.BeforeValueParam

	// Parse page
	page, err := ParsePage(pageParam)
	if err != nil {
		return ListParams{}, NewInvalidParameterError(paramNamePage, err.Error())
	}
	// Parse limit
	limit, err := ParseLimit(limitParam)
	if err != nil {
		return ListParams{}, NewInvalidParameterError(paramNameLimit, err.Error())
	}
	// Parse sort order
	sortOrder, err := ParseSortOrder(sortOrderParam)
	if err != nil {
		return ListParams{}, NewInvalidParameterError(paramNameSortOrder, err.Error())
	}
	// Parse sortBy
	sortBy, err := ParseSortBy(sortByParam, parser)
	if err != nil {
		return ListParams{}, NewInvalidParameterError(paramNameSortBy, err.Error())
	}
	// Parse ids
	afterId, err := ParseId(afterIdParam)
	if err != nil {
		return ListParams{}, NewInvalidParameterError(paramNameAfterId, err.Error())
	}
	beforeId, err := ParseId(beforeIdParam)
	if err != nil {
		return ListParams{}, NewInvalidParameterError(paramNameBeforeId, err.Error())
	}
	// Value checks for cursor pagination prerequisites...
	if raw.AfterValueParam != "" && afterId == "" {
		return ListParams{}, NewMissingRequiredParameterError(paramNameAfterId)
	}
	if raw.BeforeValueParam != "" && beforeId == "" {
		return ListParams{}, NewMissingRequiredParameterError(paramNameBeforeId)
	}
	defaultSortBy := parser.GetDefaultSortBy()
	if raw.AfterValueParam == "" && sortBy != defaultSortBy && afterId != "" {
		return ListParams{}, NewMissingRequiredParameterError(paramNameAfterValue)
	}
	if raw.BeforeValueParam == "" && sortBy != defaultSortBy && beforeId != "" {
		return ListParams{}, NewMissingRequiredParameterError(paramNameBeforeValue)
	}
	if (raw.AfterValueParam != "" || raw.BeforeValueParam != "") && sortBy == defaultSortBy {
		return ListParams{}, NewInvalidRequestError(fmt.Sprintf("cannot pass %s or %s when sorting by %s", paramNameAfterValue, paramNameBeforeValue, defaultSortBy))
	}
	// Parse afterValue and beforeValue
	var afterValue interface{} = nil
	if raw.AfterValueParam != "" {
		afterValue, err = ParseValue(afterValueParam, sortBy, parser)
		if err != nil {
			return ListParams{}, NewInvalidParameterError(paramNameAfterValue, err.Error())
		}
	}
	var beforeValue interface{} = nil
	if raw.BeforeValueParam != "" {
		beforeValue, err = ParseValue(beforeValueParam, sortBy, parser)
		if err != nil {
			return ListParams{}, NewInvalidParameterError(paramNameBeforeValue, err.Error())
		}
	}
	// Return assembled params
	return ListParams{
		Page:        page,
		Limit:       limit,
		Query:       query,
		SortBy:      sortBy,
		SortOrder:   sortOrder,
		AfterId:     afterId,
		BeforeId:    beforeId,
		AfterValue:  afterValue,
		BeforeValue: beforeValue,
		Filters:     filters,
	}, nil
}

func ListMiddleware[T ListParamParser](next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawFilterStrs := r.URL.Query()[paramNameFilters]
		var rawFilters []Filter
		for _, fstr := range rawFilterStrs {
			parts := strings.SplitN(fstr, ":", 3)
			if len(parts) != 3 {
				SendErrorResponse(w, NewInvalidParameterError(paramNameFilters, "invalid filter format, expected field:operator:value"))
				return
			}
			rawFilters = append(rawFilters, Filter{
				Field:    parts[0],
				Operator: parts[1],
				Value:    parts[2],
			})
		}

		queryParams := r.URL.Query()
		raw := RawListParams{
			PageParam:        queryParams.Get(paramNamePage),
			LimitParam:       queryParams.Get(paramNameLimit),
			QueryParam:       queryParams.Get(paramNameQuery),
			SortByParam:      queryParams.Get(paramNameSortBy),
			SortOrderParam:   queryParams.Get(paramNameSortOrder),
			AfterIdParam:     queryParams.Get(paramNameAfterId),
			BeforeIdParam:    queryParams.Get(paramNameBeforeId),
			AfterValueParam:  queryParams.Get(paramNameAfterValue),
			BeforeValueParam: queryParams.Get(paramNameBeforeValue),
			FilterObjects:    rawFilters,
		}
		lp, err := BuildListParams[T](raw)
		if err != nil {
			SendErrorResponse(w, err)
			return
		}
		// Add each field to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, contextKeyPage, lp.Page)
		ctx = context.WithValue(ctx, contextKeyLimit, lp.Limit)
		ctx = context.WithValue(ctx, contextKeyQuery, lp.Query)
		ctx = context.WithValue(ctx, contextKeySortBy, lp.SortBy)
		ctx = context.WithValue(ctx, contextKeySortOrder, lp.SortOrder)
		ctx = context.WithValue(ctx, contextKeyAfterId, lp.AfterId)
		ctx = context.WithValue(ctx, contextKeyBeforeId, lp.BeforeId)
		ctx = context.WithValue(ctx, contextKeyAfterValue, lp.AfterValue)
		ctx = context.WithValue(ctx, contextKeyBeforeValue, lp.BeforeValue)
		ctx = context.WithValue(ctx, contextKeyFilters, lp.Filters)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetListParamsFromContext(context context.Context) ListParams {
	contextPage := context.Value(contextKeyPage)
	if contextPage == nil {
		slogctx.FromCtx(context).Error("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextLimit := context.Value(contextKeyLimit)
	if contextLimit == nil {
		slogctx.FromCtx(context).Error("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextSortBy := context.Value(contextKeySortBy)
	if contextSortBy == nil {
		slogctx.FromCtx(context).Error("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextSortOrder := context.Value(contextKeySortOrder)
	if contextSortOrder == nil {
		slogctx.FromCtx(context).Error("List context not available. Did you forget to add ListMiddleware to a handler?")
	}

	contextQuery := context.Value(contextKeyQuery)
	contextAfterId := context.Value(contextKeyAfterId)
	contextBeforeId := context.Value(contextKeyBeforeId)
	contextAfterValue := context.Value(contextKeyAfterValue)
	contextBeforeValue := context.Value(contextKeyBeforeValue)

	contextFilters := context.Value(contextKeyFilters)
	if contextFilters == nil {
		slogctx.FromCtx(context).Error("List context filters not available. Did you forget to add ListMiddleware filters?")
	}

	return ListParams{
		Page:        contextPage.(int),
		Limit:       contextLimit.(int),
		Query:       contextQuery.(string),
		SortBy:      contextSortBy.(string),
		SortOrder:   contextSortOrder.(SortOrder),
		AfterId:     contextAfterId.(string),
		BeforeId:    contextBeforeId.(string),
		AfterValue:  contextAfterValue,
		BeforeValue: contextBeforeValue,
		Filters:     contextFilters.([]Filter),
	}
}
