package state

import (
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

const QueryLatest = ^uint64(0) >> 1

func (s *State) QueryAccount(account database.AccountID) (database.Account, error) {
	return s.db.Query(account)
}

// func (s *State) QueryBlocksByNumber(from, to uint64) ([]database.Block, error) {
// 	if from > to {
// 		return nil, errors.New("from must be less than or equal to to")
// 	}
// 	return s.db.QueryBlocksByNumber(from, to)
// }

func (s *State) QueryBlocksByNumber(from, to uint64) ([]database.Block, error) {
	if from > to {
		return nil, errors.New("from must be less than or equal to to")
	}

	if from == QueryLatest {
		from = s.db.LatestBlock().Header.Number
		to = from
	}

	if to == QueryLatest {
		to = s.db.LatestBlock().Header.Number
	}

	var out []database.Block
	for i := from; i <= to; i++ {
		block, err := s.db.GetBlock(i)
		if err != nil {
			s.evHandler("Error getting block %d: %v", i, err)
			return nil, err
		}
		out = append(out, block)

	}

	return out, nil
}
