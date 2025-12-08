package database

import (
	"context"
	"reflect"
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// fakeMultiTenantDB records the last call for assertion.
type fakeMultiTenantDB struct {
	lastTenant string
	lastQuery  string
	lastArgs   []interface{}
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
	f.record(query, tenant, args)
	return nil
}
func (f *fakeMultiTenantDB) Query(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	f.record(query, tenant, args)
	return nil, nil
}
func (f *fakeMultiTenantDB) QueryRow(ctx context.Context, query string, tenant string, args ...interface{}) (types.ResultRow, error) {
	f.record(query, tenant, args)
	return nil, nil
}
func (f *fakeMultiTenantDB) Select(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	f.record(query, tenant, args)
	return nil
}
func (f *fakeMultiTenantDB) QueryWithSource(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	f.record(query, tenant, args)
	return nil, nil
}

func (f *fakeMultiTenantDB) record(query, tenant string, args []interface{}) {
	f.lastTenant = tenant
	f.lastQuery = query
	f.lastArgs = args
}

func TestBindTenantForwardsTenantAndArgs(t *testing.T) {
	mt := &fakeMultiTenantDB{}
	bound := BindTenant(mt, "tenant-123")

	ctx := context.Background()

	if err := bound.Connect(ctx); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if mt.lastTenant != "tenant-123" {
		t.Fatalf("expected tenant for Connect, got %q", mt.lastTenant)
	}

	_, _ = bound.Query(ctx, "SELECT 1 WHERE a = ?", 42)
	if mt.lastTenant != "tenant-123" {
		t.Fatalf("expected tenant for Query, got %q", mt.lastTenant)
	}
	if mt.lastQuery != "SELECT 1 WHERE a = ?" {
		t.Fatalf("unexpected query recorded: %q", mt.lastQuery)
	}
	if !reflect.DeepEqual(mt.lastArgs, []interface{}{42}) {
		t.Fatalf("unexpected args, got %#v", mt.lastArgs)
	}

	if err := bound.Disconnect(ctx); err != nil {
		t.Fatalf("disconnect: %v", err)
	}
	if mt.lastTenant != "tenant-123" {
		t.Fatalf("expected tenant for Disconnect, got %q", mt.lastTenant)
	}
}

func TestBindTenantCoversAllMethods(t *testing.T) {
	mt := &fakeMultiTenantDB{}
	db := BindTenant(mt, "tenant-x")
	ctx := context.Background()

	// lifecycle helpers
	_ = db.StartJDBCRunner()
	_ = db.StopJDBCRunner()
	_ = db.PingService(ctx)
	_ = db.PingDatabase(ctx)

	// data helpers
	_ = db.Get(ctx, nil, "q1")
	_, _ = db.Query(ctx, "q2")
	_, _ = db.QueryRow(ctx, "q3")
	_ = db.Select(ctx, nil, "q4")
	_, _ = db.QueryWithSource(ctx, "q5")

	if mt.lastTenant != "tenant-x" {
		t.Fatalf("expected tenant propagated, got %q", mt.lastTenant)
	}
}
