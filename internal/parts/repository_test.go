package parts

import (
	"context"
	"strings"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type stubDB struct {
	lastQuery string
	selectErr error
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
	if s.selectErr != nil {
		return s.selectErr
	}
	return nil
}
func (s *stubDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}

func TestListPartsBuildsQuery(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	lp := types.ListParams{Query: "bolt", SortBy: "PIMDES", SortOrder: types.SortOrderDesc, Limit: 5}

	if _, err := repo.ListParts(context.Background(), lp); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	q := db.lastQuery
	if !strings.Contains(q, "PIMDES LIKE '%bolt%'") || !strings.Contains(q, "PIMPRT LIKE '%bolt%'") {
		t.Fatalf("expected LIKE search across columns: %s", q)
	}
	if !strings.Contains(q, "ORDER BY PIMDES DESC") {
		t.Fatalf("expected sort order: %s", q)
	}
	if !strings.Contains(q, "LIMIT 5") {
		t.Fatalf("expected limit: %s", q)
	}
}

func TestListPartsFiltersEscape(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	lp := types.ListParams{
		Filters: []types.Filter{{Column: "PIMVEN", Operator: "=", Value: "ACME"}},
	}
	if _, err := repo.ListParts(context.Background(), lp); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !strings.Contains(db.lastQuery, "PIMVEN = 'ACME'") {
		t.Fatalf("expected filter applied: %s", db.lastQuery)
	}
}

func TestGetPartRequiresInputs(t *testing.T) {
	db := &stubDB{}
	repo := NewRepository(db)
	if _, err := repo.GetPart(context.Background(), "", "123"); err == nil {
		t.Fatalf("expected error for missing division")
	}
	if _, err := repo.GetPart(context.Background(), "1", ""); err == nil {
		t.Fatalf("expected error for missing part number")
	}
}
