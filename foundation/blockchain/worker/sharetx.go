package worker

// CORE NOTE: Sharing new transactions received directly by a wallet is performed by this goroutines.
// When a wallet transaction is received, the request goroutine shares it with this goroutine to send it over the p2p network.
// Up to 100 transactions can be pending to be sent before new transactions are dropped and not sent.

// maxTxShareRequests is the maximum number of transactions that can be pending to be sent over the p2p network.
// Tp keep this simple, a buffered channel of this arbitrary number is being used. If the channel becomes full, requests
// for new transactions to be shared will not be accepted.
const maxTxShareRequests = 100

func (w *Worker) shareTxOperations() {
	w.evHandler("worker: shareTxOperations: started")
	defer w.evHandler("worker: shareTxOperations: stopped")

	for {
		select {
		// TODO: This is a blocking operation. If the channel is full, this will block until a new transaction is received.
		// This is not a problem for now, but it should be changed to a non-blocking operation.
		// TODO: pull more than one transaction at a time.
		case tx := <-w.txSharing:
			if !w.isShutdown() {
				w.evHandler("worker: shareTxOperations: received tx to share")
				w.state.NetSendTxToPeers(tx)
			}
		case <-w.shutdown:
			w.evHandler("worker: shareTxOperations: shutdown")
			return
		}
	}
}
