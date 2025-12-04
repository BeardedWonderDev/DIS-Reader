package disreader

import (
	"testing"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Ensures Build succeeds and auto-starts listeners when none are provided.
func TestRemoteBuilderAutoServers(t *testing.T) {
	t.Setenv("DISREADER_BRIDGE_PORT", "0")
	t.Setenv("DISREADER_BRIDGE_HTTP_PORT", "0")

	cfg := &types.DISConfig{Bridge: &types.BridgeConfig{Mode: "remote", MaxRowsPerQuery: 1000}}

	remote, err := NewDISReaderRemote(cfg, nil).Build()
	if err != nil {
		t.Fatalf("build remote: %v", err)
	}
	if remote == nil {
		t.Fatalf("expected remote service")
	}
}
