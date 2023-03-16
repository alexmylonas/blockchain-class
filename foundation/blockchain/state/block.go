package state

import (
	"context"
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

var ErrNoTransactions = errors.New("no transactions in mempool")

const (
	ConsensusPoA = "PoA"
)

func (s *State) MineNewBlock(ctx context.Context) (database.Block, error) {
	defer s.evHandler("viewer: MineNewBlock: MINING completed")

	s.evHandler("viewer: MineNewBlock: MINING checking mempool")

	if s.mempool.Count() == 0 {
		return database.Block{}, ErrNoTransactions
	}

	trans := s.mempool.PickBest(s.genesis.TransPerBlock)

	// If PoA is being used drop diffulty to 1.
	difficulty := s.genesis.Difficulty
	// if s.Consensus == ConsensusPoA {
	// difficulty = 1
	// }

	s.evHandler("viewer: MineNewBlock: MINING creating new block")

	block, err := database.POW(ctx, database.POWArgs{
		BeneficiaryID: s.beneficiaryID,
		Difficulty:    difficulty,
		MiningReward:  s.genesis.MiningReward,
		PrevBlock:     s.db.LatestBlock(),
		StateRoot:     s.db.HashState(),
		Trans:         trans,
		EvHandler:     s.evHandler,
	})
	if err != nil {
		return database.Block{}, err
	}

	if ctx.Err() != nil {
		return database.Block{}, ctx.Err()
	}
	s.evHandler("viewer: MineNewBlock: MINING adding new block to database")

	// if err := s.validateUpdateDatabase(block); err != nil {
	// }
	return block, nil
}
