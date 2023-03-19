// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/private"
	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

const version = "v1"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// PublicRoutes binds all the version 1 public routes.
func PublicRoutes(app *web.App, cfg Config) {
	pbl := public.Handlers{
		Log:   cfg.Log,
		State: cfg.State,
		NS:    cfg.NS,
	}

	// app.Handle(http.MethodGet, version, "/events", pbl.Events)
	app.Handle(http.MethodGet, version, "/genesis/list", pbl.Genesis)

	app.Handle(http.MethodGet, version, "/accounts/list", pbl.Accounts)
	app.Handle(http.MethodGet, version, "/accounts/list/:account", pbl.Accounts)

	// app.Handle(http.MethodGet, version, "/blocks/list/", pbl.BlocksByAccount)
	// app.Handle(http.MethodGet, version, "/blocks/list/:account", pbl.BlocksByAccount)

	app.Handle(http.MethodGet, version, "/tx/uncommited/list", pbl.Mempool)
	app.Handle(http.MethodGet, version, "/tx/uncommited/list/:account", pbl.Mempool)

	app.Handle(http.MethodPost, version, "/tx/commit", pbl.SubmitWalletTx)
	// app.Handle(http.MethodPost, version, "/tx/proof/:block", pbl.SubmitWalletTx)

}

// PrivateRoutes binds all the version 1 private routes.
func PrivateRoutes(app *web.App, cfg Config) {
	prv := private.Handlers{
		Log:   cfg.Log,
		NS:    cfg.NS,
		State: cfg.State,
	}

	app.Handle(http.MethodPost, version, "/node/peers", prv.SubmitPeer)
	app.Handle(http.MethodGet, version, "/node/status", prv.Status)
	app.Handle(http.MethodGet, version, "/node/tx/list", prv.Mempool)

	app.Handle(http.MethodPost, version, "/node/tx/submit", prv.SubmitNodeTransaction)

	app.Handle(http.MethodPost, version, "/node/block/propose", prv.ProposeBlock)

	blocskUri := fmt.Sprintf(peer.BlocksUri, ":from", ":to")
	app.Handle(http.MethodGet, version, "/node"+blocskUri, prv.BlocksByNumber)
}
