package parts

import (
	"context"
	"fmt"
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

const partsTable = "FILEC.DMPIMI"

// Repository encapsulates SQL used to fetch DIS part inventory rows.
type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPart(ctx context.Context, division, partNumber string) (*PartInventory, error) {
	if strings.TrimSpace(division) == "" {
		return nil, fmt.Errorf("division is required")
	}
	if strings.TrimSpace(partNumber) == "" {
		return nil, fmt.Errorf("part number is required")
	}

	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE PIMDIV = %s AND PIMPRT = %s",
		partsTable,
		quoteLiteral(division),
		quoteLiteral(partNumber),
	)

	var row PartInventory
	if err := r.db.Get(ctx, &row, query); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repository) ListParts(ctx context.Context, lp types.ListParams) ([]PartInventory, error) {
	var builder strings.Builder
	builder.WriteString("SELECT * FROM ")
	builder.WriteString(partsTable)
	builder.WriteString(" WHERE 1=1")

	if q := strings.TrimSpace(lp.Query); q != "" {
		pattern := quoteLiteral("%" + escapeLike(q) + "%")
		builder.WriteString(" AND (")
		builder.WriteString("PIMPRT LIKE ")
		builder.WriteString(pattern)
		builder.WriteString(" OR PIMDES LIKE ")
		builder.WriteString(pattern)
		builder.WriteString(" OR PIMVEN LIKE ")
		builder.WriteString(pattern)
		builder.WriteString(" OR PIMQKC LIKE ")
		builder.WriteString(pattern)
		builder.WriteString(")")
	}

	for _, f := range lp.Filters {
		builder.WriteString(" AND ")
		builder.WriteString(f.Column)
		builder.WriteRune(' ')
		builder.WriteString(f.Operator)
		builder.WriteRune(' ')
		builder.WriteString(formatFilterValue(f.Value))
	}

	sortBy := lp.SortBy
	if sortBy == "" {
		sortBy = "PIMPRT"
	}
	sortOrder := lp.SortOrder.String()
	if sortOrder == "" {
		sortOrder = types.SortOrderAsc.String()
	}
	builder.WriteString(" ORDER BY ")
	builder.WriteString(sortBy)
	builder.WriteRune(' ')
	builder.WriteString(sortOrder)

	if lp.Limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", lp.Limit))
	}

	var rows []PartInventory
	if err := r.db.Select(ctx, &rows, builder.String()); err != nil {
		return nil, err
	}
	return rows, nil
}

func quoteLiteral(val string) string {
	return "'" + strings.ReplaceAll(val, "'", "''") + "'"
}

func escapeLike(val string) string {
	return strings.ReplaceAll(val, "'", "''")
}

func formatFilterValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
			return trimmed
		}
		return quoteLiteral(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
