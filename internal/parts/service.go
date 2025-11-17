package parts

import (
	"context"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Service exposes the typed parts inventory helpers.
type Service struct {
	db   database.DB
	repo Repository
}

func NewService(db database.DB) *Service {
	return &Service{
		db:   db,
		repo: *NewRepository(db),
	}
}

func (s *Service) GetPart(ctx context.Context, division, partNumber string) (*types.PartInventorySpec, error) {
	row, err := s.repo.GetPart(ctx, division, partNumber)
	if err != nil {
		return nil, err
	}
	return row.ToPartInventorySpec(), nil
}

func (s *Service) ListParts(ctx context.Context, lp types.ListParams) ([]*types.PartInventorySpec, error) {
	rows, err := s.repo.ListParts(ctx, lp)
	if err != nil {
		return nil, err
	}

	specs := make([]*types.PartInventorySpec, len(rows))
	for i := range rows {
		rec := rows[i]
		specs[i] = rec.ToPartInventorySpec()
	}
	return specs, nil
}
