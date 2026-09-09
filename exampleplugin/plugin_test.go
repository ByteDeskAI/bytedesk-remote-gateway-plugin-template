package example

import (
	"context"
	"encoding/json"
	pluginsdk "github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"
)

type testHost struct{}

func (testHost) Publish(pluginsdk.Envelope) error                  { return nil }
func (testHost) Subscribe(string, func(pluginsdk.Envelope)) func() { return func() {} }
func (testHost) Request(context.Context, pluginsdk.Envelope) (pluginsdk.Envelope, error) {
	return pluginsdk.Envelope{}, nil
}
func (testHost) Logger() pluginsdk.Logger           { return nil }
func (testHost) StateDir(string) string             { return "" }
func (testHost) Every(time.Duration, func()) func() { return func() {} }
func (testHost) BumpContributions()                 {}

func TestIndependentConstructionAndWithdrawal(t *testing.T) {
	p := New()
	request := func() int {
		w := httptest.NewRecorder()
		p.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/example/hello", nil))
		return w.Code
	}
	if request() != 503 {
		t.Fatal("unstarted plugin served")
	}
	if err := p.Start(context.Background(), testHost{}); err != nil {
		t.Fatal(err)
	}
	if err := p.CheckActivation(context.Background()); err != nil {
		t.Fatal(err)
	}
	if request() != 200 {
		t.Fatal("started plugin did not serve full declared path")
	}
	if err := p.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if request() != 503 {
		t.Fatal("withdrawn plugin still served")
	}
}

func TestPackageManifestMatchesImplementation(t *testing.T) {
	raw, err := os.ReadFile("../plugin.json")
	if err != nil {
		t.Fatal(err)
	}
	var m pluginsdk.Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m, New().Manifest()) {
		t.Fatal("regenerate plugin.json with go run ./cmd/manifest")
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}
