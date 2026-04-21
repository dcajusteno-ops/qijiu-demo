package backend

import (
	"context"
	"net/http"
)

// Startup exposes the internal Wails startup hook to the root entry package.
func (a *App) Startup(ctx context.Context) {
	a.startup(ctx)
}

// Shutdown exposes the internal Wails shutdown hook to the root entry package.
func (a *App) Shutdown(ctx context.Context) {
	a.shutdown(ctx)
}

// ServeImage exposes the internal asset/image handler to the root entry package.
func (a *App) ServeImage(w http.ResponseWriter, r *http.Request) {
	a.serveImage(w, r)
}
