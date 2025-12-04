package bridge

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Exported via types for external consumers.
type AgentConnection = types.AgentConnection
type AgentRegistry = types.AgentRegistry
type RegistryStats = types.RegistryStats
type RegistryKey = types.RegistryKey
type LatencyAgg = types.LatencyAgg

// InMemoryRegistry is a simple in-process registry suitable for single-instance deployments.
type InMemoryRegistry struct {
	mu     sync.RWMutex
	agents map[string]map[string]AgentConnection // tenant -> agentID -> conn
	counts map[RegistryKey]int64
	lat    map[RegistryKey]LatencyAgg
}

func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		agents: map[string]map[string]AgentConnection{},
		counts: map[RegistryKey]int64{},
		lat:    map[RegistryKey]LatencyAgg{},
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
	stats := RegistryStats{Tenants: map[string]int{}, QueryCounts: map[RegistryKey]int64{}, Latency: map[RegistryKey]LatencyAgg{}}
	for tenant, agents := range r.agents {
		stats.Tenants[tenant] = len(agents)
		stats.TotalAgents += len(agents)
	}
	for k, v := range r.counts {
		stats.QueryCounts[k] = v
	}
	for k, v := range r.lat {
		stats.Latency[k] = v
	}
	return stats
}

func (r *InMemoryRegistry) ObserveQuery(tenantID, agentID string, latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := RegistryKey{Tenant: tenantID, Agent: agentID}
	r.counts[key]++
	agg := r.lat[key]
	agg.Count++
	agg.Sum += latency.Seconds()
	if latency.Seconds() > agg.Max {
		agg.Max = latency.Seconds()
	}
	r.lat[key] = agg
}
