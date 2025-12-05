package disreader

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge"
	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"github.com/BeardedWonderDev/DIS-Reader/internal/database"
	svc "github.com/BeardedWonderDev/DIS-Reader/internal/service"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
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
	auth          types.AgentAuthenticator
	reg           types.AgentRegistry
	grpcServer    *grpc.Server
	mux           *http.ServeMux
	defaultTenant string
}

func (b *RemoteBuilder) WithAuth(auth types.AgentAuthenticator) *RemoteBuilder {
	b.auth = auth
	return b
}

func (b *RemoteBuilder) WithRegistry(reg types.AgentRegistry) *RemoteBuilder {
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

	agentCfg := buildAgentConfig(b.cfg.Bridge)

	server := bridge.NewServer(auth, registry, logger,
		bridge.WithAutoConnectOnRegister(b.cfg.Bridge.AutoConnectOnRegister),
		bridge.WithAgentConfig(agentCfg))

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
		srv := grpc.NewServer(
			grpc.ChainUnaryInterceptor(grpcLoggingUnary(logger)),
			grpc.ChainStreamInterceptor(grpcLoggingStream(logger)),
		)
		remote.RegisterBridge(srv)
		go func() {
			logger.Info("bridge gRPC listening", slog.String("addr", lis.Addr().String()))
			if err := srv.Serve(lis); err != nil {
				logger.Error("bridge gRPC server error", slog.Any("err", err))
			}
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
			httpLogger := slog.NewLogLogger(logger.Handler(), slog.LevelError)
			srv := &http.Server{Addr: addr, Handler: mux, ErrorLog: httpLogger}
			logger.Info("bridge HTTP listening", slog.String("addr", addr))
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("bridge HTTP server error", slog.Any("err", err))
			}
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

// buildAgentConfig converts the bridge Loki configuration into a proto payload
// that can be sent to agents during registration.
func buildAgentConfig(cfg *types.BridgeConfig) *bridgeproto.AgentConfig {
	if cfg == nil || cfg.Loki == nil || cfg.Loki.URL == "" {
		return nil
	}

	labels := cfg.Loki.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	return &bridgeproto.AgentConfig{
		Loki: &bridgeproto.LokiConfig{
			Url:        cfg.Loki.URL,
			TenantId:   cfg.Loki.TenantID,
			ApiKey:     cfg.Loki.APIKey,
			AuthHeader: cfg.Loki.AuthHeader,
			Labels:     labels,
			MinLevel:   strings.ToLower(cfg.Loki.MinLevel.String()),
		},
	}
}

// grpcLoggingUnary logs unary RPCs using the provided slog logger.
func grpcLoggingUnary(logger *slog.Logger) grpc.UnaryServerInterceptor {
	if logger == nil {
		logger = slog.Default()
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		level := slog.LevelInfo
		if code != codes.OK {
			level = slog.LevelError
		}
		logger.LogAttrs(ctx, level, "grpc request",
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Duration("duration", time.Since(start)),
		)
		return resp, err
	}
}

// grpcLoggingStream logs stream RPCs using the provided slog logger.
func grpcLoggingStream(logger *slog.Logger) grpc.StreamServerInterceptor {
	if logger == nil {
		logger = slog.Default()
	}
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, stream)
		code := status.Code(err)
		level := slog.LevelInfo
		if code != codes.OK {
			level = slog.LevelError
		}
		logger.LogAttrs(stream.Context(), level, "grpc stream",
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Bool("is_client_stream", info.IsClientStream),
			slog.Bool("is_server_stream", info.IsServerStream),
			slog.Duration("duration", time.Since(start)),
		)
		return err
	}
}
