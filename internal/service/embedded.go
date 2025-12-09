package service

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/invoices"
	"github.com/BeardedWonderDev/DIS-Reader/internal/parts"
	"github.com/BeardedWonderDev/DIS-Reader/internal/runnerjar"
	"github.com/BeardedWonderDev/DIS-Reader/internal/unit"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/dusted-go/logging/prettylog"
)

// EmbeddedService implements the DISReaderService interface for embedded mode.
type EmbeddedService struct {
	config   *types.DISConfig
	db       database.DB
	logger   *slog.Logger
	logLevel *slog.LevelVar
	tempDir  string

	unitService    types.UnitService
	invoiceService types.InvoiceService
	partService    types.PartService
}

// NewEmbeddedService creates an embedded DIS reader, starts the JDBC runner, and validates connectivity if credentials are provided.
func NewEmbeddedService(config *types.DISConfig, logger *slog.Logger) (*EmbeddedService, error) {
	var logLevel slog.LevelVar
	logLevel.Set(config.LogLevel)

	if logger == nil {
		logger = slog.New(prettylog.NewHandler(&slog.HandlerOptions{Level: &logLevel}))
	}

	if config.JDBCConfig == nil {
		config.JDBCConfig = &types.JDBCConfig{}
	}

	extracted, err := runnerjar.Extract()
	if err != nil {
		types.LogError(logger, "Failed to prepare runner jar", err)
		return nil, err
	}
	config.JDBCConfig.JarPath = extracted.JarPath

	db := database.NewIBMi400(config, logger)
	if err := db.StartJDBCRunner(); err != nil {
		types.LogError(logger, "Failed to start JDBC runner", err)
		extracted.Cleanup()
		return nil, err
	}

	s := &EmbeddedService{
		config:         config,
		db:             db,
		logger:         logger,
		logLevel:       &logLevel,
		tempDir:        extracted.Dir,
		unitService:    unit.NewUnitService(db),
		invoiceService: invoices.NewInvoiceService(db),
		partService:    parts.NewService(db),
	}

	// Verify runner responsiveness
	if err := s.db.PingService(context.Background()); err != nil {
		types.LogError(s.logger, "JDBC runner ping failed after start", err)
		s.db.StopJDBCRunner()
		extracted.Cleanup()
		return nil, err
	}

	// Optional early connect if credentials provided
	if config.Host != "" && config.User != "" && config.Password != "" {
		if err := s.Connect(context.Background()); err != nil {
			types.LogError(logger, "DIS Connection Failed", err)
			return nil, err
		}
		if err := s.db.PingDatabase(context.Background()); err != nil {
			types.LogError(logger, "DIS Connection Failed", err)
			return nil, err
		}
	}

	return s, nil
}

// Interface implementations

func (s *EmbeddedService) GetConfig() *types.DISConfig { return s.config }

func (s *EmbeddedService) GetLogger() *slog.Logger { return s.logger }

func (s *EmbeddedService) SetLogger(logger *slog.Logger) {
	s.logger = logger
	s.logger.Info("DIS Logger Attached")
}

func (s *EmbeddedService) Shutdown() error {
	s.logger.Info("Shutting down DISReaderService")
	err := s.db.StopJDBCRunner()
	if s.tempDir != "" {
		if remErr := os.RemoveAll(s.tempDir); remErr != nil {
			types.LogError(s.logger, "Failed to remove temp directory", remErr)
		}
	}
	return err
}

func (s *EmbeddedService) AttachShutdownHook() {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		s.logger.Info("Signal received, shutting down", "signal", sig)
		s.Shutdown()
		os.Exit(0)
	}()
}

func (s *EmbeddedService) PingService(ctx context.Context) error {
	return s.db.PingService(ctx)
}

func (s *EmbeddedService) PingDatabase(ctx context.Context) error {
	return s.db.PingDatabase(ctx)
}

func (s *EmbeddedService) Connect(ctx context.Context) error {
	s.logger.Debug("DISReaderService Calling JDBC Connect")
	return s.db.Connect(ctx)
}

func (s *EmbeddedService) Disconnect(ctx context.Context) error {
	s.logger.Debug("DISReaderService Calling JDBC Disconnect")
	return s.db.Disconnect(ctx)
}

func (s *EmbeddedService) UnitService() types.UnitService       { return s.unitService }
func (s *EmbeddedService) InvoiceService() types.InvoiceService { return s.invoiceService }
func (s *EmbeddedService) PartService() types.PartService       { return s.partService }
