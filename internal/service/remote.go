package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	"github.com/BeardedWonderDev/DIS-Reader/internal/invoices"
	"github.com/BeardedWonderDev/DIS-Reader/internal/parts"
	"github.com/BeardedWonderDev/DIS-Reader/internal/unit"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"google.golang.org/grpc"
)

// RemoteService implements DISReaderRemote for multi-tenant remote mode.
type RemoteService struct {
	config   *types.DISConfig
	logger   *slog.Logger
	multiDB  database.MultiTenantDB
	registry bridge.AgentRegistry
	server   *bridge.Server

	unitFactory    func(tenant string) types.UnitService
	invoiceFactory func(tenant string) types.InvoiceService
	partFactory    func(tenant string) types.PartService
}

func NewRemoteService(cfg *types.DISConfig, logger *slog.Logger, multiDB database.MultiTenantDB, registry bridge.AgentRegistry, server *bridge.Server) *RemoteService {
	return &RemoteService{
		config:   cfg,
		logger:   logger,
		multiDB:  multiDB,
		registry: registry,
		server:   server,
		unitFactory: func(tenant string) types.UnitService {
			return unit.NewUnitService(database.BindTenant(multiDB, tenant))
		},
		invoiceFactory: func(tenant string) types.InvoiceService {
			return invoices.NewInvoiceService(database.BindTenant(multiDB, tenant))
		},
		partFactory: func(tenant string) types.PartService {
			return parts.NewService(database.BindTenant(multiDB, tenant))
		},
	}
}

// Bridge helpers
func (s *RemoteService) RegisterBridge(server *grpc.Server) {
	if s.server == nil || server == nil {
		return
	}
	proto.RegisterAgentServiceServer(server, s.server)
}

func (s *RemoteService) RegisterHealth(mux *http.ServeMux) {
	if s.registry == nil || mux == nil {
		return
	}
	mux.Handle("/healthz", bridge.HealthHandler(s.registry))
	mux.Handle("/metrics", bridge.MetricsHandler(s.registry))
	if s.config != nil && s.config.Bridge != nil && s.config.Bridge.PprofEnabled {
		base := s.config.Bridge.PprofPath
		if base == "" {
			base = "/debug/pprof/"
		}
		registerPprof(mux, base)
	}
}

// DISReaderRemote implementations
func (s *RemoteService) Connect(ctx context.Context, tenant string) error {
	return s.multiDB.Connect(ctx, tenant)
}
func (s *RemoteService) Disconnect(ctx context.Context, tenant string) error {
	return s.multiDB.Disconnect(ctx, tenant)
}
func (s *RemoteService) PingService(ctx context.Context, tenant string) error {
	return s.multiDB.PingService(ctx, tenant)
}
func (s *RemoteService) PingDatabase(ctx context.Context, tenant string) error {
	return s.multiDB.PingDatabase(ctx, tenant)
}

// PingBridge checks the bridge server/registry availability (tenant-agnostic)
func (s *RemoteService) PingBridge(ctx context.Context) error {
	if s.registry == nil {
		return fmt.Errorf("registry not initialized")
	}
	_ = ctx
	return nil
}

// PingAgent checks that an agent for the given tenant is available in the registry.
func (s *RemoteService) PingAgent(ctx context.Context, tenant string) error {
	if tenant == "" {
		return fmt.Errorf("tenant is required")
	}
	_, err := s.registry.Pick(ctx, tenant)
	return err
}

func (s *RemoteService) UnitService(tenant string) types.UnitService {
	return s.unitFactory(tenant)
}

func (s *RemoteService) InvoiceService(tenant string) types.InvoiceService {
	return s.invoiceFactory(tenant)
}

func (s *RemoteService) PartService(tenant string) types.PartService {
	return s.partFactory(tenant)
}

// Utilities reused from embedded pprof registration
func registerPprof(mux *http.ServeMux, base string) {
	if base == "" {
		base = "/debug/pprof/"
	}
	mux.Handle(base, http.HandlerFunc(pprof.Index))
	mux.Handle(base+"cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle(base+"profile", http.HandlerFunc(pprof.Profile))
	mux.Handle(base+"symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle(base+"trace", http.HandlerFunc(pprof.Trace))
}
