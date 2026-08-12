package main

import (
	"log"
	"net/http"
	"os"

	pluginsdk "github.com/ByteDeskAI/bytedesk-remote-gateway-plugin-sdk"
)

func main() {
	id := os.Getenv(pluginsdk.EnvID)
	if id == "" {
		id = "example"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/hello", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/ui", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><meta charset="utf-8"><title>Example plugin</title>
<style>body{font:14px system-ui,sans-serif;background:#10131a;color:#e9ecf1;margin:16px}
code{background:#1a1f2b;padding:2px 6px;border-radius:4px}</style></head>
<body>
<h1>Example plugin</h1>
<p>Served via host proxy <code>/p/` + id + `/</code>. No gateway restart.</p>
<p id="st">checking…</p>
<script>
fetch("hello").then(r=>r.text()).then(t=>{document.getElementById("st").textContent="API /hello → "+t;})
.catch(e=>{document.getElementById("st").textContent="API error: "+e;});
</script>
</body></html>`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	log.Printf("plugin id=%s", id)
	if err := pluginsdk.Serve(pluginsdk.Config{ID: id, Handler: mux}); err != nil {
		log.Fatal(err)
	}
}
