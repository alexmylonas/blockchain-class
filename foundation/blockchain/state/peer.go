package state

import "github.com/ardanlabs/blockchain/foundation/blockchain/peer"

func (s *State) KnowExternalPeers() []peer.Peer {
	return s.knownPeers.Copy(s.host)
}

func (s *State) Host() string {
	return s.host
}

func (s *State) AddKnownPeer(peer peer.Peer) bool {
	return s.knownPeers.Add(peer)
}

func (s *State) RemoveKnownPeer(peer peer.Peer) bool {
	return s.knownPeers.Remove(peer)
}

func (s *State) KnownPeers() []peer.Peer {
	return s.knownPeers.Copy("")
}
