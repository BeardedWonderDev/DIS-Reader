package database

import (
	"context"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// tenantBinding adapts a MultiTenantDB to the DB interface by injecting a fixed tenant.
// Domain services and UI continue to use DB; remote mode supplies tenant via this adapter.
type tenantBinding struct {
	tenant string
	mt     MultiTenantDB
}

// BindTenant returns a DB that forwards calls to the provided MultiTenantDB using tenant.
func BindTenant(mt MultiTenantDB, tenant string) DB {
	return &tenantBinding{tenant: tenant, mt: mt}
}

func (t *tenantBinding) StartJDBCRunner() error { return t.mt.StartJDBCRunner() }
func (t *tenantBinding) StopJDBCRunner() error  { return t.mt.StopJDBCRunner() }

func (t *tenantBinding) Connect(ctx context.Context) error    { return t.mt.Connect(ctx, t.tenant) }
func (t *tenantBinding) Disconnect(ctx context.Context) error { return t.mt.Disconnect(ctx, t.tenant) }
func (t *tenantBinding) PingService(ctx context.Context) error {
	return t.mt.PingService(ctx, t.tenant)
}
func (t *tenantBinding) PingDatabase(ctx context.Context) error {
	return t.mt.PingDatabase(ctx, t.tenant)
}

func (t *tenantBinding) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return t.mt.Get(ctx, dest, query, t.tenant, args...)
}

func (t *tenantBinding) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return t.mt.Query(ctx, query, t.tenant, args...)
}

func (t *tenantBinding) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	return t.mt.QueryRow(ctx, query, t.tenant, args...)
}

func (t *tenantBinding) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return t.mt.Select(ctx, dest, query, t.tenant, args...)
}

func (t *tenantBinding) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	return t.mt.QueryWithSource(ctx, query, t.tenant, args...)
}
