package internal

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/invoices"
	"github.com/BeardedWonderDev/DIS-Reader/internal/parts"
	"github.com/BeardedWonderDev/DIS-Reader/internal/unit"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/dusted-go/logging/prettylog"
	"google.golang.org/grpc"
	"net/http"
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

	bridgeRegistry bridge.AgentRegistry
	bridgeServer   *bridge.Server

	unitService    types.UnitService
	invoiceService types.InvoiceService
	partService    types.PartService
}

// NewDISReaderService creates a DISReaderService, writes embedded JAR,
// starts the JDBC runner, verifies connectivity, and returns an error on failure.
// Note: This only starts the Java process; AS/400 connectivity is established via Connect.
func NewDISReaderService(config *types.DISConfig, logger *slog.Logger) (*DISReaderService, error) {
	return NewDISReaderServiceWithAuth(config, logger, nil)
}

// NewDISReaderServiceWithAuth allows callers to supply a custom AgentAuthenticator.
// When auth is nil, a static authenticator is derived from bridge config.
func NewDISReaderServiceWithAuth(config *types.DISConfig, logger *slog.Logger, auth bridge.AgentAuthenticator) (*DISReaderService, error) {
	var logLevel slog.LevelVar
	logLevel.Set(config.LogLevel)

	if logger == nil {
		logger = slog.New(prettylog.NewHandler(&slog.HandlerOptions{
			Level: &logLevel,
		}))
	}

	if config.Bridge != nil && strings.EqualFold(config.Bridge.Mode, "remote") {
		registry := bridge.NewInMemoryRegistry()
		authenticator := buildAuthenticator(config.Bridge, auth)
		db := database.NewRemoteDB(config.Bridge.TenantID, registry, logger)

		s := &DISReaderService{
			config:         config,
			db:             db,
			logger:         logger,
			logLevel:       &logLevel,
			bridgeRegistry: registry,
			bridgeServer:   bridge.NewServer(authenticator, registry, logger),
			unitService:    unit.NewUnitService(db),
			invoiceService: invoices.NewInvoiceService(db),
			partService:    parts.NewService(db),
		}
		return s, nil
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
	if s.tempDir != "" {
		if remErr := os.RemoveAll(s.tempDir); remErr != nil {
			types.LogError(s.logger, "Failed to remove temp directory", remErr)
		}
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

// BridgeServer exposes the bridge server for hosting gRPC handlers when running in remote mode.
// Returns nil when bridge mode is not enabled.
func (s *DISReaderService) BridgeServer() proto.AgentServiceServer {
	return s.bridgeServer
}

// RegisterHealth registers /healthz and /metrics on the supplied mux when bridge mode is enabled.
func (s *DISReaderService) RegisterHealth(mux *http.ServeMux) {
	if s.bridgeRegistry == nil {
		return
	}
	mux.Handle("/healthz", bridge.HealthHandler(s.bridgeRegistry))
	mux.Handle("/metrics", bridge.MetricsHandler(s.bridgeRegistry))
}

// RegisterBridge registers the bridge gRPC handler on the provided server.
// No-op when bridge mode is disabled.
func (s *DISReaderService) RegisterBridge(server *grpc.Server) {
	if s.bridgeServer == nil {
		return
	}
	proto.RegisterAgentServiceServer(server, s.bridgeServer)
}

func buildAuthenticator(cfg *types.BridgeConfig, override bridge.AgentAuthenticator) bridge.AgentAuthenticator {
	if override != nil {
		return override
	}
	if cfg == nil {
		return &bridge.StaticAuthenticator{Secrets: map[string]bridge.StaticAgentSecret{}}
	}
	secrets := map[string]bridge.StaticAgentSecret{}
	for _, agent := range cfg.Allowed {
		secrets[agent.ClientID] = bridge.StaticAgentSecret{
			ClientSecret: agent.ClientSecret,
			TenantID:     agent.TenantID,
			AgentID:      agent.AgentID,
		}
	}
	if cfg.ClientID != "" {
		secrets[cfg.ClientID] = bridge.StaticAgentSecret{
			ClientSecret: cfg.ClientSecret,
			TenantID:     cfg.TenantID,
			AgentID:      cfg.ClientID,
		}
	}
	return &bridge.StaticAuthenticator{Secrets: secrets}
}
