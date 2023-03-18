package peer

import (
	"fmt"
	"sync"
)

const (
	BaseUrl     = "http://%s/v1/node"
	StatusUri   = "/status"
	MempoolUri  = "/tx/list"
	BlocksUri   = "/block/list/%s/%s"
	PeerUri     = "/peers"
	TxSubmitUri = "/tx/submit"
)

type Peer struct {
	Host string
}

func New(host string) Peer {
	return Peer{
		Host: host,
	}
}

func (p Peer) Match(host string) bool {
	return p.Host == host
}

type PeerStatus struct {
	LatestBlockHash string `json:"latest_block_hash"`
	LatestBlockNum  uint64 `json:"latest_block_num"`
	KnownPeers      []Peer `json:"known_peers"`
}

type PeerSet struct {
	mu  sync.RWMutex
	set map[Peer]struct{}
}

func NewPeerSet() *PeerSet {
	return &PeerSet{
		set: make(map[Peer]struct{}),
	}
}

func (p Peer) Url() string {
	return fmt.Sprintf(BaseUrl, p.Host)
}

func (ps *PeerSet) Add(peer Peer) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, exists := ps.set[peer]
	if !exists {
		ps.set[peer] = struct{}{}
		return true
	}

	return false
}

func (ps *PeerSet) Remove(peer Peer) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, exists := ps.set[peer]
	if exists {
		delete(ps.set, peer)
		return true
	}

	return false
}

func (ps *PeerSet) Copy(host string) []Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	peers := make([]Peer, 0, len(ps.set)-1)
	for peer := range ps.set {
		if peer.Host != host { // Removing self
			peers = append(peers, peer)
		}
	}
	return peers
}
