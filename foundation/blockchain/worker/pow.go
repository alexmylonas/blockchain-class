package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
)

// CORE NOTE: The PoW mining operation is managed by this functions which runs on its
// own goroutine. When a startMining signal is received (mainly because a wallet transaction was received)
// a block is created and then the POW operations starts.
// This operations can be cancelled if a proposed block is received and it's validated.

// powOperations is the main POW mining operation.
func (w *Worker) powOperations() {
	w.evHandler("worker: powOperations: Goroutine started")
	defer w.evHandler("worker: powOperations: Goroutine completed")

	for {
		select {
		case <-w.startMining:
			if !w.isShutdown() {
				w.runPowOperations()
			}
		case <-w.shutdown:
			w.evHandler("worker: powOperations: received shut signal")
			return
		}
	}
}

func (w *Worker) runPowOperations() {
	w.evHandler("worker: runPowOperations: started")
	defer w.evHandler("worker: runPowOperations: completed")

	// if !w.state.IsMiningAllowed() {
	// 	w.evHandler("worker: runPowOperations: mining is not allowed")
	// 	return
	// }

	memLen := w.state.MempoolLength()
	if memLen == 0 {
		w.evHandler("worker: runPowOperations: mempool is empty")
		return
	}

	// After running a mining operation, check if a new operation should
	// signaled again
	defer func() {
		length := w.state.MempoolLength()
		if length > 0 {
			w.evHandler("worker: runPowOperations: signaling new mining operations Txs[%d]", length)
			w.SignalStartMining()
		}
	}()

	// Drain the cancel signal channel
	select {
	case <-w.cancelMining:
		w.evHandler("worker: runPowOperations: drained cancel channel")
	default:
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		select {
		case <-w.cancelMining:
			w.evHandler("worker: runPowOperations: Cancel Requested")
		case <-ctx.Done():
		}
	}()

	// The GoRoutine that performs mining operations
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		t := time.Now()
		_, err := w.state.MineNewBlock(ctx)
		duration := time.Since(t)

		w.evHandler("worker: runPowOperations: MineNewBlock completed in %s", duration)

		if err != nil {
			switch {
			case errors.Is(err, state.ErrNoTransactions):
				w.evHandler("worker: runPowOperations: MineNewBlock: no transactions in mempool")
			case ctx.Err() != nil:
				w.evHandler("worker: runPowOperations: MineNewBlock: context canceled")
			default:
				w.evHandler("worker: runPowOperations: MineNewBlock: %s", err)
			}
			return
		}

		// WOW, we mined a block
		// TODO: send the block to the network
	}()

	wg.Wait()
}
