package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// DB defines the operations for interacting with the AS/400 database service.
// It includes lifecycle control, health checks, and standard CRUD-style querying methods.
type DB interface {
	StartJDBCRunner() error
	StopJDBCRunner() error
	Connect(ctx context.Context) error
	Ping(ctx context.Context) error
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error)
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error)
}

// IBMi400 is a Go wrapper around JDBCRunnerService, exposing high-level database methods.
type IBMi400 struct {
	JDBCRunner *JDBCRunnerService
	Config     *types.DISConfig
	Logger     *slog.Logger
}

// NewIBMi400 constructs an IBMi400 given configuration and logger.
func NewIBMi400(config *types.DISConfig, logger *slog.Logger) *IBMi400 {
	return &IBMi400{
		JDBCRunner: NewJDBCRunnerService(config, logger),
		Config:     config,
		Logger:     logger,
	}
}

// StartJDBCRunner launches the Java JDBC runner process and waits until it's ready.
func (ds *IBMi400) StartJDBCRunner() error {
	return ds.JDBCRunner.Start()
}

// StopJDBCRunner terminates the Java JDBC runner process.
func (ds *IBMi400) StopJDBCRunner() error {
	ds.Logger.Debug("Stopping JDBC runner")
	ds.JDBCRunner.Shutdown()
	return nil
}

// Connect establishes a new TCP connection to the JDBC runner, sending a connect command and verifying the response.
func (ds *IBMi400) Connect(ctx context.Context) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Connect", "requestId", requestID)
	return ds.JDBCRunner.Connect(ctx)
}

// Ping performs a health check by sending a ping command to the JDBC runner.
func (ds *IBMi400) Ping(ctx context.Context) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Ping", "requestId", requestID)
	return ds.JDBCRunner.Ping(ctx)
}

// Query executes a SQL query that may return multiple rows.
func (ds *IBMi400) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Query", "requestId", requestID, "query", query)
	return ds.JDBCRunner.Query(ctx, query)
}

// QueryRow executes a SQL query expected to return exactly one row.
func (ds *IBMi400) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB QueryRow", "requestId", requestID, "query", query)
	rows, err := ds.JDBCRunner.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows returned for query %q", query)
	}
	return rows[0], nil
}

// Get executes a single-row query and decodes the result into dest.
func (ds *IBMi400) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Get", "requestId", requestID, "query", query)
	row, err := ds.QueryRow(ctx, query, args...)
	if err != nil {
		return err
	}
	return mapstructure.Decode(row, dest)
}

// Select executes the query and decodes all rows into dest.
func (ds *IBMi400) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Select", "requestId", requestID, "query", query)
	rows, err := ds.JDBCRunner.Query(ctx, query)
	if err != nil {
		return err
	}
	return mapstructure.Decode(rows, dest)
}

// QueryWithSource executes a SQL query with src_table included on each row.
func (ds *IBMi400) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB QueryWithSource", "requestId", requestID, "query", query)
	payload := fmt.Sprintf(`{"cmd":"query","requestId":%q,"sql":%q,"includeSrc":"true"}`, requestID, query)
	return ds.JDBCRunner.rawQuery(ctx, payload)
}
