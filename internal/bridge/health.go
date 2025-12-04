package bridge

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HealthHandler returns a simple JSON health/readiness report.
// Ready = agents connected; Live = process serving.
func HealthHandler(registry AgentRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := registry.Stats()
		status := "ready"
		code := http.StatusOK
		if stats.TotalAgents == 0 {
			status = "not_ready"
			code = http.StatusServiceUnavailable
		}
		resp := map[string]any{
			"status":        status,
			"total_agents":  stats.TotalAgents,
			"tenants":       stats.Tenants,
			"documentation": "wiki/bridge_mode.md",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// MetricsHandler exposes minimal Prometheus-style gauges for agent counts.
func MetricsHandler(registry AgentRegistry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := registry.Stats()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP bridge_agents_total Connected agents\n")
		fmt.Fprintf(w, "# TYPE bridge_agents_total gauge\n")
		fmt.Fprintf(w, "bridge_agents_total %d\n", stats.TotalAgents)
		for tenant, count := range stats.Tenants {
			fmt.Fprintf(w, "bridge_agents_per_tenant{tenant=\"%s\"} %d\n", tenant, count)
		}
	})
}
