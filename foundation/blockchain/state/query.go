package state

import (
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
