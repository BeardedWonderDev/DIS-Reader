package unit

import "github.com/BeardedWonderDev/DIS-Reader/internal"

type UnitService struct {
	dis  internal.DISReaderPvtService
	repo UnitRepository
}

func NewUnitService(d internal.DISReaderPvtService) *UnitService {
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
