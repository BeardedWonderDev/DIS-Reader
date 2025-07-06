package internal

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

//go:embed dis-runner-1.0.0.jar
var runnerJar []byte

// DISReaderService manages the lifecycle of the Java JDBC runner and exposes query APIs.
type DISReaderService struct {
	config  *types.DISConfig
	db      database.DB
	logger  *slog.Logger
	tempDir string
}

// NewDISReaderService creates a DISReaderService, writes embedded JAR,
// starts the JDBC runner, verifies connectivity, and returns an error on failure.
func NewDISReaderService(config *types.DISConfig, logger *slog.Logger) (*DISReaderService, error) {
	if logger == nil {
		logger = slog.Default()
	}

	tmp, err := os.MkdirTemp("", "disreader-jdbc-*")
	if err != nil {
		logger.Error("Failed to create temp directory", "error", err)
		return nil, err
	}

	jarPath := tmp + string(os.PathSeparator) + "dis-runner-1.0.0.jar"
	if err := os.WriteFile(jarPath, runnerJar, 0644); err != nil {
		logger.Error("Failed to write runner jar", "error", err)
		os.RemoveAll(tmp)
		return nil, err
	}
	config.JDBCConfig.JarPath = jarPath

	// Initialize and start the JDBC runner
	db := database.NewIBMi400(config, logger)
	if err := db.StartJDBCRunner(); err != nil {
		logger.Error("Failed to start JDBC runner", "error", err)
		os.RemoveAll(tmp)
		return nil, err
	}

	s := &DISReaderService{
		config:  config,
		db:      db,
		logger:  logger,
		tempDir: tmp,
	}

	// Verify the runner is responsive
	if err := s.db.Ping(context.Background()); err != nil {
		s.logger.Error("JDBC runner ping failed after start", "error", err)
		s.db.StopJDBCRunner()
		os.RemoveAll(tmp)
		return nil, err
	}

	return s, nil
}

// GetConfig returns the underlying configuration.
func (s *DISReaderService) GetConfig() *types.DISConfig {
	return s.config
}

// TestConnection runs a ping health check on the JDBC runner.
func (s *DISReaderService) TestConnection(ctx context.Context) error {
	if err := s.db.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping JDBC runner: %w", err)
	}
	return nil
}

// Shutdown stops the JDBC runner and cleans up temporary files.
func (s *DISReaderService) Shutdown() error {
	s.logger.Info("Shutting down DISReaderService")
	err := s.db.StopJDBCRunner()
	if remErr := os.RemoveAll(s.tempDir); remErr != nil {
		s.logger.Error("Failed to remove temp directory", "error", remErr)
	}
	return err
}

// AttachShutdownHook registers SIGINT/SIGTERM handlers to gracefully shutdown.
func (s *DISReaderService) AttachShutdownHook() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		s.logger.Info("Signal received, shutting down", "signal", sig)
		s.Shutdown()
		os.Exit(0)
	}()
}
