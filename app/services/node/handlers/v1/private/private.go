// Package private maintains the group of handlers for node to node access.
package private

import (
	"context"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// Sample just provides a starting point for the class.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latestBlock := h.State.LatestBlock()

	status := peer.PeerStatus{
		LatestBlockHash: latestBlock.Hash(),
		LatestBlockNum:  latestBlock.Header.Number,
		KnownPeers:      h.State.KnowExternalPeers(),
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
