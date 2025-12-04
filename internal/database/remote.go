package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

// RemoteDB satisfies the MultiTenantDB interface by proxying calls to a remote agent over the bridge.
type RemoteDB struct {
	registry       bridge.AgentRegistry
	Logger         *slog.Logger
	maxRows        int
	maxResultBytes int64
}

func NewRemoteDB(registry bridge.AgentRegistry, logger *slog.Logger, maxRows int, maxBytes int64) *RemoteDB {
	if maxRows <= 0 {
		maxRows = 1000
	}
	return &RemoteDB{
		registry:       registry,
		Logger:         logger,
		maxRows:        maxRows,
		maxResultBytes: maxBytes,
	}
}

// Lifecycle methods are no-ops for remote connections.
func (r *RemoteDB) StartJDBCRunner() error { return nil }
func (r *RemoteDB) StopJDBCRunner() error  { return nil }

func (r *RemoteDB) Connect(ctx context.Context, tenant string) error {
	_, err := r.sendJob(ctx, tenant, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_CONNECT,
	})
	return err
}

func (r *RemoteDB) Disconnect(ctx context.Context, tenant string) error {
	_, err := r.sendJob(ctx, tenant, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_DISCONNECT,
	})
	return err
}

func (r *RemoteDB) PingService(ctx context.Context, tenant string) error {
	_, err := r.sendJob(ctx, tenant, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_PING_SERVICE,
	})
	return err
}

func (r *RemoteDB) PingDatabase(ctx context.Context, tenant string) error {
	_, err := r.sendJob(ctx, tenant, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_PING_DATABASE,
	})
	return err
}

func (r *RemoteDB) Query(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	req := &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_QUERY,
		Sql:   query,
	}
	res, err := r.sendJob(ctx, tenant, req)
	if err != nil {
		return nil, err
	}
	rows := protoRowsToResultRows(res.Rows)
	return r.enforceLimits(rows), nil
}

func (r *RemoteDB) QueryRow(ctx context.Context, query string, tenant string, args ...interface{}) (types.ResultRow, error) {
	rows, err := r.Query(ctx, query, tenant, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows returned for query %q", query)
	}
	return rows[0], nil
}

func (r *RemoteDB) Select(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	rows, err := r.Query(ctx, query, tenant, args...)
	if err != nil {
		return err
	}
	return r.decodeResult(rows, dest)
}

func (r *RemoteDB) Get(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error {
	row, err := r.QueryRow(ctx, query, tenant, args...)
	if err != nil {
		return err
	}
	return r.decodeResult(row, dest)
}

func (r *RemoteDB) QueryWithSource(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error) {
	req := &bridgeproto.JobRequest{
		JobId:      uuid.New().String(),
		Kind:       bridgeproto.JobKind_JOB_KIND_QUERY,
		Sql:        query,
		IncludeSrc: true,
	}
	res, err := r.sendJob(ctx, tenant, req)
	if err != nil {
		return nil, err
	}
	rows := protoRowsToResultRows(res.Rows)
	return r.enforceLimits(rows), nil
}

// --- helpers ---

func (r *RemoteDB) sendJob(ctx context.Context, tenant string, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	if tenant == "" {
		return nil, fmt.Errorf("tenant is required for remote DB call")
	}
	agent, err := r.registry.Pick(ctx, tenant)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	res, err := agent.SendJob(ctx, req)
	r.registry.ObserveQuery(tenant, agent.AgentID(), time.Since(start))
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("nil job result")
	}
	if res.Status == bridgeproto.Status_STATUS_ERROR {
		return nil, fmt.Errorf("remote error: %s", res.Message)
	}
	return res, nil
}

func protoRowsToResultRows(rows []*bridgeproto.Row) []types.ResultRow {
	out := make([]types.ResultRow, 0, len(rows))
	for _, row := range rows {
		m := types.ResultRow{}
		for k, v := range row.Fields {
			m[k] = v.AsInterface()
		}
		out = append(out, m)
	}
	return out
}

func (r *RemoteDB) enforceLimits(rows []types.ResultRow) []types.ResultRow {
	limited := rows
	if r.maxRows > 0 && len(limited) > r.maxRows {
		limited = limited[:r.maxRows]
	}
	if r.maxResultBytes > 0 {
		var total int64
		out := make([]types.ResultRow, 0, len(limited))
		for _, row := range limited {
			b, _ := json.Marshal(row)
			if total+int64(len(b)) > r.maxResultBytes {
				break
			}
			total += int64(len(b))
			out = append(out, row)
		}
		limited = out
	}
	return limited
}

func (r *RemoteDB) decodeResult(raw interface{}, dest interface{}) error {
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: as400DateHook(),
		Result:     dest,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}
	return decoder.Decode(raw)
}
