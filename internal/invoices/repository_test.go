package invoices

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type stubDB struct {
	lastQuery string
}

func (s *stubDB) StartJDBCRunner() error                 { return nil }
func (s *stubDB) StopJDBCRunner() error                  { return nil }
func (s *stubDB) Connect(ctx context.Context) error      { return nil }
func (s *stubDB) Disconnect(ctx context.Context) error   { return nil }
func (s *stubDB) PingService(ctx context.Context) error  { return nil }
func (s *stubDB) PingDatabase(ctx context.Context) error { return nil }
func (s *stubDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	s.lastQuery = query
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
