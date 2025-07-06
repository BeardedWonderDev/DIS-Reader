package database

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os/exec"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/mitchellh/mapstructure"
)

const (
	serverStartRetries  = 10
	serverStartInterval = 500 * time.Millisecond
)

// Context key for request ID propagation
type contextKey string

const requestIDKey contextKey = "requestId"

type JDBCRunnerService struct {
	config  *types.DISConfig
	javaCmd *exec.Cmd
	Logger  *slog.Logger
}

func NewJDBCRunnerService(config *types.DISConfig, logger *slog.Logger) *JDBCRunnerService {
	return &JDBCRunnerService{
		config: config,
		Logger: logger,
	}
}

func logStructuredLine(logger *slog.Logger, line string, level slog.Level) {
	var msg map[string]any
	if err := json.Unmarshal([]byte(line), &msg); err != nil {
		if level == slog.LevelError {
			types.LogError(logger, "Database service stderr", err, slog.String("output", line))
		} else {
			logger.Info("Database service stdout", slog.String("output", line))
		}
		return
	}

	attrs := []any{}
	if status, ok := msg["status"]; ok {
		attrs = append(attrs, slog.String("status", fmt.Sprint(status)))
	}
	if message, ok := msg["message"]; ok {
		attrs = append(attrs, slog.String("message", fmt.Sprint(message)))
	}
	if event, ok := msg["event"]; ok {
		attrs = append(attrs, slog.String("event", fmt.Sprint(event)))
	}
	if requestID, ok := msg["requestId"]; ok {
		attrs = append(attrs, slog.String("request_id", fmt.Sprint(requestID)))
	}

	switch level {
	case slog.LevelInfo:
		logger.Info("Database service", attrs...)
	case slog.LevelError:
		logger.Error("Database service", attrs...)
	default:
		logger.Debug("Database service", attrs...)
	}
}

func (j *JDBCRunnerService) Start() error {
	if j.javaCmd != nil && j.javaCmd.ProcessState == nil {
		j.Logger.Debug("Java process already running", "pid", j.javaCmd.Process.Pid)
		return nil
	}
	cmd := exec.Command(
		j.config.JDBCConfig.JavaPath,
		"-jar", j.config.JDBCConfig.JarPath,
		j.config.JDBCConfig.JDBCPort,
	)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			logStructuredLine(j.Logger, scanner.Text(), slog.LevelInfo)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			logStructuredLine(j.Logger, scanner.Text(), slog.LevelError)
		}
	}()
	if err := cmd.Start(); err != nil {
		types.LogError(j.Logger, "Failed to start java process", err)
		return fmt.Errorf("failed to start java process: %w", err)
	}
	j.javaCmd = cmd
	j.Logger.Info("Started java process", slog.Int("pid", cmd.Process.Pid))

	go func() {
		err := cmd.Wait()
		if err != nil {
			types.LogError(j.Logger, "Java process exited with error", err)
		} else {
			j.Logger.Info("Java process exited normally")
		}
		j.javaCmd = nil
	}()

	for i := 0; i < serverStartRetries; i++ {
		connTest, err := net.DialTimeout("tcp", "localhost:"+j.config.JDBCConfig.JDBCPort, serverStartInterval)
		if err == nil {
			connTest.Close()
			j.Logger.Info("Java server port is open", slog.String("port", j.config.JDBCConfig.JDBCPort))
			return nil
		}
		j.Logger.Debug("Waiting for java server port", slog.String("port", j.config.JDBCConfig.JDBCPort), slog.Int("attempt", i+1), slog.Int("max_attempts", serverStartRetries))
		time.Sleep(serverStartInterval)
	}
	return fmt.Errorf("java server port %s did not open after %d retries", j.config.JDBCConfig.JDBCPort, serverStartRetries)
}

func (j *JDBCRunnerService) Shutdown() {
	if j.javaCmd != nil && j.javaCmd.Process != nil {
		if err := j.javaCmd.Process.Kill(); err != nil {
			types.LogError(j.Logger, "Failed to kill Java process", err)
		} else {
			j.Logger.Info("Java process terminated")
		}
		j.javaCmd = nil
	}
}

func (j *JDBCRunnerService) Connect(ctx context.Context) error {
	requestID, _ := ctx.Value(requestIDKey).(string)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+j.config.JDBCConfig.JDBCPort)
	if err != nil {
		return fmt.Errorf("failed to dial java server: %w", err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}
	cmd := fmt.Sprintf(
		`{"cmd":"connect","requestId":%q,"host":%q,"user":%q,"pass":%q}`,
		requestID, j.config.Host, j.config.User, j.config.Password,
	)
	if _, err := conn.Write([]byte(cmd + "\n")); err != nil {
		return fmt.Errorf("failed to send connect: %w", err)
	}
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var res types.ResultRow
		if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
			return fmt.Errorf("invalid connect response: %w", err)
		}
		if status, ok := res["status"]; !ok || status != "ok" {
			return fmt.Errorf("connect error: %v", res["message"])
		}
		return nil
	}
	return fmt.Errorf("no connect response")
}

