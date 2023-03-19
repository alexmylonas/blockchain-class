package worker

import (
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
)

type Worker struct {
	state        *state.State
	wg           sync.WaitGroup
	ticker       time.Ticker
	shutdown     chan struct{}
	startMining  chan bool
	cancelMining chan bool
	txSharing    chan database.BlockTx
	evHandler    state.EventHandler
}

// Run creates a worker, registers it with the state and starts it.
func Run(st *state.State, evHandler state.EventHandler) {

	w := Worker{
		state:        st,
		ticker:       *time.NewTicker(powpeerInterval),
		shutdown:     make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		txSharing:    make(chan database.BlockTx, maxTxShareRequests),
		evHandler:    evHandler,
	}

	consensusOperation := w.powOperations
	if st.Consensus() == state.ConsensusPoA {
		consensusOperation = w.poaOperations
		// Validate ticker works as intended
		// w.ticker = *time.NewTicker(poaPeerInterval)
	}

	st.Worker = &w

	// Sync node before starting any worker operations.
	w.Sync()

	// Load the set of operations to run.

	operations := []func(){
		consensusOperation,
		w.peerOperations,
		w.shareTxOperations,
	}

	g := len(operations)
	w.wg.Add(g)

	// We don't want to return until we know all the goroutines have started.
	hasStarted := make(chan bool)

	for _, op := range operations {
		go func(op func()) {
			defer w.wg.Done()
			hasStarted <- true
			op()
		}(op)
	}

	// Wait for all the goroutines to start, before returnin to main.
	for i := 0; i < g; i++ {
		<-hasStarted
	}
}

func (w *Worker) Shutdown() {
	w.evHandler("worker: shutdown: started")
	defer w.evHandler("worker: shutdown: completed")

	w.evHandler("worker: shutting: signal cancel mining")
	w.SignalCancelMining()

	w.evHandler("worker: shutdown: terminate goroutines")
	close(w.shutdown)
	w.wg.Wait()
}

func (w *Worker) isShutdown() bool {
	select {
	case <-w.shutdown:
		return true
	default:
		return false
	}
}

func (w *Worker) SignalStartMining() {
	// if !w.state.IsMiningAllowed() {
	// 	w.evHandler("state: MinePeerBlock: accepting blocks is turned off")
	// 	return
	// }

	// Only PoW requires signaling to start mining
	// if w.state.Consensus != state.ConsensusPoW {
	// 	return
	// }

	select {
	case w.startMining <- true:
	default:
	}
	w.evHandler("worker: SignalStartMing: mining signaled")
}

func (w *Worker) SignalShareTx(tx database.BlockTx) {
	select {
	case w.txSharing <- tx:
		w.evHandler("worker: SignalShareTx: tx sharing signaled")
	default:
		// This means the channel is full. We don't want to block here.
		w.evHandler("worker: SignalShareTx: tx sharing signaled (dropped)")
	}
}

func (w *Worker) SignalCancelMining() {
	select {
	case w.cancelMining <- true:
	default:
	}
	w.evHandler("worker: SignalCancelMining: mining canceled")
}
