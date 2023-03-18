package state

import (
	"context"
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

var ErrNoTransactions = errors.New("no transactions in mempool")

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

	if err := s.validateUpdateDatabase(block); err != nil {
		return database.Block{}, err
	}
	return block, nil
}

// =============================================================================
func (s *State) validateUpdateDatabase(block database.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evHandler("state: validateUpdateDatabase: VALIDATING block")

	// CORE NOTE: We could add logic to determine if this block was mine or not.
	// If it was mined by this, even if a peer beat me to this function for the same block
	// number, I could replace the peer block with my own.

	if err := block.ValidateBlock(s.db.LatestBlock(), s.db.HashState(), s.evHandler); err != nil {
		return err
	}

	// Write block to database.
	if err := s.db.Write(block); err != nil {
		return err
	}

	// Update the state with the new block.
	s.db.UpdateLatestBlock(block)

	s.evHandler("state: validateUpdateDatabase: UPDATED LATEST BLOCK head is now [%d]", block.Header.Number)

	s.evHandler("state: validateUpdateDatabase: UPDATING state with new block")

	for _, tx := range block.MerkleTree.Values() {
		s.evHandler("state: validateUpdateDatabase: UPDATING state with tx [%s]", tx)

		s.mempool.Delete(tx)

		if err := s.db.ApplyTransaction(block, tx); err != nil {
			s.evHandler("state: validateUpdateDatabase: ERROR [%s]", err)
			continue
		}
	}

	s.evHandler("state: validateUpdateDatabase: applying Mining Reward")

	s.db.ApplyMiningReward(block)

	// Send an event about this new block
	// s.blockEvent(block)

	return nil
}
