package database

import (
	"context"
	"testing"
	"time"

	bridgeproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Ensure Select/Get decode path hits mapstructure hook (dates, strings).
func TestRemoteDBSelectDecodesStruct(t *testing.T) {
	type dest struct {
		Name string    `mapstructure:"name"`
		When time.Time `mapstructure:"when"`
	}
	var out []dest
	now := time.UnixMilli(1_700_000_000_000) // fixed time

	agent := &fakeAgent{
		res: &bridgeproto.JobResult{
			Status: bridgeproto.Status_STATUS_OK,
			Rows: []*bridgeproto.Row{{
				Fields: map[string]*structpb.Value{
					"name": structpb.NewStringValue("alice"),
					"when": structpb.NewNumberValue(float64(now.UnixMilli())),
				},
			}},
		},
	}
	reg := &fakeRegistry{agent: agent}
	db := NewRemoteDB(reg, nil, 10, 0)

	err := db.Select(context.Background(), &out, "select", "tenant-1")
	if err != nil {
		t.Fatalf("select err: %v", err)
	}
	if len(out) != 1 || out[0].Name != "alice" || out[0].When.IsZero() {
		t.Fatalf("expected decoded struct slice, got %+v", out)
	}
}
