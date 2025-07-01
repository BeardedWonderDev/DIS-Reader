package unit

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/service"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type UnitRepository struct {
	DIS            types.DISReaderService
	allowedColumns map[string]string
}

func NewRoleRepository(d types.DISReaderService) *UnitRepository {
	return &UnitRepository{
		DIS: d,
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

func (repo *UnitRepository) GetByUnitNumber(unitNum string) (Model, error) {
	query := fmt.Sprintf("SELECT * FROM FILEC.DMUNITM1 WHERE UNIT = %s", unitNum)
	results, err := repo.DIS.Query(query)
	if err != nil {
		return nil, err
	}

	units, err := database.DecodeRows[Unit](results)
	if err != nil {
		return nil, err
	}

	if len(units) > 1 {
		return nil, fmt.Errorf("more than one result was returned: %v", units)
	}

	unit := units[len(units)-1]

	return &unit, nil
}

func (repo *UnitRepository) ListUnits(ctx context.Context, lp service.ListParams) ([]Unit, error) {
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
	rows, err := repo.DIS.Query(base)
	if err != nil {
		return nil, err
	}
	units, err := database.DecodeRows[Unit](rows)
	if err != nil {
		return nil, err
	}
	return units, nil
}
