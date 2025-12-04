package disreader

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	svc "github.com/BeardedWonderDev/DIS-Reader/internal/service"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"google.golang.org/grpc"
)

// Embedded entrypoint
func NewDISReaderEmbedded(config *types.DISConfig, logger *slog.Logger) (types.DISReaderService, error) {
	return svc.NewEmbeddedService(config, logger)
}

// RemoteBuilder creates tenant-aware remote services with fluent options.
func NewDISReaderRemote(config *types.DISConfig, logger *slog.Logger) *RemoteBuilder {
	return &RemoteBuilder{cfg: config, logger: logger}
}

type RemoteBuilder struct {
	cfg           *types.DISConfig
	logger        *slog.Logger
	auth          bridge.AgentAuthenticator
	reg           bridge.AgentRegistry
	grpcServer    *grpc.Server
	mux           *http.ServeMux
	defaultTenant string
}

func (b *RemoteBuilder) WithAuth(auth bridge.AgentAuthenticator) *RemoteBuilder {
	b.auth = auth
	return b
}

func (b *RemoteBuilder) WithRegistry(reg bridge.AgentRegistry) *RemoteBuilder {
	b.reg = reg
	return b
}

func (b *RemoteBuilder) WithGRPC(server *grpc.Server) *RemoteBuilder {
	b.grpcServer = server
	return b
}

func (b *RemoteBuilder) WithMux(mux *http.ServeMux) *RemoteBuilder {
	b.mux = mux
	return b
}

func (b *RemoteBuilder) WithDefaultTenant(t string) *RemoteBuilder {
	b.defaultTenant = t
	return b
}

func (b *RemoteBuilder) Build() (types.DISReaderRemote, error) {
	if b.cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if b.cfg.Bridge == nil || b.cfg.Bridge.Mode == "" {
		return nil, fmt.Errorf("bridge config is required for remote mode")
	}
	if b.cfg.Bridge.Mode != "remote" {
		return nil, fmt.Errorf("bridge.mode must be remote for remote builder")
	}

	// Allow config-specified default tenant if caller didn't set WithDefaultTenant.
	if b.defaultTenant == "" && b.cfg.Bridge.DefaultTenant != "" {
		b.defaultTenant = b.cfg.Bridge.DefaultTenant
	}

	logger := b.logger
	if logger == nil {
		logger = slog.Default()
	}

	registry := b.reg
	if registry == nil {
		registry = bridge.NewInMemoryRegistry()
	}

	auth := b.auth
	if auth == nil {
		auth = buildAuthenticator(b.cfg.Bridge, nil)
	}

	server := bridge.NewServer(auth, registry, logger)

	multiDB := database.NewRemoteDB(registry, logger, b.cfg.Bridge.MaxRowsPerQuery, b.cfg.Bridge.MaxResultBytes)

	remote := svc.NewRemoteService(b.cfg, logger, multiDB, registry, server)
	remote.WithDefaultTenant(b.defaultTenant)

	// gRPC server: use provided or start our own
	if b.grpcServer != nil {
		remote.RegisterBridge(b.grpcServer)
	} else {
		grpcPort := envDefault("DISREADER_BRIDGE_PORT", "8443")
		lis, err := net.Listen("tcp", ":"+grpcPort)
		if err != nil {
			return nil, fmt.Errorf("listen gRPC: %w", err)
		}
		srv := grpc.NewServer()
		remote.RegisterBridge(srv)
		go func() {
			logger.Info("bridge gRPC listening", slog.String("addr", lis.Addr().String()))
			_ = srv.Serve(lis)
		}()
	}

	// HTTP mux: use provided or start our own
	if b.mux != nil {
		remote.RegisterHealth(b.mux)
	} else {
		httpPort := envDefault("DISREADER_BRIDGE_HTTP_PORT", "8080")
		mux := http.NewServeMux()
		remote.RegisterHealth(mux)
		go func() {
			addr := ":" + httpPort
			logger.Info("bridge HTTP listening", slog.String("addr", addr))
			_ = http.ListenAndServe(addr, mux)
		}()
	}

	return remote, nil
}

// envDefault returns the env var or fallback.
func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// buildAuthenticator mirrors legacy behavior, now used by the remote builder.
func buildAuthenticator(cfg *types.BridgeConfig, override bridge.AgentAuthenticator) bridge.AgentAuthenticator {
	if override != nil {
		return override
	}
	if cfg == nil {
		return &bridge.StaticAuthenticator{Secrets: map[string]bridge.StaticAgentSecret{}}
	}
	if cfg.CredentialFile != "" {
		if fa, err := bridge.NewFileAuthenticator(cfg.CredentialFile); err == nil {
			return fa
		}
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
