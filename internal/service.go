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
	"github.com/BeardedWonderDev/DIS-Reader/internal/invoices"
	"github.com/BeardedWonderDev/DIS-Reader/internal/parts"
	"github.com/BeardedWonderDev/DIS-Reader/internal/unit"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/dusted-go/logging/prettylog"
)

//go:embed dis-runner-0.1.1.jar
var runnerJar []byte

// DISReaderService manages the lifecycle of the Java JDBC runner and exposes query APIs.
type DISReaderService struct {
	config   *types.DISConfig
	db       database.DB
	logger   *slog.Logger
	logLevel *slog.LevelVar
	tempDir  string

	unitService    types.UnitService
	invoiceService types.InvoiceService
	partService    types.PartService
}

// NewDISReaderService creates a DISReaderService, writes embedded JAR,
// starts the JDBC runner, verifies connectivity, and returns an error on failure.
// Note: This only starts the Java process; AS/400 connectivity is established via Connect.
func NewDISReaderService(config *types.DISConfig, logger *slog.Logger) (*DISReaderService, error) {
	var logLevel slog.LevelVar
	logLevel.Set(config.LogLevel)

	if logger == nil {
		logger = slog.New(prettylog.NewHandler(&slog.HandlerOptions{
			Level: &logLevel,
		}))
	}

	tmp, err := os.MkdirTemp("", "disreader-jdbc-*")
	if err != nil {
		types.LogError(logger, "Failed to create temp directory", err)
		return nil, err
	}

	jarPath := tmp + string(os.PathSeparator) + "dis-runner-0.1.1.jar"
	if err := os.WriteFile(jarPath, runnerJar, 0644); err != nil {
		types.LogError(logger, "Failed to write runner jar", err)
		os.RemoveAll(tmp)
		return nil, err
	}
	config.JDBCConfig.JarPath = jarPath

	// Initialize and start the JDBC runner
	db := database.NewIBMi400(config, logger)
	if err := db.StartJDBCRunner(); err != nil {
		types.LogError(logger, "Failed to start JDBC runner", err)
		os.RemoveAll(tmp)
		return nil, err
	}

	s := &DISReaderService{
		config:         config,
		db:             db,
		logger:         logger,
		logLevel:       &logLevel,
		tempDir:        tmp,
		unitService:    unit.NewUnitService(db),
		invoiceService: invoices.NewInvoiceService(db),
		partService:    parts.NewService(db),
	}

	// Verify the runner is responsive
	if err := s.db.PingService(context.Background()); err != nil {
		types.LogError(s.logger, "JDBC runner ping failed after start", err)
		s.db.StopJDBCRunner()
		os.RemoveAll(tmp)
		return nil, err
	}

	if config.Host != "" && config.User != "" && config.Password != "" {
		ctx := context.TODO()
		if err := s.Connect(ctx); err != nil {
			types.LogError(logger, "DIS Connection Failed", err)
			return nil, err
		}

		if err := s.db.PingDatabase(ctx); err != nil {
			types.LogError(logger, "DIS Connection Failed", err)
			return nil, err
		}
	}

	return s, nil
}

// GetConfig returns the underlying configuration.
func (s *DISReaderService) GetConfig() *types.DISConfig {
	return s.config
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
		types.LogError(s.logger, "Failed to remove temp directory", remErr)
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

func (s *DISReaderService) UnitService() types.UnitService {
	return s.unitService
}

func (s *DISReaderService) InvoiceService() types.InvoiceService {
	return s.invoiceService
}

func (s *DISReaderService) PartService() types.PartService {
	return s.partService
}
