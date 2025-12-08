package invoices

import (
	"context"
	"errors"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// stubServiceDB implements database.DB for service-layer tests.
type stubServiceDB struct {
	item         *InvoiceItem
	items        []InvoiceItem
	wgInvoice    *WholeGoodsInvoice
	getErr       error
	selectErr    error
	lastSelect   string
	lastGetQuery string
}

func (s *stubServiceDB) StartJDBCRunner() error                 { return nil }
func (s *stubServiceDB) StopJDBCRunner() error                  { return nil }
func (s *stubServiceDB) Connect(ctx context.Context) error      { return nil }
func (s *stubServiceDB) Disconnect(ctx context.Context) error   { return nil }
func (s *stubServiceDB) PingService(ctx context.Context) error  { return nil }
func (s *stubServiceDB) PingDatabase(ctx context.Context) error { return nil }
func (s *stubServiceDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	s.lastGetQuery = query
	if s.getErr != nil {
		return s.getErr
	}
	switch d := dest.(type) {
	case *InvoiceItem:
		if s.item != nil {
			*d = *s.item
		}
	case *WholeGoodsInvoice:
		if s.wgInvoice != nil {
			*d = *s.wgInvoice
		}
	}
	return nil
}
func (s *stubServiceDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return nil, nil
}
func (s *stubServiceDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	return nil, nil
}
func (s *stubServiceDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	s.lastSelect = query
	if s.selectErr != nil {
		return s.selectErr
	}
	if list, ok := dest.(*[]InvoiceItem); ok {
		*list = append(*list, s.items...)
	}
	if list, ok := dest.(*[]WholeGoodsInvoice); ok {
		*list = append(*list, *s.wgInvoice)
	}
	return nil
}
func (s *stubServiceDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return nil, nil
}

func TestServiceGetItemSuccess(t *testing.T) {
	db := &stubServiceDB{item: &InvoiceItem{DocumentNumber: "D1", LineID: "L1", Division: "01"}}
	svc := NewInvoiceService(db)
	spec, err := svc.GetItem(context.Background(), "D1", "L1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.DocumentNumber != "D1" || spec.LineID != "L1" {
		t.Fatalf("expected spec mapping, got %+v", spec)
	}
	if db.lastGetQuery == "" {
		t.Fatalf("expected repository Get to be called")
	}
}

func TestServiceGetItemError(t *testing.T) {
	db := &stubServiceDB{getErr: errors.New("boom")}
	svc := NewInvoiceService(db)
	if _, err := svc.GetItem(context.Background(), "D1", "L1"); err == nil {
		t.Fatalf("expected error to propagate")
	}
}

func TestServiceListItems(t *testing.T) {
	db := &stubServiceDB{items: []InvoiceItem{{DocumentNumber: "D1"}, {DocumentNumber: "D2"}}}
	svc := NewInvoiceService(db)
	specs, err := svc.ListItems(context.Background(), types.ListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 2 || specs[0].DocumentNumber != "D1" || specs[1].DocumentNumber != "D2" {
		t.Fatalf("expected mapped specs, got %+v", specs)
	}
}

func TestServiceListItemsError(t *testing.T) {
	db := &stubServiceDB{selectErr: errors.New("select fail")}
	svc := NewInvoiceService(db)
	if _, err := svc.ListItems(context.Background(), types.ListParams{}); err == nil {
		t.Fatalf("expected error to propagate")
	}
}

func TestServiceGetWholeGoodsInvoice(t *testing.T) {
	db := &stubServiceDB{wgInvoice: &WholeGoodsInvoice{InvoiceNumber: "INV", LineItemNumber: 1}}
	svc := NewInvoiceService(db)
	spec, err := svc.GetWholeGoodsInvoice(context.Background(), "INV", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.InvoiceNumber != "INV" || spec.LineItemNumber != 1 {
		t.Fatalf("expected spec mapping, got %+v", spec)
	}
}

func TestServiceGetWholeGoodsInvoiceError(t *testing.T) {
	db := &stubServiceDB{getErr: errors.New("nope")}
	svc := NewInvoiceService(db)
	if _, err := svc.GetWholeGoodsInvoice(context.Background(), "INV", "1"); err == nil {
		t.Fatalf("expected error propagation")
	}
}

func TestServiceListWholeGoodsInvoices(t *testing.T) {
	db := &stubServiceDB{wgInvoice: &WholeGoodsInvoice{InvoiceNumber: "INV"}} // Select appends one
	svc := NewInvoiceService(db)
	specs, err := svc.ListWholeGoodsInvoices(context.Background(), types.ListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 1 || specs[0].InvoiceNumber != "INV" {
		t.Fatalf("expected mapped whole goods specs, got %+v", specs)
	}
}

func TestServiceListWholeGoodsInvoicesError(t *testing.T) {
	db := &stubServiceDB{selectErr: errors.New("fail")}
	svc := NewInvoiceService(db)
	if _, err := svc.ListWholeGoodsInvoices(context.Background(), types.ListParams{}); err == nil {
		t.Fatalf("expected error propagation")
	}
}
