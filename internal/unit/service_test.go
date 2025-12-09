package unit

import (
	"context"
	"errors"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type stubServiceDB struct {
	getUnit   *Unit
	units     []Unit
	getErr    error
	selectErr error
	lastQuery string
}

func (s *stubServiceDB) StartJDBCRunner() error                 { return nil }
func (s *stubServiceDB) StopJDBCRunner() error                  { return nil }
func (s *stubServiceDB) Connect(ctx context.Context) error      { return nil }
func (s *stubServiceDB) Disconnect(ctx context.Context) error   { return nil }
func (s *stubServiceDB) PingService(ctx context.Context) error  { return nil }
func (s *stubServiceDB) PingDatabase(ctx context.Context) error { return nil }
func (s *stubServiceDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
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
func (s *stubServiceDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}
func (s *stubServiceDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}
func (s *stubServiceDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	s.lastQuery = query
	if s.selectErr != nil {
		return s.selectErr
	}
	if list, ok := dest.(*[]Unit); ok {
		*list = append(*list, s.units...)
	}
	return nil
}
func (s *stubServiceDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}

func TestUnitServiceGetByUnitNumberSuccess(t *testing.T) {
	db := &stubServiceDB{getUnit: &Unit{UnitID: "U123", Make: "Make"}}
	svc := NewUnitService(db)

	spec, err := svc.GetByUnitNumber(context.Background(), "U123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.UnitID != "U123" || spec.Make != "Make" {
		t.Fatalf("expected spec to map fields, got %+v", spec)
	}
	if db.lastQuery == "" {
		t.Fatalf("expected repository query to be issued")
	}
}

func TestUnitServiceGetByUnitNumberError(t *testing.T) {
	db := &stubServiceDB{getErr: errors.New("boom")}
	svc := NewUnitService(db)

	if _, err := svc.GetByUnitNumber(context.Background(), "U123"); err == nil {
		t.Fatalf("expected error propagated from repository")
	}
}

func TestUnitServiceListUnitsSuccess(t *testing.T) {
	db := &stubServiceDB{
		units: []Unit{
			{UnitID: "U1", Make: "M1"},
			{UnitID: "U2", Make: "M2"},
		},
	}
	svc := NewUnitService(db)

	specs, err := svc.ListUnits(context.Background(), types.ListParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(specs) != 2 || specs[0].UnitID != "U1" || specs[1].UnitID != "U2" {
		t.Fatalf("expected mapped specs, got %+v", specs)
	}
}

func TestUnitServiceListUnitsError(t *testing.T) {
	db := &stubServiceDB{selectErr: errors.New("select fail")}
	svc := NewUnitService(db)

	if _, err := svc.ListUnits(context.Background(), types.ListParams{}); err == nil {
		t.Fatalf("expected error propagated from repository")
	}
}
