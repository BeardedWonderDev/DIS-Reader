package bridge

import (
	"context"
	"errors"
	"sync"

	"github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
)

// AgentConnection abstracts an active agent stream.
type AgentConnection interface {
	TenantID() string
	AgentID() string
	SendJob(ctx context.Context, req *proto.JobRequest) (*proto.JobResult, error)
	Close() error
}

// AgentRegistry tracks agent connections per tenant.
type AgentRegistry interface {
	Register(ctx context.Context, tenantID string, agentID string, conn AgentConnection) error
	Unregister(ctx context.Context, tenantID string, agentID string)
	Pick(ctx context.Context, tenantID string) (AgentConnection, error)
	Stats() RegistryStats
}

// RegistryStats is a snapshot of connected agents per tenant.
type RegistryStats struct {
	TotalAgents int
	Tenants     map[string]int
}

// InMemoryRegistry is a simple in-process registry suitable for single-instance deployments.
type InMemoryRegistry struct {
	mu     sync.RWMutex
	agents map[string]map[string]AgentConnection // tenant -> agentID -> conn
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		agents: map[string]map[string]AgentConnection{},
	}
}

func (r *InMemoryRegistry) Register(ctx context.Context, tenantID string, agentID string, conn AgentConnection) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agents[tenantID]; !ok {
		r.agents[tenantID] = map[string]AgentConnection{}
	}
	r.agents[tenantID][agentID] = conn
	return nil
}

func (r *InMemoryRegistry) Unregister(ctx context.Context, tenantID string, agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.agents[tenantID]; !ok {
		return
	}
	delete(r.agents[tenantID], agentID)
	if len(r.agents[tenantID]) == 0 {
		delete(r.agents, tenantID)
	}
}

func (r *InMemoryRegistry) Pick(ctx context.Context, tenantID string) (AgentConnection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agents, ok := r.agents[tenantID]
	if !ok || len(agents) == 0 {
		return nil, errors.New("no agents available for tenant")
	}
	for _, conn := range agents {
		return conn, nil // first available
	}
	return nil, errors.New("no agents available")
}

func (r *InMemoryRegistry) Stats() RegistryStats {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stats := RegistryStats{Tenants: map[string]int{}}
	for tenant, agents := range r.agents {
		stats.Tenants[tenant] = len(agents)
		stats.TotalAgents += len(agents)
	}
	return stats
}
