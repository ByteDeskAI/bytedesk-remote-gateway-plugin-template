# ByteDesk remote-gateway plugin template

The `example` package implements the canonical SDK `Plugin`, `HTTPPlugin` and optional `ActivationChecker`. It imports no gateway implementation. A linked host can construct `example.New()`; `cmd/example` serves the same implementation as a process using `ServePlugin`.

1. Rename the module, package ID and declared routes in `exampleplugin/plugin.go`.
2. Implement domain behavior in that independent package. Request only exact host operations needed through manifest Permissions; host policy grants authority.
3. Regenerate the package manifest with `go run ./cmd/manifest > plugin.json`.
4. Run `go test ./...`, then build `go build -o example ./cmd/example`.
5. Validate with `go run github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk/cmd/plugin-sdk validate --dir .` and pack with the same pinned command using `pack --dir . --out dist`.
6. Install the package through the Store/control plane, then enable it. Disabled installation exposes no contributions. The host resolves declared `/example/` routes and `/plugins/example` navigation; legacy `/p/example/` proxy access strips that conventional prefix only.

The template requires protocol major 1 scoped-host and activation-check features. An older host must be upgraded before activation; the SDK negotiation fails closed. `Start` acquires resources, `CheckActivation` admits the candidate before serving, and `Stop` honors its deadline. Register callbacks/timers through the scoped SDK Host and cancel them during Stop. Never control, execute or proxy another plugin.

`ServePlugin` reads `GATEWAY_PLUGIN_ID`, `GATEWAY_PLUGIN_SOCKET` and `GATEWAY_HOST_SOCKET`. New executable code runs in a spawned process, not a Go shared object. Only the host owns operator authentication, admission, transport routing and process supervision. Preserve full declared request paths in handlers.

The UI is self-contained for host CSP. For modules, use the versioned `@bytedesk/gateway-plugin-ui` mount/cleanup contract; untrusted UI requires the host sandbox/broker, not an in-page privilege grant inferred from a signature. See the gateway plugin author guide for operator authorization and containment policy.
