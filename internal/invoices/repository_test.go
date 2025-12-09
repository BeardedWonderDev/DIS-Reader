package invoices

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type stubDB struct {
	lastQuery string
	getErr    error
	selectErr error
	getItem   *InvoiceItem
	getWG     *WholeGoodsInvoice
}

func (s *stubDB) StartJDBCRunner() error                 { return nil }
func (s *stubDB) StopJDBCRunner() error                  { return nil }
func (s *stubDB) Connect(ctx context.Context) error      { return nil }
func (s *stubDB) Disconnect(ctx context.Context) error   { return nil }
func (s *stubDB) PingService(ctx context.Context) error  { return nil }
func (s *stubDB) PingDatabase(ctx context.Context) error { return nil }
func (s *stubDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	s.lastQuery = query
	if s.getErr != nil {
		return s.getErr
	}
	switch d := dest.(type) {
	case *InvoiceItem:
		if s.getItem != nil {
			*d = *s.getItem
		}
	case *WholeGoodsInvoice:
		if s.getWG != nil {
			*d = *s.getWG
		}
	}
	return nil
}
func (s *stubDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}
func (s *stubDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}
func (s *stubDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	s.lastQuery = query
	if s.selectErr != nil {
		return s.selectErr
	}
	return nil
}
func (s *stubDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}

func TestListInvoiceItemsQueryAndSort(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	lp := types.ListParams{Query: "ABC", SortBy: "AHPDTE", SortOrder: types.SortOrderAsc, Limit: 3}
	if _, err := repo.ListInvoiceItems(context.Background(), lp); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(db.lastQuery, "AHDOC LIKE '%ABC%'") {
		t.Fatalf("expected search clause: %s", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "ORDER BY ") {
		t.Fatalf("expected order clause: %s", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "LIMIT 3") {
		t.Fatalf("expected limit: %s", db.lastQuery)
	}
}

func TestListInvoiceItemsDateFilter(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	lp := types.ListParams{Filters: []types.Filter{{Field: "postingdate", Operator: ">=", Value: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)}}}
	if _, err := repo.ListInvoiceItems(context.Background(), lp); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(db.lastQuery, "AHPDTE") || !strings.Contains(db.lastQuery, "DATE('2024-01-02')") {
		t.Fatalf("expected date filter: %s", db.lastQuery)
	}
}

func TestListInvoiceItemsInvalidDateFilterType(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	lp := types.ListParams{Filters: []types.Filter{{Field: "postingdate", Operator: ">=", Value: "bad"}}}
	if _, err := repo.ListInvoiceItems(context.Background(), lp); err == nil {
		t.Fatalf("expected error for non-time postingdate filter")
	}
}

func TestWholeGoodsListAndValidation(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	lp := types.ListParams{Query: "abc", SortOrder: types.SortOrderDesc, Limit: 2}
	if _, err := repo.ListWholeGoodsInvoices(context.Background(), lp); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(db.lastQuery, "DSHUA LIKE '%abc%'") {
		t.Fatalf("expected search in query: %s", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "ORDER BY") {
		t.Fatalf("expected order clause: %s", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "LIMIT 2") {
		t.Fatalf("expected limit: %s", db.lastQuery)
	}
}

func TestGetWholeGoodsInvoiceValidation(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	if _, err := repo.GetWholeGoodsInvoice(context.Background(), "", "1"); err == nil {
		t.Fatalf("expected error for empty invoice number")
	}
	if _, err := repo.GetWholeGoodsInvoice(context.Background(), "INV", "abc"); err == nil {
		t.Fatalf("expected numeric validation error for line item")
	}
}

func TestNumericLiteralValidation(t *testing.T) {
	if _, err := numericLiteral(""); err == nil {
		t.Fatalf("expected error for empty line item")
	}
	if _, err := numericLiteral("  "); err == nil {
		t.Fatalf("expected error for whitespace line item")
	}
	if _, err := numericLiteral("abc"); err == nil {
		t.Fatalf("expected error for non-numeric line item")
	}
	val, err := numericLiteral("42")
	if err != nil || val != "42" {
		t.Fatalf("expected numeric literal 42, got %s (err=%v)", val, err)
	}
}

func TestGetByDocumentAndLineSuccess(t *testing.T) {
	db := &stubDB{getItem: &InvoiceItem{DocumentNumber: "DOC1", LineID: "1"}}
	repo := NewRepository(db)
	model, err := repo.GetByDocumentAndLine(context.Background(), "DOC1", "1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	item, ok := model.(*InvoiceItem)
	if !ok {
		t.Fatalf("expected *InvoiceItem type")
	}
	if item.DocumentNumber != "DOC1" || item.LineID != "1" {
		t.Fatalf("expected mapped invoice item, got %+v", item)
	}
	if !strings.Contains(db.lastQuery, "FILEC.IAH") || !strings.Contains(db.lastQuery, "DOC1") {
		t.Fatalf("expected query against invoice table: %s", db.lastQuery)
	}
}

func TestGetByDocumentAndLineError(t *testing.T) {
	db := &stubDB{getErr: fmt.Errorf("boom")}
	repo := NewRepository(db)
	if _, err := repo.GetByDocumentAndLine(context.Background(), "DOC1", "1"); err == nil {
		t.Fatalf("expected error propagation from DB")
	}
}

func TestGetWholeGoodsInvoiceSuccess(t *testing.T) {
	db := &stubDB{getWG: &WholeGoodsInvoice{InvoiceNumber: "INV1", LineItemNumber: 2}}
	repo := NewRepository(db)
	row, err := repo.GetWholeGoodsInvoice(context.Background(), "INV1", "2")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if row.InvoiceNumber != "INV1" || row.LineItemNumber != 2 {
		t.Fatalf("expected mapped whole goods row, got %+v", row)
	}
	if !strings.Contains(db.lastQuery, "CUSINV") || !strings.Contains(db.lastQuery, "INV1") {
		t.Fatalf("expected query against CUSINV: %s", db.lastQuery)
	}
}

func TestGetWholeGoodsInvoiceError(t *testing.T) {
	db := &stubDB{getErr: fmt.Errorf("fail")}
	repo := NewRepository(db)
	if _, err := repo.GetWholeGoodsInvoice(context.Background(), "INV1", "2"); err == nil {
		t.Fatalf("expected error propagation")
	}
}

func TestFormatAndEscapeHelpers(t *testing.T) {
	if got := escapeLiteral("O'Hare"); got != "O''Hare" {
		t.Fatalf("expected quote escape, got %s", got)
	}
	if got := formatFilterValue(" abc "); got != "' abc '" {
		t.Fatalf("expected string quoted, got %s", got)
	}
	if got := formatFilterValue(5); got != "5" {
		t.Fatalf("expected numeric passthrough, got %s", got)
	}
}
