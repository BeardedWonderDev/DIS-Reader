package database

import (
	"context"
	"errors"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

type fakeMultiTenantDB struct {
	lastTenant string
}

func (f *fakeMultiTenantDB) StartJDBCRunner(tenant string) error { f.lastTenant = tenant; return nil }
func (f *fakeMultiTenantDB) StopJDBCRunner(tenant string) error  { f.lastTenant = tenant; return nil }
func (f *fakeMultiTenantDB) Connect(ctx context.Context, tenant string) error {
	f.lastTenant = tenant
	return nil
}
func (f *fakeMultiTenantDB) Disconnect(ctx context.Context, tenant string) error {
	f.lastTenant = tenant
	return nil
}
func (f *fakeMultiTenantDB) PingService(ctx context.Context, tenant string) error {
	f.lastTenant = tenant
	return nil
}
func (f *fakeMultiTenantDB) PingDatabase(ctx context.Context, tenant string) error {
	f.lastTenant = tenant
	return nil
}
func (f *fakeMultiTenantDB) Get(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	f.lastTenant = tenant
	if destMap, ok := dest.(*map[string]string); ok {
		(*destMap)["query"] = query
	}
	return nil
}
func (f *fakeMultiTenantDB) Query(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	f.lastTenant = tenant
	return []types.ResultRow{{"query": query}}, nil
}
func (f *fakeMultiTenantDB) QueryRow(ctx context.Context, query string, tenant string, args ...interface{}) (types.ResultRow, error) {
	f.lastTenant = tenant
	return types.ResultRow{"query": query}, nil
}
func (f *fakeMultiTenantDB) Select(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	f.lastTenant = tenant
	if slice, ok := dest.(*[]types.ResultRow); ok {
		*slice = []types.ResultRow{{"query": query}}
		return nil
	}
	return errors.New("dest not slice")
}
func (f *fakeMultiTenantDB) QueryWithSource(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	f.lastTenant = tenant
	return []types.ResultRow{{"query": query}}, nil
}

func TestBindTenantForwardsTenant(t *testing.T) {
	mt := &fakeMultiTenantDB{}
	db := BindTenant(mt, "tenant-123")

	if err := db.Connect(context.Background()); err != nil {
		t.Fatalf("connect err: %v", err)
	}
	if mt.lastTenant != "tenant-123" {
		t.Fatalf("expected tenant propagated, got %s", mt.lastTenant)
	}

	rows, err := db.Query(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("query err: %v", err)
	}
	if rows[0]["query"] != "SELECT 1" {
		t.Fatalf("expected query to pass through")
	}

	var dest map[string]string = map[string]string{}
	if err := db.Get(context.Background(), &dest, "SELECT 2"); err != nil {
		t.Fatalf("get err: %v", err)
	}
	if dest["query"] != "SELECT 2" {
		t.Fatalf("expected dest to be filled from fake")
	}
}
