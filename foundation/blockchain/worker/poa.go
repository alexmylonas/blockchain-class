package worker

import (
	"context"
	"errors"
	"hash/fnv"
	"sort"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
)

// CORE NOTE The POA mining operation is managed by this functio which runs on its own goroutine
// The node starts a loop that is on an 5 second interval. At the beggingin of each cycle the selection algorithm
// is executed to determine if this node needs to mine the next block. If this node is not selected,
// it will wait for the next cycle. If this node is selected, it will mine the next block.

// cycleDuration sets the operations to every 5 seconds

const secondsPerCycle = 5
const cycleDuration = secondsPerCycle * time.Second

// poaOperations is the main loop for the worker
func (w *Worker) poaOperations() {
	w.evHandler("worker: poaOperations: Goroutine started")
	defer w.evHandler("worker: poaOperations Goroutine completed")

	// On startup talk to the leader node and get an updated peers list.
	// Then share with the network that this node is available for transactions
	// and block submissions.

	ticker := time.NewTicker(cycleDuration)

	resetTicker(ticker, cycleDuration)

	for {
		select {
		case <-ticker.C:
			if !w.isShutdown() {
				w.runPoaOperations()
			}
		case <-w.shutdown:
			w.evHandler("worker: poaOperations: shutdown received")
			return
		}

		// Reset the ticker for the next cycle.
		resetTicker(ticker, 0)
	}
}

// resetTicker resets the ticker to the next cycle.
func resetTicker(ticker *time.Ticker, waitOnSecond time.Duration) {
	nextTick := time.Now().Add(cycleDuration).Round(waitOnSecond)
	diff := time.Until(nextTick)
	ticker.Reset(diff)
}

// runPOAOperations is the main loop for the worker
func (w *Worker) runPoaOperations() {
	w.evHandler("worker: runPoaOperations: started")
	defer w.evHandler("worker: runPoaOperations: completed")

	peer := w.selection()
	w.evHandler("worker: runPoaOperations: Host %s, SELECTED PEER %s", w.state.Host(), peer)
	if peer != w.state.Host() {
		return
	}

	length := w.state.MempoolLength()
	if length == 0 {
		w.evHandler("worker: runPoaOperations: Host %s, No transactions in mempool", w.state.Host())
		return
	}

	select {
	case <-w.cancelMining:
		w.evHandler("worker: runPoaOperations: MINING cancelled")
		return
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
			w.evHandler("worker: runPoaOperations: MINING cancelled")
		case <-ctx.Done():
		}
	}()

	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		t := time.Now()
		block, err := w.state.MineNewBlock(ctx)
		duration := time.Since(t)
		w.evHandler("worker: runPoaOperations: MINING completed in %v", duration)

		if err != nil {
			switch {
			case errors.Is(err, state.ErrNoTransactions):
				w.evHandler("worker: runPoaOperations: MINING failed: %v", err)
			case ctx.Err() != nil:
				w.evHandler("worker: runPoaOperations: MINING cancelled")
			default:
				w.evHandler("worker: runPoaOperations: MINING failed: %v", err)
			}
		}

		// The block is mined. Propose the new block to the network.
		if err := w.state.NetSendBlockToPeers(block); err != nil {
			w.evHandler("worker: runPoaOperations: MINING: proposeBlockToPeers failed: %v", err)
		}
	}()

}

func (w *Worker) selection() string {
	peers := w.state.KnownPeers()

	w.evHandler("worker: runPoaOperations: selection: Host %s, List of Peers %v", w.state.Host(), peers)

	names := make([]string, len(peers))
	for i, peer := range peers {
		names[i] = peer.Host
	}

	sort.Strings(names)

	// Based on the latest block, pick an index number from the registry

	latestBlock := w.state.LatestBlock()
	lastBlockHash := latestBlock.Hash()
	h := fnv.New32a()

	h.Write([]byte(lastBlockHash))
	integerHash := h.Sum32()
	i := integerHash % uint32(len(names))

	return names[i]
}
