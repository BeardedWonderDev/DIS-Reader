package database

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"time"

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
	Disconnect(ctx context.Context) error
	// PingService checks that the Java TCP service is running.
	PingService(ctx context.Context) error
	// PingDatabase checks that the Java service is connected to the AS/400 database.
	PingDatabase(ctx context.Context) error
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error)
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error)
}

// MultiTenantDB mirrors DB but requires the caller to supply a tenant string.
// This is used by remote mode, while domain code continues to depend on DB.
type MultiTenantDB interface {
	StartJDBCRunner() error
	StopJDBCRunner() error
	Connect(ctx context.Context, tenant string) error
	Disconnect(ctx context.Context, tenant string) error
	PingService(ctx context.Context, tenant string) error
	PingDatabase(ctx context.Context, tenant string) error
	Get(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error
	Query(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error)
	QueryRow(ctx context.Context, query string, tenant string, args ...interface{}) (types.ResultRow, error)
	Select(ctx context.Context, dest interface{}, query string, tenant string, args ...interface{}) error
	QueryWithSource(ctx context.Context, query string, tenant string, args ...interface{}) ([]types.ResultRow, error)
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
	ds.Logger.Debug("DB Connect", slog.String("request_id", requestID))
	return ds.JDBCRunner.Connect(ctx)
}

// Disconnect sends a disconnect command to close the connection pool.
func (ds *IBMi400) Disconnect(ctx context.Context) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Disconnect", slog.String("request_id", requestID))
	return ds.JDBCRunner.Disconnect(ctx)
}

// PingService checks the Java service process is alive and responding to commands.
func (ds *IBMi400) PingService(ctx context.Context) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB PingService", slog.String("request_id", requestID))
	return ds.JDBCRunner.PingService(ctx)
}

// PingDatabase performs a health check ensuring the Java service is connected to the database.
func (ds *IBMi400) PingDatabase(ctx context.Context) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB PingDatabase", slog.String("request_id", requestID))
	return ds.JDBCRunner.PingDatabase(ctx)
}

// Query executes a SQL query that may return multiple rows.
func (ds *IBMi400) Query(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Query", slog.String("request_id", requestID), slog.String("query", query))
	return ds.JDBCRunner.Query(ctx, query)
}

// QueryRow executes a SQL query expected to return exactly one row.
func (ds *IBMi400) QueryRow(ctx context.Context, query string, args ...interface{}) (types.ResultRow, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB QueryRow", slog.String("request_id", requestID), slog.String("query", query))
	rows, err := ds.JDBCRunner.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows returned for query %q", query)
	}
	return rows[0], nil
}

// decodeResult is a helper to decode raw data into dest using mapstructure with AS/400 date hook.
func (ds *IBMi400) decodeResult(ctx context.Context, requestID string, raw interface{}, dest interface{}) error {
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

// Get executes a single-row query and decodes the result into dest.
func (ds *IBMi400) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Get", slog.String("request_id", requestID), slog.String("query", query))
	row, err := ds.QueryRow(ctx, query, args...)
	if err != nil {
		return err
	}
	return ds.decodeResult(ctx, requestID, row, dest)
}

// Select executes the query and decodes all rows into dest.
func (ds *IBMi400) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB Select", slog.String("request_id", requestID), slog.String("query", query))
	rows, err := ds.JDBCRunner.Query(ctx, query)
	if err != nil {
		return err
	}
	return ds.decodeResult(ctx, requestID, rows, dest)
}

// QueryWithSource executes a SQL query with src_table included on each row.
func (ds *IBMi400) QueryWithSource(ctx context.Context, query string, args ...interface{}) ([]types.ResultRow, error) {
	requestID := uuid.New().String()
	ctx = context.WithValue(ctx, requestIDKey, requestID)
	ds.Logger.Debug("DB QueryWithSource", slog.String("request_id", requestID), slog.String("query", query))
	payload := fmt.Sprintf(`{"cmd":"query","requestId":%q,"sql":%q,"includeSrc":"true"}`, requestID, query)
	return ds.JDBCRunner.rawQuery(ctx, payload)
}

// as400DateHook returns a DecodeHookFunc for mapstructure to decode AS/400 date formats (YYYYMMDD or MMDDYY numeric).
func as400DateHook() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() == reflect.Float64 && (to == reflect.TypeOf(time.Time{}) || to == reflect.TypeOf(&time.Time{})) {
			num := int64(data.(float64))
			if num <= 0 {
				if to == reflect.TypeOf(&time.Time{}) {
					return nil, nil
				}
				return time.Time{}, nil
			}

			str := fmt.Sprintf("%d", num)
			// Unix epoch values (seconds or milliseconds) arrive as long integers from
			// the JDBC runner. Handle them before falling back to legacy MMDDYY/YYMMDD
			// numeric date encodings.
			if len(str) > 10 {
				parsed := time.UnixMilli(num).UTC()
				if to == reflect.TypeOf(&time.Time{}) {
					return &parsed, nil
				}
				return parsed, nil
			}
			if len(str) == 10 {
				parsed := time.Unix(num, 0).UTC()
				if to == reflect.TypeOf(&time.Time{}) {
					return &parsed, nil
				}
				return parsed, nil
			}
			var dateStr string

			switch len(str) {
			case 8:
				// Assume YYYYMMDD
				year := str[0:4]
				month := str[4:6]
				day := str[6:8]
				dateStr = fmt.Sprintf("%s-%s-%s", year, month, day)
			case 6:
				// Assume MMDDYY
				mm := str[0:2]
				dd := str[2:4]
				yy := str[4:6]
				dateStr = fmt.Sprintf("20%s-%s-%s", yy, mm, dd)
			case 5:
				// Assume MDDYY
				mm := "0" + str[0:1]
				dd := str[1:3]
				yy := str[3:5]
				dateStr = fmt.Sprintf("20%s-%s-%s", yy, mm, dd)
			default:
				fmt.Printf("Unknown AS/400 date format: input=%d\n", num)
				if to == reflect.TypeOf(&time.Time{}) {
					return nil, nil
				}
				return time.Time{}, nil
			}

			parsed, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				fmt.Printf("Failed to parse AS/400 date: input=%d parsed=%s err=%v\n", num, dateStr, err)
				if to == reflect.TypeOf(&time.Time{}) {
					return nil, nil
				}
				return time.Time{}, nil
			}
			if to == reflect.TypeOf(&time.Time{}) {
				return &parsed, nil
			}
			return parsed, nil
		}
		return data, nil
	}
}
