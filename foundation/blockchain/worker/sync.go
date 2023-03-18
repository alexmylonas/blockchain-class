package worker

// CORE NOTE: This function is called when the node starts. It will sync the node with the
// network. This is a blocking operation. It includes the mempool and blochain database.
// This operation needs to finish before the node can participate in the network.

func (w *Worker) Sync() {
	w.evHandler("worker: SYNC: started")
	defer w.evHandler("worker: SYNC: completed")

	for _, peer := range w.state.KnowExternalPeers() {

		peerStatus, err := w.state.NetRequestPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: SYNC: queryPeerStatus: %s ERROR %s", peer.Host, err)
			continue
		}

		w.addNewPeers(peerStatus.KnownPeers)

		pool, err := w.state.NetRequestMempool(peer)
		if err != nil {
			w.evHandler("worker: SYNC: retrievePeerMempool: %s ERROR %s", peer.Host, err)
			continue
		}

		for _, tx := range pool {
			w.evHandler("worker: SYNC: retrievePeerMempool: %s TX %s", peer.Host, tx.SignatureString()[:16])
			w.state.UpsertMempool(tx)
		}

		if peerStatus.LatestBlock > w.state.LatestBlock().Header.Number {
			w.evHandler("worker: SYNC: retrivePeerBlocks: %s: latestBlockNumber [%d]", peer.Host, peerStatus.LatestBlock)

			if err := w.state.NetRequestPeerBlocks(peer); err != nil {
				w.evHandler("worker: SYNC: retrievePeerBlockchain: %s ERROR %s", peer.Host, err)
			}
		}
	}

	// Share with peers this is available to participate in the network.
	w.state.NetSendNodeAvailableToPeers()
}
