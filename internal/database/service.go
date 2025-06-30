package database

import (
	"bufio"
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
	className = "JDBCRunner"
)

type JDBCRunnerService struct {
	config  *types.Config
	javaCmd *exec.Cmd
	conn    net.Conn
	Logger  *slog.Logger
}

func NewJDBCRunnerService(config *types.Config, logger *slog.Logger) *JDBCRunnerService {
	return &JDBCRunnerService{
		config: config,
		Logger: logger,
	}
}

func (j *JDBCRunnerService) Start() error {
	if j.javaCmd != nil && j.javaCmd.ProcessState == nil {
		j.Logger.Debug("Java process already running", "pid", j.javaCmd.Process.Pid)
		return nil // already running
	}
	cmd := exec.Command(j.config.JavaPath, "-cp", fmt.Sprintf("%s:.", j.config.JarPath), className, j.config.Host, j.config.User, j.config.Password, j.config.JDBCPort)
	stdoutPipe, _ := cmd.StdoutPipe()
	stderrPipe, _ := cmd.StderrPipe()

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			j.Logger.Debug(fmt.Sprintf("Database Service: %s", scanner.Text()))
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			j.Logger.Warn(fmt.Sprintf("Database Service: %s", scanner.Text()))
		}
	}()

	if err := cmd.Start(); err != nil {
		j.Logger.Error("Failed to start java process", "error", err)
		return fmt.Errorf("failed to start java process: %w", err)
	}
	j.javaCmd = cmd
	j.Logger.Info("Started java process", "pid", cmd.Process.Pid)

	// Monitor java process in background
	go func() {
		err := cmd.Wait()
		if err != nil {
			j.Logger.Error("Java process exited with error", "error", err)
		} else {
			j.Logger.Info("Java process exited normally")
		}
		j.javaCmd = nil
	}()

	// Wait for Java server port to be open
	retries := 10
	for i := 0; i < retries; i++ {
		conn, err := net.DialTimeout("tcp", "localhost:"+j.config.JDBCPort, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			j.Logger.Info("Java server port is open", "port", j.config.JDBCPort)
			return nil
		}
		j.Logger.Debug("Waiting for java server port to open...", "port", j.config.JDBCPort, "attempt", i+1, "max", retries)
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("java server port %s did not open after %d retries", j.config.JDBCPort, retries)
}

func (j *JDBCRunnerService) Shutdown() {
	if j.conn != nil {
		j.conn.Close()
		j.conn = nil
	}
	if j.javaCmd != nil && j.javaCmd.Process != nil {
		if err := j.javaCmd.Process.Kill(); err != nil {
			j.Logger.Error("Failed to kill Java process", "error", err)
		} else {
			j.Logger.Info("Java process terminated")
		}
		j.javaCmd = nil
	}
}

func (j *JDBCRunnerService) Connect() error {
	if j.conn == nil {
		j.Logger.Debug("Attempting to connect to JDBC socket", "port", j.config.JDBCPort)
		conn, err := net.Dial("tcp", "localhost:"+j.config.JDBCPort)
		if err != nil {
			j.Logger.Error("Failed to connect to JDBC socket", "error", err)
			return fmt.Errorf("failed to connect to JDBC runner: %w", err)
		}
		j.conn = conn

		msg := `{"cmd":"connect"}`
		if _, err := j.conn.Write([]byte(msg + "\n")); err != nil {
			j.Logger.Error("Failed to send connect command", "error", err)
			return fmt.Errorf("failed to send connect command: %w", err)
		}

		scanner := bufio.NewScanner(j.conn)
		if scanner.Scan() {
			var res types.JDBCResult
			if err := json.Unmarshal(scanner.Bytes(), &res); err != nil {
				j.Logger.Error("Failed to parse response", "error", err)
				return fmt.Errorf("failed to parse response: %w", err)
			}
			if res.Status != "ok" {
				err := fmt.Errorf("connect error: %s", res.Message)
				j.Logger.Error("Connect error", "message", res.Message)
				return err
			}
		}
		j.Logger.Info("Successfully connected to JDBC socket", "port", j.config.JDBCPort)
	}
	return nil
}

func (j *JDBCRunnerService) Query(sql string) ([]types.ResultRow, error) {
	j.Logger.Debug("Sending query to JDBC socket", "sql", sql)
	if j.conn == nil {
		return nil, fmt.Errorf("not connected to JDBC runner")
	}

	queryCmd := fmt.Sprintf(`{"cmd":"query","sql":%q}`, sql)
	if _, err := j.conn.Write([]byte(queryCmd + "\n")); err != nil {
		j.Logger.Error("Query failed", "error", err)
		return nil, fmt.Errorf("failed to send query: %w", err)
	}

	var results []types.ResultRow
	scanner := bufio.NewScanner(j.conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		var result types.ResultRow
		if err := json.Unmarshal(line, &result); err != nil {
			j.Logger.Error("Query failed", "error", err)
			return nil, fmt.Errorf("failed to parse row: %w", err)
		}
		if status, exists := result["status"]; exists {
			switch status {
			case "ok":
				// Query completed successfully, continue reading until "done"
				continue
			case "error":
				err := fmt.Errorf("query error: %s", result["message"])
				j.Logger.Error("Query failed", "error", err)
				return nil, err
			case "done":
				// End of query results
				break
			default:
				results = append(results, result)
			}
			if status == "done" {
				break
			}
		} else {
			results = append(results, result)
		}
	}
	j.Logger.Info("Query completed", "rows", len(results))
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
