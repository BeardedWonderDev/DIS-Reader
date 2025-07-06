package unit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type UnitRepository struct {
	db             database.DB
	allowedColumns map[string]string
}

func NewUnitRepository(d database.DB) *UnitRepository {
	return &UnitRepository{
		db: d,
		allowedColumns: map[string]string{
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
		},
	}
}

func (repo *UnitRepository) GetByUnitNumber(ctx context.Context, unitNum string) (Model, error) {
	query := fmt.Sprintf("SELECT * FROM FILEC.DMUNITM1 WHERE UNIT = %s", unitNum)
	var unit Unit
	if err := repo.db.Get(ctx, &unit, query); err != nil {
		return nil, err
	}

	return &unit, nil
}

func (repo *UnitRepository) ListUnits(ctx context.Context, lp types.ListParams) ([]Unit, error) {
	base := "SELECT * FROM FILEC.DMUNITM1 WHERE 1=1"
	// Apply full-text search
	if lp.Query != "" {
		q := "%" + lp.Query + "%"
		base += " AND (UNIT LIKE '" + q + "' OR MAKE LIKE '" + q + "' OR MODEL LIKE '" + q + "')"
	}
	// Apply each filter
	for _, f := range lp.Filters {
		if f.Field == "indate" {
			expr := database.As400DateExpr("INDATE")
			t, ok := f.Value.(time.Time)
			if !ok {
				return nil, fmt.Errorf("expected time.Time for indate filter")
			}
			base += fmt.Sprintf(" AND %s %s DATE('%s')", expr, f.Operator, t.Format("2006-01-02"))
			continue
		}
		col, ok := repo.allowedColumns[f.Field]
		if !ok {
			return nil, fmt.Errorf("unsupported filter field %q", f.Field)
		}
		// Handle slice for IN/NOT IN
		switch v := f.Value.(type) {
		case []string:
			placeholders := make([]string, len(v))
			for i, s := range v {
				placeholders[i] = "'" + strings.ReplaceAll(s, "'", "''") + "'"
			}
			base += fmt.Sprintf(" AND %s %s (%s)", col, f.Operator, strings.Join(placeholders, ","))
		case string:
			base += fmt.Sprintf(" AND %s %s '%s'", col, f.Operator, v)
		default:
			base += fmt.Sprintf(" AND %s %s %v", col, f.Operator, v)
		}
	}
	// Sorting (handle AS/400 indate correctly)
	if lp.SortBy == "indate" {
		expr := database.As400DateExpr("INDATE")
		base += fmt.Sprintf(" ORDER BY %s %s", expr, lp.SortOrder.String())
	} else if sortCol, ok := repo.allowedColumns[lp.SortBy]; ok {
		base += fmt.Sprintf(" ORDER BY %s %s", sortCol, lp.SortOrder.String())
	}
	// Limit
	if lp.Limit > 0 {
		base += fmt.Sprintf(" LIMIT %d", lp.Limit)
	}
	// Execute
	var units []Unit
	if err := repo.db.Select(ctx, &units, base); err != nil {
		return nil, err
	}

	return units, nil
}
