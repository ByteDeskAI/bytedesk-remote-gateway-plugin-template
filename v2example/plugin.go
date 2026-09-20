// Package v2example is the smallest complete v2 (NATS transport) reference
// plugin: one typed event published on a schedule, a KV bucket, and one served
// command that reads and updates it.
//
// This is a different "v2" from exampleplugin's: that one demonstrates the
// manifest CONTRACT version (kind, identity, static, ...) over the v1 SDK's
// unix-socket RPC wire. This one demonstrates the v2 SDK MODULE — the
// generation-2 transport where a plugin embeds plugin.Base and talks to the
// gateway over real NATS, with no host-provided translator. A plugin author
// starting a v2 (real-bus) plugin should copy this package; one starting a v1
// (RPC) plugin should copy exampleplugin instead.
//
// The heartbeat event and the state command are deliberately NOT wired
// together: OwnNamespace grants a plugin publish (and serve) rights over its
// own event.<id>.>/svc.<id>.> trees but not subscribe, because a plugin's own
// broadcast is for OTHER plugins to consume — the validator refuses a manifest
// that declares a subscribe permission for its own namespace as restating an
// implicit grant (BDP2114). This reference therefore shows the typed event as
// what it actually is — an announcement for someone else — and demonstrates
// the KV bucket in the served command's own read-increment-write instead.
//
// The typed descriptors below (heartbeatEvent, stateCommand, stateBucketDesc)
// are normally output by contractgen from a schema, once per contract package
// (see bytedesk-sdk-dependencies/v2/plugin/typed.go). This template hand-builds
// the three of them so the whole reference fits in one file; a real plugin with
// more than a couple of operations should generate them instead.
package v2example

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pluginsdk "github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk/v2"
	"github.com/ByteDeskAI/bytedesk-sdk-dependencies/v2/bus"
	"github.com/ByteDeskAI/bytedesk-sdk-dependencies/v2/plugin"
)

// ID is this plugin's identity. Every subject below sits under svc.<ID>. or
// event.<ID>., which is what OwnNamespace grants and Manifest.Serves declares.
const ID = "example-v2"

const (
	heartbeatSubject  bus.Subject = "event.example-v2.v1.heartbeat"
	stateSubject      bus.Subject = "svc.example-v2.v1.state"
	stateBucketName               = "example-v2-state"
	stateKey                      = "count"
	heartbeatSchedule             = "example-v2-heartbeat"
	heartbeatEvery                = 30 * time.Second
)

// Heartbeat is the one typed event this plugin emits: a fixed payload on a
// timer, for whichever other plugin cares to subscribe. bus.Scheduler.Every
// publishes a payload captured at registration, not a live callback, so a
// value that needs to change per tick belongs to a stream or a served command
// instead, not to a scheduled event.
type Heartbeat struct {
	Source string `json:"source"`
}

// State is the KV-backed counter the served command reads and increments.
type State struct {
	Count int `json:"count"`
}

// stateRequest is deliberately empty: the served command takes no input.
type stateRequest struct{}

// The schema hash below is a placeholder, not a real content hash: contractgen
// computes one from the payload's actual JSON schema, and a descriptor with an
// EMPTY hash stamps nothing and is refused on arrival by design (typed.go's
// checkSchema — an author cannot accidentally opt out of the check). Any
// non-empty, stable string works here because this package is both the only
// sender and the only receiver; a generated contract's hash additionally lets
// an unrelated package detect a revision mismatch, which a hand-rolled string
// cannot.
var (
	heartbeatEvent  = plugin.NewEvent[Heartbeat]("event.example-v2.v1.heartbeat", 1, "example-v2-heartbeat-v1", heartbeatSubject)
	stateCommand    = plugin.NewCommand[stateRequest, State]("svc.example-v2.v1.state", 1, "example-v2-state-v1", stateSubject)
	stateBucketDesc = plugin.NewBucketDescriptor[State](stateBucketName, 1, "example-v2-state-v1")
)

// Plugin embeds Base and nothing else: every capability it uses (Bus, Logger)
// comes from the accessors Base inherits once the host binds it.
type Plugin struct {
	plugin.Base

	svc bus.Service
}

