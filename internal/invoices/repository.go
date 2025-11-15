package invoices

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Repository handles raw SQL access for invoice items.
type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

// GetByDocumentAndLine returns a single invoice line given its document and line identifiers.
func (repo *Repository) GetByDocumentAndLine(ctx context.Context, documentNumber string, lineID string) (Model, error) {
	query := fmt.Sprintf(
		"SELECT * FROM FILEC.IAH WHERE AHDOC = '%s' AND AHLINE = '%s'",
		escapeLiteral(documentNumber),
		escapeLiteral(lineID),
	)
	var item InvoiceItem
	if err := repo.db.Get(ctx, &item, query); err != nil {
		return nil, err
	}
	return &item, nil
}

// ListInvoiceItems returns invoice rows filtered, searched, and sorted according to ListParams.
func (repo *Repository) ListInvoiceItems(ctx context.Context, lp types.ListParams) ([]InvoiceItem, error) {
	base := "SELECT * FROM FILEC.IAH WHERE 1=1"

	if lp.Query != "" {
		search := "%" + escapeLike(lp.Query) + "%"
		base += fmt.Sprintf(
			" AND (AHDOC LIKE '%s' ESCAPE '\\' OR AHPART LIKE '%s' ESCAPE '\\' OR AHAS# LIKE '%s' ESCAPE '\\')",
			search, search, search,
		)
	}

	for _, f := range lp.Filters {
		if strings.EqualFold(f.Field, "postingdate") {
			expr := database.As400DateExpr("AHPDTE")
			t, ok := f.Value.(time.Time)
			if !ok {
				return nil, fmt.Errorf("expected time.Time for postingDate filter")
			}
			base += fmt.Sprintf(" AND %s %s DATE('%s')", expr, f.Operator, t.Format("2006-01-02"))
			continue
		}

		val := formatFilterValue(f.Value)
		base += fmt.Sprintf(" AND %s %s %s", f.Column, f.Operator, val)
	}

	sortBy := lp.SortBy
	if strings.EqualFold(sortBy, "AHPDTE") {
		sortBy = database.As400DateExpr("AHPDTE")
	}
	if sortBy == "" {
		sortBy = "AHPDTE"
	}
	base += fmt.Sprintf(" ORDER BY %s %s", sortBy, lp.SortOrder.String())

	if lp.Limit > 0 {
		base += fmt.Sprintf(" LIMIT %d", lp.Limit)
	}

	var items []InvoiceItem
	if err := repo.db.Select(ctx, &items, base); err != nil {
		return nil, err
	}
	return items, nil
}

func escapeLiteral(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

func escapeLike(val string) string {
	val = strings.ReplaceAll(val, `\`, `\\`)
	val = strings.ReplaceAll(val, "%", `\%`)
	val = strings.ReplaceAll(val, "_", `\_`)
	val = strings.ReplaceAll(val, "'", "''")
	return val
}

func formatFilterValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "(") && strings.HasSuffix(v, ")") {
			return v
		}
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case time.Time:
		return fmt.Sprintf("DATE('%s')", v.Format("2006-01-02"))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetWholeGoodsInvoice returns a single whole goods invoice row.
func (repo *Repository) GetWholeGoodsInvoice(ctx context.Context, invoiceNumber string, lineItemNumber string) (*WholeGoodsInvoice, error) {
	inv := strings.TrimSpace(invoiceNumber)
	if inv == "" {
		return nil, fmt.Errorf("invoice number is required")
	}
	lineLiteral, err := numericLiteral(lineItemNumber)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT * FROM FILEC.CUSINV WHERE DSHUA = '%s' AND DS1XQA = %s",
		escapeLiteral(inv),
		lineLiteral,
	)

	var invoice WholeGoodsInvoice
	if err := repo.db.Get(ctx, &invoice, query); err != nil {
		return nil, err
	}
	return &invoice, nil
}

// ListWholeGoodsInvoices fetches whole goods invoices with filtering.
func (repo *Repository) ListWholeGoodsInvoices(ctx context.Context, lp types.ListParams) ([]WholeGoodsInvoice, error) {
	base := "SELECT * FROM FILEC.CUSINV WHERE 1=1"

	if lp.Query != "" {
		search := "%" + escapeLike(lp.Query) + "%"
		base += fmt.Sprintf(
			" AND (DSHUA LIKE '%s' ESCAPE '\\' OR DSHXA LIKE '%s' ESCAPE '\\' OR DS22UA LIKE '%s' ESCAPE '\\')",
			search, search, search,
		)
	}

	for _, f := range lp.Filters {
		if strings.EqualFold(f.Field, "invoicedate") {
			val, ok := f.Value.(time.Time)
			if !ok {
				return nil, fmt.Errorf("expected time.Time for invoiceDate filter")
			}
			base += fmt.Sprintf(" AND DATE(DSHVA) %s DATE('%s')", f.Operator, val.Format("2006-01-02"))
			continue
		}
		val := formatFilterValue(f.Value)
		base += fmt.Sprintf(" AND %s %s %s", f.Column, f.Operator, val)
	}

	sortBy := lp.SortBy
	if sortBy == "" {
		sortBy = "DSHVA"
	}
	base += fmt.Sprintf(" ORDER BY %s %s", sortBy, lp.SortOrder.String())

	if lp.Limit > 0 {
		base += fmt.Sprintf(" LIMIT %d", lp.Limit)
	}

	var invoices []WholeGoodsInvoice
	if err := repo.db.Select(ctx, &invoices, base); err != nil {
		return nil, err
	}
	return invoices, nil
}

func numericLiteral(val string) (string, error) {
	trimmed := strings.TrimSpace(val)
	if trimmed == "" {
		return "", fmt.Errorf("line item is required")
	}
	if _, err := strconv.Atoi(trimmed); err != nil {
		return "", fmt.Errorf("line item must be numeric: %w", err)
	}
	return trimmed, nil
}
