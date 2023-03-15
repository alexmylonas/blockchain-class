// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/private"
	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

const version = "v1"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
}

// PublicRoutes binds all the version 1 public routes.
func PublicRoutes(app *web.App, cfg Config) {
	pbl := public.Handlers{
		Log:   cfg.Log,
		State: cfg.State,
	}

	// app.Handle(http.MethodGet, version, "/events", pbl.Events)
	app.Handle(http.MethodGet, version, "/genesis/list", pbl.Genesis)

	app.Handle(http.MethodGet, version, "/accounts/list", pbl.Accounts)
	app.Handle(http.MethodGet, version, "/accounts/list/:account", pbl.Accounts)

	// app.Handle(http.MethodGet, version, "/blocks/list/", pbl.BlocksByAccount)
	// app.Handle(http.MethodGet, version, "/blocks/list/:account", pbl.BlocksByAccount)

	// app.Handle(http.MethodGet, version, "/tx/uncommited/list", pbl.Mempool)
	// app.Handle(http.MethodGet, version, "/tx/uncommited/list/:account", pbl.Mempool)

	// app.Handle(http.MethodPost, version, "/tx/commit", pbl.SubmitWalletTx)
	// app.Handle(http.MethodPost, version, "/tx/proof/:block", pbl.SubmitWalletTx)

}

// PrivateRoutes binds all the version 1 private routes.
func PrivateRoutes(app *web.App, cfg Config) {
	prv := private.Handlers{
		Log: cfg.Log,
	}

	app.Handle(http.MethodGet, version, "/node/sample", prv.Sample)
}