// New constructs an unbound Plugin. It becomes live when the host (or, in a
// test, plugin.Bind) installs a Binding.
func New() *Plugin { return &Plugin{} }

func (p *Plugin) ID() string { return ID }

// Manifest declares the served endpoint, the KV bucket the host provisions,
// and nothing under Permissions: publishing this plugin's own event and
// serving its own svc.<id>.> tree are both implicit (OwnNamespace), and
// restating either is refused. A real plugin also fills Version from its
// release process; this template pins it for reproducibility.
func (p *Plugin) Manifest() pluginsdk.Manifest {
	return pluginsdk.Manifest{
		ID:      ID,
		Version: "0.1.0",
		Kind:    pluginsdk.KindProcess,
		Targets: []string{pluginsdk.TargetGateway},
		Binary:  "example-v2",
		Socket:  "plugin.sock",
		Identity: &pluginsdk.ManifestIdentity{
			DisplayName: "Example v2 plugin",
			Description: "Reference for the v2 SDK module: a typed event, a KV bucket and one served command.",
		},
		// A real plugin names its own publisher; this is the template's
		// placeholder for whoever forks it.
		Publisher: &pluginsdk.Publisher{ID: "example", Name: "Example"},
		Serves: []pluginsdk.ServiceDecl{{
			Name:    "example-v2",
			Version: "1.0.0",
			Endpoints: []pluginsdk.EndpointDecl{{
				Name:    "state",
				Subject: stateSubject,
			}},
		}},
		KV: []pluginsdk.KVDecl{{Name: stateBucketName, MaxBytes: 65536, History: 1}},
	}
}

// Start opens the bucket, serves the state command, then schedules the
// heartbeat. Unwinding on a partial failure matters here for the same reason
// ServePlugin's own doc says it does: a plugin half-wired is worse than one
// that never started.
func (p *Plugin) Start(ctx context.Context) (err error) {
	b := p.Bus()
	if _, err := plugin.OpenBucket(ctx, b, stateBucketDesc); err != nil {
		return fmt.Errorf("open bucket %s: %w", stateBucketName, err)
	}

	svc, err := plugin.Serve(ctx, b, stateCommand, p.serveState)
	if err != nil {
		return fmt.Errorf("serve %s: %w", stateSubject, err)
	}
	defer func() {
		if err != nil {
			_ = svc.Stop(ctx)
		}
	}()

	payload, err := json.Marshal(Heartbeat{Source: p.ID()})
	if err != nil {
		return fmt.Errorf("encode heartbeat: %w", err)
	}
	if err := b.Schedule().Every(ctx, heartbeatSchedule, heartbeatEvery, heartbeatSubject, payload,
		bus.WithSchema(heartbeatEvent.Descriptor().SchemaHash())); err != nil {
		return fmt.Errorf("schedule heartbeat: %w", err)
	}

	p.svc = svc
	return nil
}

// Stop cancels the schedule and releases the served endpoint in the reverse
// order Start acquired them.
func (p *Plugin) Stop(ctx context.Context) error {
	_ = p.Bus().Schedule().Cancel(ctx, heartbeatSchedule)
	if p.svc != nil {
		_ = p.svc.Stop(ctx)
	}
	return nil
}

// serveState answers with the count after incrementing it, or 1 on the first
// call. It reads-then-writes rather than using Update's compare-and-swap
// because this reference has exactly one writer; a design with more than one
// should retry on bus.FaultConflict instead.
func (p *Plugin) serveState(ctx context.Context, _ stateRequest, _ bus.Caller) (State, error) {
	bucket, err := plugin.OpenBucket(ctx, p.Bus(), stateBucketDesc)
	if err != nil {
		return State{}, err
	}
	cur, _, err := bucket.Get(ctx, stateKey)
	if err != nil && !isNotFound(err) {
		return State{}, err
	}
	cur.Count++
	if _, err := bucket.Put(ctx, stateKey, cur); err != nil {
		return State{}, err
	}
	return cur, nil
}

func isNotFound(err error) bool {
	f, ok := err.(bus.Fault)
	return ok && f.Code == bus.FaultNotFound
}

var _ pluginsdk.Plugin = (*Plugin)(nil)
