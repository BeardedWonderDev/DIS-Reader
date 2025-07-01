package unit

import "github.com/BeardedWonderDev/DIS-Reader/types"

type UnitService struct {
	dis  types.DISReaderService
	repo UnitRepository
}

func NewUnitService(d types.DISReaderService) *UnitService {
	return &UnitService{
		dis:  d,
		repo: *NewRoleRepository(d),
	}
}

func (u *UnitService) GetByUnitNumber(unitNum string) (*UnitSpec, error) {
	unit, err := u.repo.GetByUnitNumber(unitNum)
	if err != nil {
		return nil, err
	}

	return unit.ToUnitSpec(), nil
}
