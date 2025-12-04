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
		// Counters
		fmt.Fprintf(w, "# HELP bridge_queries_total Total queries routed\n")
		fmt.Fprintf(w, "# TYPE bridge_queries_total counter\n")
		for key, count := range stats.QueryCounts {
			fmt.Fprintf(w, "bridge_queries_total{tenant=\"%s\",agent=\"%s\"} %d\n", key.Tenant, key.Agent, count)
		}
		fmt.Fprintf(w, "# HELP bridge_query_latency_seconds Latency histogram (seconds)\n")
		fmt.Fprintf(w, "# TYPE bridge_query_latency_seconds summary\n")
		for key, agg := range stats.Latency {
			if agg.Count == 0 {
				continue
			}
			mean := agg.Sum / float64(agg.Count)
			fmt.Fprintf(w, "bridge_query_latency_seconds{tenant=\"%s\",agent=\"%s\",quantile=\"0.5\"} %.6f\n", key.Tenant, key.Agent, mean) // coarse; real percentile not tracked
			fmt.Fprintf(w, "bridge_query_latency_seconds_sum{tenant=\"%s\",agent=\"%s\"} %.6f\n", key.Tenant, key.Agent, agg.Sum)
			fmt.Fprintf(w, "bridge_query_latency_seconds_count{tenant=\"%s\",agent=\"%s\"} %d\n", key.Tenant, key.Agent, agg.Count)
		}
	})
}
