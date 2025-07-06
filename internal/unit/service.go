package unit

import (
	"context"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type UnitService struct {
	db   database.DB
	repo UnitRepository
}

func NewUnitService(d database.DB) *UnitService {
	return &UnitService{
		db:   d,
		repo: *NewUnitRepository(d),
	}
}

func (u *UnitService) GetByUnitNumber(ctx context.Context, unitNum string) (*types.UnitSpec, error) {
	unit, err := u.repo.GetByUnitNumber(ctx, unitNum)
	if err != nil {
		return nil, err
	}

	return unit.ToUnitSpec(), nil
}

// ListUnits fetches a list of units based on ListParams and returns their specs.
func (u *UnitService) ListUnits(ctx context.Context, lp types.ListParams) ([]*types.UnitSpec, error) {
	units, err := u.repo.ListUnits(ctx, lp)
	if err != nil {
		return nil, err
	}

	specs := make([]*types.UnitSpec, len(units))
	for i, unit := range units {
		specs[i] = unit.ToUnitSpec()
	}
	return specs, nil
}
