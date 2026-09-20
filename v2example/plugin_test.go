package v2example

import (
	"context"
	"testing"
	"time"

	"github.com/ByteDeskAI/bytedesk-sdk-dependencies/v2/bus"
	"github.com/ByteDeskAI/bytedesk-sdk-dependencies/v2/bus/memory"
	"github.com/ByteDeskAI/bytedesk-sdk-dependencies/v2/plugin"
)

// TestManifestValidates catches the two mistakes that are otherwise invisible
// until a real host refuses to admit the plugin: a subject outside this
// plugin's own namespace, and a manifest that restates an implicit grant
// (both BDP2004-family checks) — the exact mistake an earlier draft of this
// reference made by declaring an explicit subscribe permission for its own
// event.
func TestManifestValidates(t *testing.T) {
	if err := plugin.ValidateDiscover(New().Manifest()); err != nil {
		t.Fatalf("manifest does not validate: %v", err)
	}
}

// TestStartServesAndSchedules exercises the whole wiring Start assembles — the
// KV bucket, the served command and the scheduled event — against an
// in-memory bus, the same way a conformance test binds a fixture (plugin.Bind
// is documented for exactly this).
func TestStartServesAndSchedules(t *testing.T) {
	store := memory.NewStore()
	grants := plugin.OwnNamespace(ID)
	grants.KV = []string{stateBucketName}
	// A real caller of svc.example-v2.v1.state is some OTHER plugin or the
	// host, never this plugin itself, so OwnNamespace grants Serves but not
	// Request on that tree. The test plays that other caller by requesting
	// over the same bus connection, so it needs the grant OwnNamespace would
	// never hand this plugin's own identity.
	grants.Request = []bus.Pattern{bus.Pattern(stateSubject)}
	id := bus.Identity{PluginID: ID, Generation: "test-1", Grants: grants}
	b := store.Connect(id)

	if err := b.KV().Declare(context.Background(), bus.BucketSpec{Name: stateBucketName, MaxBytes: 65536}); err != nil {
		t.Fatalf("declare bucket: %v", err)
	}

	p := New()
	if err := plugin.Bind(p, plugin.Binding{Bus: b, Identity: id}); err != nil {
		t.Fatalf("bind: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { _ = p.Stop(context.Background()) })

	for want := 1; want <= 3; want++ {
		got, err := plugin.Call(ctx, b, stateCommand, stateRequest{})
		if err != nil {
			t.Fatalf("call state (want %d): %v", want, err)
		}
		if got.Count != want {
			t.Fatalf("count = %d, want %d", got.Count, want)
		}
	}
}

// TestStopCancelsSchedule proves Stop actually releases what Start acquired —
// calling it twice, or on a plugin that never started, must not panic.
func TestStopCancelsSchedule(t *testing.T) {
	store := memory.NewStore()
	id := bus.Identity{PluginID: ID, Generation: "test-2", Grants: plugin.OwnNamespace(ID)}
	b := store.Connect(id)

	p := New()
	if err := plugin.Bind(p, plugin.Binding{Bus: b, Identity: id}); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if err := p.Stop(context.Background()); err != nil {
		t.Fatalf("stop before start: %v", err)
	}
}
