package worker

import "github.com/ardanlabs/blockchain/foundation/blockchain/peer"

// CORE NOTE The p2p network is managed by this this goroutine.
// There is a single node that is considered the "leader" of the network.
// The default in main.go represents the leader. That leader node must running first.
// All new peer nodes will connect to the leader node to identify the network.
// The topology of the network is a star topology. The leader node is the center
// of the star and all other nodes are the points of the star.
// If a node does not respond to a network call it will be removed from the network.
func (w *Worker) peerOperations() {

}

func (w *Worker) addNewPeers(knownPeers []peer.Peer) error {
	w.evHandler("worker: runPeerUpdateOperations: addNewPeers started")
	defer w.evHandler("worker: runPeerUpdateOperations: addNewPeers completed")

	for _, peer := range knownPeers {
		if peer.Match(w.state.Host()) {
			continue
		}

		if w.state.AddKnownPeer(peer) {
			w.evHandler("worker: runPeerUpdateOperations: addNewPeers: added peer %s", peer.Host)
		}
	}

	return nil
}
