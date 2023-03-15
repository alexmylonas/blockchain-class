// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	// NS *nameservice.nameservice
	// WS websocket.Upgrader
	// Evts *events.Events
}

// Sample just provides a starting point for the class.
func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h.Log.Infow("genesis", "traceID", web.GetTraceID(ctx))
	gen := h.State.Genesis()

	return web.Respond(ctx, w, gen, http.StatusOK)
}
