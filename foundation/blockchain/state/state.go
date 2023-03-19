package state

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

const (
	ConsensusPoA = "PoA"
	ConsensusPoW = "PoW"
)

// EventHandler defines a function that can be called when an event occurs.
type EventHandler func(v string, args ...any)

// Worker interface represents the behaviour required to be implemented by an package
// providing support for mining, peer update and trasaction sharing.
type Worker interface {
	Shutdown()
	Sync()
	SignalStartMining()
	SignalCancelMining()
	SignalShareTx(blockTx database.BlockTx)
}

type Config struct {
	Beneficiary    database.AccountID
	Host           string
	Storage        database.Storage
	Genesis        genesis.Genesis
	SelectStrategy string
	KnownPeers     *peer.PeerSet
	EvHandler      EventHandler
	Consensus      string
}

type State struct {
	mu sync.RWMutex
	// resyncWG    sync.WaitGroup
	// allowMining bool

	beneficiaryID database.AccountID
	evHandler     EventHandler
	host          string
	consensus     string

	knownPeers *peer.PeerSet
	storage    database.Storage
	genesis    genesis.Genesis
	mempool    *mempool.Mempool

	db *database.Database

	Worker Worker
}

func New(cfg Config, ev func(v string, args ...any)) (*State, error) {

	db, err := database.New(cfg.Genesis, cfg.Storage, ev)
	if err != nil {
		return nil, err
	}

	mempool, err := mempool.NewWithStrategy(cfg.SelectStrategy)
	if err != nil {
		return nil, err
	}

	state := State{
		beneficiaryID: cfg.Beneficiary,
		storage:       cfg.Storage,
		evHandler:     ev,
		host:          cfg.Host,
		consensus:     cfg.Consensus,

		knownPeers: cfg.KnownPeers,
		genesis:    cfg.Genesis,
		mempool:    mempool,
		db:         db,
	}
	// The Worker is not set here. The call to worker.Run will assign itself
	// and start everything up and running for the node

	return &state, nil
}

func (s *State) Shutdown() error {
	s.evHandler("state: shutdown started")
	defer s.evHandler("state: shutdown completed")

	// defer func() {
	// 	s.db.Close()
	// }()

	// Stop all the blockchain writing activity.
	s.Worker.Shutdown()

	// Wait for the resync to complete.
	// s.resyncWG.Wait()

	return nil
}

func (s *State) Consensus() string {
	return s.consensus
}

func (s *State) Genesis() genesis.Genesis {
	return s.genesis
}

func (s *State) LatestBlock() database.Block {
	return s.db.LatestBlock()
}
func (s *State) Mempool() []database.BlockTx {
	return s.mempool.PickBest()
}

func (s *State) MempoolLength() int {
	return s.mempool.Count()
}

func (s *State) UpsertMempool(tx database.BlockTx) error {
	return s.mempool.Upsert(tx)
}

func (s *State) Accounts() map[database.AccountID]database.Account {
	return s.db.Copy()
}
