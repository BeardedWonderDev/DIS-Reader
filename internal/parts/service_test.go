package parts

import (
	"context"
	"errors"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// stubServiceDB satisfies database.DB for service-level tests.
type stubServiceDB struct {
	part      *PartInventory
	parts     []PartInventory
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
	if s.part != nil {
		if p, ok := dest.(*PartInventory); ok {
			*p = *s.part
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
	if destSlice, ok := dest.(*[]PartInventory); ok {
		*destSlice = append([]PartInventory{}, s.parts...)
	}
	return nil
}

func (s *stubServiceDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	s.lastQuery = query
	return nil, nil
}

func TestServiceGetPartSuccess(t *testing.T) {
	db := &stubServiceDB{
		part: &PartInventory{
			Division:    "01",
			PartNumber:  " P123 ",
			Description: " Widget ",
		},
	}
	svc := NewService(db)

	spec, err := svc.GetPart(context.Background(), "01", "P123")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if spec.PartNumber != "P123" || spec.Description != "Widget" {
		t.Fatalf("expected trimmed values mapped to spec, got %+v", spec)
	}
	if db.lastQuery == "" {
		t.Fatalf("expected repository to issue a query")
	}
}

func TestServiceGetPartError(t *testing.T) {
	db := &stubServiceDB{getErr: errors.New("boom")}
	svc := NewService(db)

	if _, err := svc.GetPart(context.Background(), "01", "P123"); err == nil {
		t.Fatalf("expected error to propagate from repository")
	}
}

func TestServiceListPartsSuccess(t *testing.T) {
	db := &stubServiceDB{
		parts: []PartInventory{
			{Division: "01", PartNumber: "P1", Description: "One"},
			{Division: "02", PartNumber: "P2", Description: "Two"},
		},
	}
	svc := NewService(db)

	specs, err := svc.ListParts(context.Background(), types.ListParams{})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(specs) != 2 {
		t.Fatalf("expected two specs, got %d", len(specs))
	}
	if specs[0].PartNumber != "P1" || specs[1].PartNumber != "P2" {
		t.Fatalf("expected part numbers to map through, got %+v", specs)
	}
	if db.lastQuery == "" {
		t.Fatalf("expected repository to issue a query")
	}
}

func TestServiceListPartsError(t *testing.T) {
	db := &stubServiceDB{selectErr: errors.New("select failed")}
	svc := NewService(db)

	if _, err := svc.ListParts(context.Background(), types.ListParams{}); err == nil {
		t.Fatalf("expected error to propagate from repository")
	}
}
