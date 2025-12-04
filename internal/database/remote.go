package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
)

// RemoteDB satisfies the DB interface by proxying calls to a remote agent over the bridge.
type RemoteDB struct {
	tenantID string
	registry bridge.AgentRegistry
	Logger   *slog.Logger
}

func NewRemoteDB(tenantID string, registry bridge.AgentRegistry, logger *slog.Logger) *RemoteDB {
	return &RemoteDB{
		tenantID: tenantID,
		registry: registry,
		Logger:   logger,
	}
}

// Lifecycle methods are no-ops for remote connections.
func (r *RemoteDB) StartJDBCRunner() error { return nil }
func (r *RemoteDB) StopJDBCRunner() error  { return nil }

func (r *RemoteDB) Connect(ctx context.Context) error {
	_, err := r.sendJob(ctx, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_CONNECT,
	})
	return err
}

func (r *RemoteDB) Disconnect(ctx context.Context) error {
	_, err := r.sendJob(ctx, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_DISCONNECT,
	})
	return err
}

func (r *RemoteDB) PingService(ctx context.Context) error {
	_, err := r.sendJob(ctx, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_PING_SERVICE,
	})
	return err
}

func (r *RemoteDB) PingDatabase(ctx context.Context) error {
	_, err := r.sendJob(ctx, &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_PING_DATABASE,
	})
	return err
}

func (r *RemoteDB) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	req := &bridgeproto.JobRequest{
		JobId: uuid.New().String(),
		Kind:  bridgeproto.JobKind_JOB_KIND_QUERY,
		Sql:   query,
	}
	res, err := r.sendJob(ctx, req)
	if err != nil {
		return nil, err
	}
	return protoRowsToResultRows(res.Rows), nil
}

func (r *RemoteDB) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	rows, err := r.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows returned for query %q", query)
	}
	return rows[0], nil
}

func (r *RemoteDB) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	rows, err := r.Query(ctx, query, args...)
	if err != nil {
		return err
	}
	return r.decodeResult(rows, dest)
}

func (r *RemoteDB) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	row, err := r.QueryRow(ctx, query, args...)
	if err != nil {
		return err
	}
	return r.decodeResult(row, dest)
}

func (r *RemoteDB) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	req := &bridgeproto.JobRequest{
		JobId:      uuid.New().String(),
		Kind:       bridgeproto.JobKind_JOB_KIND_QUERY,
		Sql:        query,
		IncludeSrc: true,
	}
	res, err := r.sendJob(ctx, req)
	if err != nil {
		return nil, err
	}
	return protoRowsToResultRows(res.Rows), nil
}

// --- helpers ---

func (r *RemoteDB) sendJob(ctx context.Context, req *bridgeproto.JobRequest) (*bridgeproto.JobResult, error) {
	agent, err := r.registry.Pick(ctx, r.tenantID)
	if err != nil {
		return nil, err
	}
	res, err := agent.SendJob(ctx, req)
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
