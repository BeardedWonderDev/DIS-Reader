package service

import (
	"context"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// stubMultiTenantDB records the last tenant passed to lifecycle/health methods.
type stubMultiTenantDB struct {
	last string
}

func (s *stubMultiTenantDB) StartJDBCRunner(tenant string) error { s.last = tenant; return nil }
func (s *stubMultiTenantDB) StopJDBCRunner(tenant string) error  { s.last = tenant; return nil }
func (s *stubMultiTenantDB) Connect(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) Disconnect(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) PingService(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) PingDatabase(ctx context.Context, tenant string) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) Get(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) Query(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	s.last = tenant
	return nil, nil
}
func (s *stubMultiTenantDB) QueryRow(ctx context.Context, query string, tenant string, args ...interface{}) (types.ResultRow, error) {
	s.last = tenant
	return nil, nil
}
func (s *stubMultiTenantDB) Select(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	s.last = tenant
	return nil
}
func (s *stubMultiTenantDB) QueryWithSource(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	s.last = tenant
	return nil, nil
}

func TestResolveTenantFallsBackToDefault(t *testing.T) {
	mt := &stubMultiTenantDB{}
	s := NewRemoteService(&types.DISConfig{}, nil, mt, bridge.NewInMemoryRegistry(), nil)
	s.WithDefaultTenant("tenant-default")

	if err := s.Connect(context.Background(), ""); err != nil {
		t.Fatalf("expected nil error with default tenant, got %v", err)
	}
	if mt.last != "tenant-default" {
		t.Fatalf("expected default tenant to be used, got %q", mt.last)
	}
}

func TestResolveTenantPrefersExplicit(t *testing.T) {
	mt := &stubMultiTenantDB{}
	s := NewRemoteService(&types.DISConfig{}, nil, mt, bridge.NewInMemoryRegistry(), nil)
	s.WithDefaultTenant("tenant-default")

	if err := s.Connect(context.Background(), "tenant-explicit"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if mt.last != "tenant-explicit" {
		t.Fatalf("expected explicit tenant to win, got %q", mt.last)
	}
}

func TestResolveTenantErrorWhenMissing(t *testing.T) {
	mt := &stubMultiTenantDB{}
	s := NewRemoteService(&types.DISConfig{}, nil, mt, bridge.NewInMemoryRegistry(), nil)

	if err := s.Connect(context.Background(), ""); err == nil {
		t.Fatalf("expected error when no tenant and no default")
	}
}