// Disconnect sends a disconnect command to the Java server to close the pool.
func (j *JDBCRunnerService) Disconnect(ctx context.Context) error {
	requestID, _ := ctx.Value(requestIDKey).(string)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+j.config.JDBCConfig.JDBCPort)
	if err != nil {
		return fmt.Errorf("failed to dial java server for disconnect: %w", err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}
	cmd := fmt.Sprintf(`{"cmd":"disconnect","requestId":%q}`, requestID)
	if _, err := conn.Write([]byte(cmd + "\n")); err != nil {
		return fmt.Errorf("failed to send disconnect: %w", err)
	}
	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		var res types.ResultRow
		if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
			return fmt.Errorf("invalid disconnect response: %w", err)
		}
		if status, ok := res["status"]; !ok || status != "ok" {
			return fmt.Errorf("disconnect error: %v", res["message"])
		}
		return nil
	}
	return fmt.Errorf("no disconnect response")
}

func (j *JDBCRunnerService) Query(ctx context.Context, sql string) ([]types.ResultRow, error) {
	requestID, _ := ctx.Value(requestIDKey).(string)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+j.config.JDBCConfig.JDBCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to dial java server: %w", err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}
	j.Logger.Debug("Sending query", slog.String("request_id", requestID), slog.String("sql", sql))
	payload := fmt.Sprintf(`{"cmd":"query","requestId":%q,"sql":%q}`, requestID, sql)
	return j.rawQuery(ctx, payload)
}

// PingService checks that the Java service process is alive and responding to commands.
func (j *JDBCRunnerService) PingService(ctx context.Context) error {
	requestID, _ := ctx.Value(requestIDKey).(string)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+j.config.JDBCConfig.JDBCPort)
	if err != nil {
		return fmt.Errorf("java service not running: %w", err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}
	// Send ping command
	payload := fmt.Sprintf(`{"cmd":"ping","requestId":%q}`, requestID)
	if _, err := conn.Write([]byte(payload + "\n")); err != nil {
		return fmt.Errorf("failed to send ping service command: %w", err)
	}
	// Read response
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return fmt.Errorf("no response from JDBC server to ping service")
	}
	var res types.ResultRow
	if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
		return fmt.Errorf("invalid ping service response: %w", err)
	}
	status, hasStatus := res["status"]
	message := res["message"]
	if !hasStatus {
		return fmt.Errorf("ping service response missing status field")
	}
	if status == "ok" || message == "Not connected to DB" {
		return nil
	}
	return fmt.Errorf("ping service error: %v", message)
}

// PingDatabase checks that the Java service is connected to the database.
func (j *JDBCRunnerService) PingDatabase(ctx context.Context) error {
	requestID, _ := ctx.Value(requestIDKey).(string)
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+j.config.JDBCConfig.JDBCPort)
	if err != nil {
		return fmt.Errorf("failed to dial java server: %w", err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}
	// Send ping command
	payload := fmt.Sprintf(`{"cmd":"ping","requestId":%q}`, requestID)
	if _, err := conn.Write([]byte(payload + "\n")); err != nil {
		return fmt.Errorf("failed to send ping: %w", err)
	}
	// Read response
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return fmt.Errorf("no response from JDBC server to ping")
	}
	var res types.ResultRow
	if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
		return fmt.Errorf("invalid ping response: %w", err)
	}
	status, hasStatus := res["status"]
	message := res["message"]
	if !hasStatus {
		return fmt.Errorf("ping response missing status field")
	}
	if status == "ok" {
		return nil
	}
	// Server responded with error status
	return fmt.Errorf("ping database error: %v", message)
}

// rawQuery sends a pre-built JSON payload over TCP and returns parsed ResultRows.
func (j *JDBCRunnerService) rawQuery(ctx context.Context, payload string) ([]types.ResultRow, error) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", "localhost:"+j.config.JDBCConfig.JDBCPort)
	if err != nil {
		return nil, fmt.Errorf("failed to dial java server: %w", err)
	}
	defer conn.Close()
	if dl, ok := ctx.Deadline(); ok {
		conn.SetDeadline(dl)
	}
	if _, err := conn.Write([]byte(payload + "\n")); err != nil {
		return nil, fmt.Errorf("failed to send payload: %w", err)
	}
	var results []types.ResultRow
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		var m types.ResultRow
		if err := json.Unmarshal(scanner.Bytes(), &m); err != nil {
			return nil, fmt.Errorf("failed to parse row: %w", err)
		}
		if status, ok := m["status"]; ok {
			switch status {
			case "error":
				return nil, fmt.Errorf("query error: %s", m["message"])
			case "done":
				return results, nil
			}
		} else {
			results = append(results, m)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// DecodeRows maps a slice of types.ResultRow to a typed slice using mapstructure.
func DecodeRows[T any](rows []types.ResultRow) ([]T, error) {
	var results []T
	for i, row := range rows {
		var decoded T
		err := mapstructure.Decode(row, &decoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode row %d: %w", i, err)
		}
		results = append(results, decoded)
	}
	return results, nil
}

// As400DateExpr returns the SQL expression to convert a 6-digit AS/400 numeric date column into YYYY-MM-DD.
func As400DateExpr(col string) string {
	return "DATE(" +
		"'20'||SUBSTR(RIGHT('000000'||TRIM(CHAR(" + col + ")),6),5,2)||'-'" +
		"||SUBSTR(RIGHT('000000'||TRIM(CHAR(" + col + ")),6),1,2)||'-'" +
		"||SUBSTR(RIGHT('000000'||TRIM(CHAR(" + col + ")),6),3,2)" +
		")"
}
