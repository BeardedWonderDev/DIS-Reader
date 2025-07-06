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
	"github.com/BeardedWonderDev/DIS-Reader/internal/unit"
	"github.com/BeardedWonderDev/DIS-Reader/types"
)

//go:embed dis-runner-0.0.4.jar
var runnerJar []byte

// DISReaderService manages the lifecycle of the Java JDBC runner and exposes query APIs.
type DISReaderService struct {
	config   *types.Config
	db       database.DB
	logger   *slog.Logger
	logLevel *slog.LevelVar
	tempDir  string

	unitService types.UnitService
}

// NewDISReaderService creates a DISReaderService, writes embedded JAR,
// starts the JDBC runner, verifies connectivity, and returns an error on failure.
// Note: This only starts the Java process; AS/400 connectivity is established via Connect.
func NewDISReaderService(config *types.Config, logger *slog.Logger) (*DISReaderService, error) {
	var logLevel slog.LevelVar
	logLevel.Set(config.LogLevel)

	if logger == nil {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: &logLevel,
		})
		logger = slog.New(handler)
	}

	tmp, err := os.MkdirTemp("", "disreader-jdbc-*")
	if err != nil {
		logger.Error("Failed to create temp directory", "error", err)
		return nil, err
	}

	jarPath := tmp + string(os.PathSeparator) + "dis-runner-0.0.4.jar"
	if err := os.WriteFile(jarPath, runnerJar, 0644); err != nil {
		logger.Error("Failed to write runner jar", "error", err)
		os.RemoveAll(tmp)
		return nil, err
	}
	config.DIS.JDBCConfig.JarPath = jarPath

	// Initialize and start the JDBC runner
	db := database.NewIBMi400(config.DIS, logger)
	if err := db.StartJDBCRunner(); err != nil {
		logger.Error("Failed to start JDBC runner", "error", err)
		os.RemoveAll(tmp)
		return nil, err
	}

	s := &DISReaderService{
		config:      config,
		db:          db,
		logger:      logger,
		logLevel:    &logLevel,
		tempDir:     tmp,
		unitService: unit.NewUnitService(db),
	}

	// Verify the runner is responsive
	if err := s.db.PingService(context.Background()); err != nil {
		s.logger.Error("JDBC runner ping failed after start", "error", err)
		s.db.StopJDBCRunner()
		os.RemoveAll(tmp)
		return nil, err
	}

	return s, nil
}

// GetConfig returns the underlying configuration.
func (s *DISReaderService) GetConfig() *types.DISConfig {
	return s.config.DIS
}

// GetLogger returns the underlying logger.
func (s *DISReaderService) GetLogger() *slog.Logger {
	return s.logger
}

func (s *DISReaderService) SetLogger(logger *slog.Logger) {
	s.logger = logger
	s.logger.Info("DIS Logger Attached")
}

// TestConnection verifies both the Java service is running and the database is connected.
func (s *DISReaderService) TestConnection(ctx context.Context) error {
	// Check Java service responsiveness
	if err := s.db.PingService(ctx); err != nil {
		return fmt.Errorf("service ping failed: %w", err)
	}
	// Check AS/400 database connectivity
	if err := s.db.PingDatabase(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

// Connect forwards the connect command to the Java server to establish AS/400 connectivity.
func (s *DISReaderService) Connect(ctx context.Context) error {
	s.logger.Debug("DISReaderService Calling JDBC Connect")
	return s.db.Connect(ctx)
}

// Disconnect forwards the disconnect command to the Java server to close AS/400 connectivity.
func (s *DISReaderService) Disconnect(ctx context.Context) error {
	s.logger.Debug("DISReaderService Calling JDBC Disconnect")
	return s.db.Disconnect(ctx)
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
