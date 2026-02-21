# Plan: Optional Static File Serving in Go Backend

## Scope
Backend only — no changes to frontend, Dockerfile, or anything outside `backend/`.

## Change: `backend/adp/server/server.go`

The only file that needs to change. At the end of the `New()` function, after the existing route registration, check if a `./static` directory exists. If it does, register a `/` catch-all handler that serves files from it with SPA fallback. If the directory doesn't exist, skip it entirely — the server starts as before (API-only mode).

### Concrete diff

```go
 import (
 	"fmt"
+	"log/slog"
 	"net/http"
+	"os"
 	"sudojo/svc/tenant"

 	"github.com/gorilla/websocket"
 	"github.com/prometheus/client_golang/prometheus/promhttp"
 )
```

Then at the end of `New()`, right before `return s` (after line 70):

```go
	s.router.Handle("/metrics", promhttp.Handler())

+	staticDir := "./static"
+	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
+		slog.Info("serving static files", "dir", staticDir)
+		fs := http.Dir(staticDir)
+		fileServer := http.FileServer(fs)
+		s.router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
+			if f, err := fs.Open(r.URL.Path); err == nil {
+				f.Close()
+				fileServer.ServeHTTP(w, r)
+				return
+			}
+			http.ServeFile(w, r, staticDir+"/index.html")
+		}))
+	}

	return s
```

### How it works

1. **Optional**: `os.Stat("./static")` checks if the directory exists at startup. If it doesn't (or isn't a directory), the block is skipped entirely. No error, no panic — the server runs in API-only mode exactly as before.

2. **Static file serving**: `http.FileServer` serves real files (JS, CSS, images) from `./static` with correct content types and caching headers.

3. **SPA fallback**: If the requested path doesn't match a file on disk (`fs.Open` fails), it serves `./static/index.html`. This lets React Router handle client-side routes like `/l/:id`.

4. **No conflicts with API**: The `/` pattern is the lowest-priority catch-all in `http.ServeMux`. Routes like `/api/lobbies/{id}` and `/metrics` are more specific and always matched first.

5. **Log visibility**: An `slog.Info` line makes it clear in logs whether static serving is active.

### What does NOT change
- `main.go` — no new env vars or parameters
- `handler.go` — no handler changes
- `server` struct — no new fields
- API behavior — completely unchanged
- Server startup — still works without `./static` present

## Files changed

| File | What |
|---|---|
| `backend/adp/server/server.go` | Add `os`, `log/slog` imports; add optional static file handler after route registration |

That's it — one file, ~12 lines added.
