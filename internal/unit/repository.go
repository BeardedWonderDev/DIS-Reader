package unit

import (
	"context"
	"fmt"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type UnitRepository struct {
	db database.DB
}

func NewUnitRepository(d database.DB) *UnitRepository {
	return &UnitRepository{
		db: d,
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

		base += fmt.Sprintf(" AND %s %s %v", f.Column, f.Operator, f.Value)
	}
	// Sorting (handle AS/400 indate correctly)
	if lp.SortBy == "INDATE" {
		expr := database.As400DateExpr("INDATE")
		base += fmt.Sprintf(" ORDER BY %s %s", expr, lp.SortOrder.String())
	} else {
		base += fmt.Sprintf(" ORDER BY %s %s", lp.SortBy, lp.SortOrder.String())
	}
	// Limit
	if lp.Limit > 0 {
		base += fmt.Sprintf(" LIMIT %d", lp.Limit)
	}

	fmt.Println(base)
	// Execute
	var units []Unit
	if err := repo.db.Select(ctx, &units, base); err != nil {
		return nil, err
	}

	return units, nil
}
