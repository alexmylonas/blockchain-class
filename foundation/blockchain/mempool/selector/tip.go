package selector

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

var tipSelect = func(m map[database.AccountID][]database.BlockTx, howMany int) []database.BlockTx {

	// Sorting account transactions by nonce.
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(ByNonce(m[key]))
		}
	}

	// Pick the first transaction in the slice for each account. each iteration represents
	// a new row of selection keep doing until all the transactions are selected.
	var rows [][]database.BlockTx

	for {
		var row []database.BlockTx
		for key := range m {
			if len(m[key] > 0) {
				row = append(row, m[key][0])
				m[key] = m[key][1:]
			}
		}
		if row == nil {
			break
		}
		rows = append(rows, row)
	}

	// Sort the rows by tip and unless we weill take all transactions from that row anywway
	// Then try to select the number of requested transactions. Keep pulling transactions from
	// each row until we have the number of transactions requested.
	// or we run out of transactions.
	final := make([]database.BlockTx, 0, howMany)
	for _, row := range rows {
		need := howMany - len(final)
		if len(row) > need {
			sort.Sort(ByTip(row))
			final = append(final, row[:need]...)
			break
		}
		final = append(final, row...)
	}
	return final
}
