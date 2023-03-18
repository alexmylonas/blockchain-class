package worker

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
)

type Worker struct {
	state *state.State
	wg    sync.WaitGroup
	// ticker       time.Ticker
	shutdown     chan struct{}
	startMining  chan bool
	cancelMining chan bool
	// txSharing    chan database.BlockTx
	evHandler state.EventHandler
}

// Run creates a worker, registers it with the state and starts it.
func Run(st *state.State, evHandler state.EventHandler) {
	w := Worker{
		state:        st,
		shutdown:     make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		// txSharing:    make(chan database.BlockTx),
		evHandler: evHandler,
	}

	st.Worker = &w

	// Sync node before starting any worker operations.
	w.Sync()
	consensusOperation := w.powOperations
	// if st.Consensus == state.ConsensusPoA {
	// 	consensusOperation = w.poaOperations
	// }

	operations := []func(){
		consensusOperation,
		// w.peerOperations,
		// w.shareTxOperations,
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

	// Wait for all the goroutines to start.
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

func (w *Worker) SignalCancelMining() {
	select {
	case w.cancelMining <- true:
	default:
	}
	w.evHandler("worker: SignalCancelMining: mining canceled")
}
