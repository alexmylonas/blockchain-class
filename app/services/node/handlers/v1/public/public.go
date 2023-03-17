// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"net/http"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
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
	// WS websocket.Upgrader
	// Evts *events.Events
}

func (h Handlers) Cancel(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h.State.Worker.SignalStartMining()
	resp := struct {
		Status string `json:"status"`
	}{
		Status: "cancel",
	}
	return web.Respond(ctx, w, resp, http.StatusOK)
}

func (h Handlers) SubmitWalletTx(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("getting values")
	}

	var signedTx database.SignedTx
	if err := web.Decode(r, &signedTx); err != nil {
		return err
	}

	h.Log.Infow("submitting transaction", "traceID", v.TraceID, "sig:nonce", signedTx, "from", signedTx.FromID, "to", signedTx.ToID, "value", signedTx.Value, "tip", signedTx.Tip)

	// Asks the state to add the transaction to the mempool.
	// Only the checks that the transaction signature and the reciept account format
	// Its up to wallet to check the balance and nonce.
	// Fee will be taken if this transaction is included in a block.
	if err := h.State.UpsertWalletTx(signedTx); err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}
	resp := struct {
		Status string `json:"status"`
	}{
		Status: "transaction added to mempool",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Sample just provides a starting point for the class.
func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen := h.State.Genesis()

	return web.Respond(ctx, w, gen, http.StatusOK)
}

func (h Handlers) Accounts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accouStr := web.Param(r, "account")

	var accounts map[database.AccountID]database.Account
	switch accouStr {
	case "":
		accounts = h.State.Accounts()

	default:
		accountID, err := database.ToAccountID(accouStr)
		if err != nil {
			return err
		}

		account, err := h.State.QueryAccount(accountID)
		if err != nil {
			return err
		}
		accounts = map[database.AccountID]database.Account{accountID: account}
	}

	return web.Respond(ctx, w, accounts, http.StatusOK)
}

func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "account")

	mempool := h.State.Mempool()

	trans := []tx{}

	for _, tran := range mempool {
		// Ignoring transactions that don't match the account.
		if acct != "" && acct != string(tran.FromID) && (acct != string(tran.ToID)) {
			continue
		}

		trans = append(trans, tx{
			FromAccount: tran.FromID,
			ToAccount:   tran.ToID,
			FromName:    h.NS.Lookup(tran.FromID),
			ToName:      h.NS.Lookup(tran.ToID),
			ChainID:     tran.ChainID,
			Nonce:       tran.Nonce,
			Value:       tran.Value,
			Tip:         tran.Tip,
			Data:        tran.Data,
			TimeStamp:   tran.TimeStamp,
			GasPrice:    tran.GasPrice,
			GasUnits:    tran.GasUnits,
			Sig:         tran.SignatureString(),
		})
	}
	return web.Respond(ctx, w, trans, http.StatusOK)
}
