package state

import (
	"context"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

func (s *State) UpsertWalletTx(signedTx database.SignedTx) error {
	if err := signedTx.Validate(s.genesis.ChainID); err != nil {
		return err
	}

	const oneUnitOfGas = 1
	tx := database.NewBlockTx(signedTx, uint64(s.genesis.GasPrice), oneUnitOfGas)
	if err := s.mempool.Upsert(tx); err != nil {
		return err
	}

	if s.mempool.Count() == 6 {
		go func() {
			s.MineNewBlock(context.Background())
			s.mempool.Truncate()
		}()
	}
	// s.Worker.SignalShareTx(tx)
	// s.Worker.SingalStartMining()
	return nil
}
