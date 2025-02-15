// Package private maintains the group of handlers for node to node access.
package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
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

func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	tsx := h.State.Mempool()
	return web.Respond(ctx, w, tsx, http.StatusOK)
}

func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fromStr := web.Param(r, "from")
	if fromStr == "" || fromStr == "latest" {
		fromStr = fmt.Sprintf("%d", state.QueryLatest)
	}

	toStr := web.Param(r, "to")
	if toStr == "" || toStr == "latest" {
		toStr = fmt.Sprintf("%d", state.QueryLatest)
	}

	from, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	to, err := strconv.ParseUint(toStr, 10, 64)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	if from > to {
		return v1.NewRequestError(fmt.Errorf("to must be greater than from"), http.StatusBadRequest)
	}

	blocks, err := h.State.QueryBlocksByNumber(from, to)
	if err != nil {
		return v1.NewRequestError(err, http.StatusInternalServerError)
	}
	if len(blocks) == 0 {
		return v1.NewRequestError(fmt.Errorf("no blocks found"), http.StatusNotFound)
	}

	blockData := make([]database.BlockData, len(blocks))
	for i, block := range blocks {
		blockData[i] = database.NewBlockData(block)
	}

	return web.Respond(ctx, w, blockData, http.StatusOK)
}

func (h Handlers) SubmitPeer(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value misisng from context")
	}

	var peer peer.Peer
	if err := web.Decode(r, &peer); err != nil {
		return web.NewShutdownError("unabled to decode peer payload")
	}

	ok := h.State.AddKnownPeer(peer)
	if !ok {
		h.Log.Infow("adding peer", "traceId", v.TraceID, "host", peer.Host)
	}

	return web.Respond(ctx, w, nil, http.StatusOK)

}

func (h Handlers) SubmitNodeTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value misisng from context")
	}

	var tx database.BlockTx
	if err := web.Decode(r, &tx); err != nil {
		return web.NewShutdownError("unabled to decode peer payload")
	}

	h.Log.Infow("submitting transaction", "traceId", v.TraceID, "tx", tx)
	if err := h.State.UpsertNodeTransaction(tx); err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}

// ProposeBlock
func (h Handlers) ProposeBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	// Decode the JSON in the post call into a file system block
	var blockData database.BlockData
	if err := web.Decode(r, &blockData); err != nil {
		return err
	}

	// Convert the block data into a block. This action will create a merkle
	// tree of the set of transactions required for blockchain operations

	block, err := database.ToBlock(blockData)
	if err != nil {
		return err
	}

	if err := h.State.ProcessProposedBlock(block); err != nil {
		// if errors.Is(err, database.ErrChainForked) {
		// h.State.Reorganize()
		// }

		return v1.NewRequestError(errors.New("block rejected"), http.StatusNotAcceptable)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}
