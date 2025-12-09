package unit

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
	selectErr error
	getErr    error
	getUnit   *Unit
	units     []Unit
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
	if s.getUnit != nil {
		if u, ok := dest.(*Unit); ok {
			*u = *s.getUnit
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
	if u, ok := dest.(*[]Unit); ok {
		*u = append(*u, s.units...)
	}
	return nil
}
func (s *stubDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}

func TestListUnitsBuildsQuery(t *testing.T) {
	db := &stubDB{units: []Unit{{UnitID: "1"}}}
	repo := NewUnitRepository(db)
	lp := types.ListParams{Query: "Ford", SortBy: "UNIT", SortOrder: types.SortOrderDesc, Limit: 10}

	if _, err := repo.ListUnits(context.Background(), lp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q := db.lastQuery
	if !strings.Contains(q, "UNIT LIKE '%Ford%'") || !strings.Contains(q, "MAKE LIKE '%Ford%'") {
		t.Fatalf("expected search clauses in query: %s", q)
	}
	if !strings.Contains(q, "ORDER BY UNIT DESC") {
		t.Fatalf("expected order by clause: %s", q)
	}
	if !strings.Contains(q, "LIMIT 10") {
		t.Fatalf("expected limit clause: %s", q)
	}
}

func TestListUnitsIndateFilter(t *testing.T) {
	db := &stubDB{units: []Unit{}}
	repo := NewUnitRepository(db)
	lp := types.ListParams{
		SortBy:    "INDATE",
		SortOrder: types.SortOrderAsc,
		Filters:   []types.Filter{{Field: "indate", Operator: ">=", Value: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)}},
	}
	if _, err := repo.ListUnits(context.Background(), lp); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(db.lastQuery, "INDATE") {
		t.Fatalf("expected indate expression in query: %s", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "DATE('2024-01-02')") {
		t.Fatalf("expected date literal in query: %s", db.lastQuery)
	}
	if !strings.Contains(db.lastQuery, "ORDER BY") {
		t.Fatalf("expected order clause: %s", db.lastQuery)
	}
}

func TestListUnitsIndateFilterTypeGuard(t *testing.T) {
	db := &stubDB{}
	repo := NewUnitRepository(db)
	lp := types.ListParams{
		Filters: []types.Filter{{Field: "indate", Operator: "=", Value: "not-a-time"}},
	}
	_, err := repo.ListUnits(context.Background(), lp)
	if err == nil {
		t.Fatalf("expected error for non-time filter value")
	}
}

func TestGetByUnitNumberSuccess(t *testing.T) {
	db := &stubDB{getUnit: &Unit{UnitID: "123"}}
	repo := NewUnitRepository(db)

	model, err := repo.GetByUnitNumber(context.Background(), "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.GetUnitID() != "123" {
		t.Fatalf("expected unit id mapped, got %s", model.GetUnitID())
	}
	if !strings.Contains(db.lastQuery, "FILEC.DMUNITM1") || !strings.Contains(db.lastQuery, "123") {
		t.Fatalf("expected select against unit table, got %s", db.lastQuery)
	}
}

func TestGetByUnitNumberError(t *testing.T) {
	db := &stubDB{getErr: fmt.Errorf("boom")}
	repo := NewUnitRepository(db)
	if _, err := repo.GetByUnitNumber(context.Background(), "123"); err == nil {
		t.Fatalf("expected error from DB get")
	}
}
