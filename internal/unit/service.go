package unit

import (
	"context"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
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

func (u *UnitService) GetByUnitNumber(ctx context.Context, unitNum string) (*UnitSpec, error) {
	unit, err := u.repo.GetByUnitNumber(ctx, unitNum)
	if err != nil {
		return nil, err
	}

	return unit.ToUnitSpec(), nil
}
