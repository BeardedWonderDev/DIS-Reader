package invoices

import (
	"context"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Service exposes typed helpers for invoice data.
type Service struct {
	db   database.DB
	repo Repository
}

func NewInvoiceService(db database.DB) *Service {
	return &Service{
		db:   db,
		repo: *NewRepository(db),
	}
}

func (s *Service) GetItem(ctx context.Context, documentNumber string, lineID string) (*types.InvoiceItemSpec, error) {
	item, err := s.repo.GetByDocumentAndLine(ctx, documentNumber, lineID)
	if err != nil {
		return nil, err
	}
	return item.ToInvoiceItemSpec(), nil
}

func (s *Service) ListItems(ctx context.Context, lp types.ListParams) ([]*types.InvoiceItemSpec, error) {
	rows, err := s.repo.ListInvoiceItems(ctx, lp)
	if err != nil {
		return nil, err
	}

	specs := make([]*types.InvoiceItemSpec, len(rows))
	for i, row := range rows {
		spec := row.ToInvoiceItemSpec()
		specs[i] = spec
	}
	return specs, nil
}

func (s *Service) GetWholeGoodsInvoice(ctx context.Context, invoiceNumber string, lineItemNumber string) (*types.WholeGoodsInvoiceSpec, error) {
	invoice, err := s.repo.GetWholeGoodsInvoice(ctx, invoiceNumber, lineItemNumber)
	if err != nil {
		return nil, err
	}
	return invoice.ToWholeGoodsInvoiceSpec(), nil
}

func (s *Service) ListWholeGoodsInvoices(ctx context.Context, lp types.ListParams) ([]*types.WholeGoodsInvoiceSpec, error) {
	rows, err := s.repo.ListWholeGoodsInvoices(ctx, lp)
	if err != nil {
		return nil, err
	}

	specs := make([]*types.WholeGoodsInvoiceSpec, len(rows))
	for i, row := range rows {
		specs[i] = row.ToWholeGoodsInvoiceSpec()
	}
	return specs, nil
}
