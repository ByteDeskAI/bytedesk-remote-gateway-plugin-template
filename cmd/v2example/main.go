// Command v2example runs the v2 (NATS transport) reference plugin as a spawned
// gateway process.
package main

import (
	"context"
	"log"

	pluginsdk "github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk/v2"
	v2example "github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-template/v2example"
)

func main() {
	if err := pluginsdk.ServePlugin(context.Background(), v2example.New(), pluginsdk.PluginConfig{}); err != nil {
		log.Fatal(err)
	}
}
