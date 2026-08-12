# ByteDesk remote-gateway plugin template

GitHub **template** for a gateway process plugin.

1. **Use this template** (or clone) and rename the Go module + `plugin.json` `id`.
2. Implement handlers in `cmd/<id>/`.
3. Validate and pack with the plugin SDK.
4. Install under `GATEWAY_HOME/plugins/<id>/` and Control Bus `rescan` + `enable`.
   Nav/panels appear **without restarting** the gateway.

```bash
# Rename: plugin.json id, binary, scopes, module path, cmd/ directory.

go build -o example ./cmd/example
go run github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk/cmd/plugin-sdk@latest validate --dir .
go run github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk/cmd/plugin-sdk@latest pack --dir . --out dist

mkdir -p ~/.bytedesk-emote-gateway/plugins/example
cp plugin.json example ~/.bytedesk-emote-gateway/plugins/example/
# POST /api/plugins/control {"id":"example","action":"rescan"}
# POST /api/plugins/control {"id":"example","action":"enable"}
# GET /p/example/ui  (session cookie)
```

Host ABI: unix socket `GATEWAY_PLUGIN_SOCKET`. Do not use Go `.so`.
Keep UI self-contained (host CSP). See gateway ADR 0014 and
`docs/plugins/PLUGIN_AUTHOR_GUIDE.md`.
