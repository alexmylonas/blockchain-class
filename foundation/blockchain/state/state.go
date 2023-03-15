package state

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

// EventHandler defines a function that can be called when an event occurs.
type EventHandler func(v string, args ...any)

type Config struct {
	Beneficiary database.AccountID
	// Host           string
	// Storage        database.Storage
	Genesis genesis.Genesis
	// SelectStrategy string
	// KnownPeers     *peer.PeerSet
	EvHandler EventHandler
	// Consensus      string
}

type State struct {
	mu sync.RWMutex
	// resyncWG    sync.WaitGroup
	// allowMining bool

	beneficiaryID database.AccountID
	evHandler     EventHandler
	// host          string
	// consensus     string

	// knownPeers *peer.PeerSet
	// storage    database.Storage
	genesis genesis.Genesis
	// mempool    *mempool.Mempool
	db *database.Database

	// Worker Worker
}

func New(cfg Config, ev func(v string, args ...any)) (*State, error) {

	db, err := database.New(cfg.Genesis, ev)
	if err != nil {
		return nil, err
	}

	state := State{
		beneficiaryID: cfg.Beneficiary,
		evHandler:     ev,
		// host:          cfg.Host,
		// consensus:     cfg.Consensus,
		// knownPeers:    cfg.KnownPeers,
		// storage:       cfg.Storage,
		genesis: cfg.Genesis,
		// mempool:       mempool.New(),
		db: db,
	}

	return &state, nil
}

func (s *State) Shutdown() error {
	s.evHandler("state: shutdown started")
	defer s.evHandler("state: shutdown completed")

	// defer func() {
	// 	s.db.Close()
	// }()

	// Stop all the blockchain writing activity.
	// s.Worker.Shutdown()

	// Wait for the resync to complete.
	// s.resyncWG.Wait()

	return nil
}
