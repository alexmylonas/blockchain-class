package selector

import (
	"sort"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

func allTransactions(m map[database.AccountID][]database.BlockTx) []database.BlockTx {
	var all []database.BlockTx
	for _, txs := range m {
		all = append(all, txs...)
	}
	return all
}

func newAdvancedTip(m map[database.AccountID][]database.BlockTx, howMany int) map[database.AccountID]int {
	at := make(map[database.AccountID]int)
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(ByNonce(m[key]))
		}
	}

	for {
		for key := range m {
			if len(m[key]) > 0 {
				at[key]++
				m[key] = m[key][1:]
			}
		}

		if len(at) == 0 {
			break
		}

		var total int
		for _, num := range at {
			total += num
		}
		if total >= howMany {
			break
		}
	}
	return at
}

var advancedTipSelect = func(m map[database.AccountID][]database.BlockTx, howMany int) []database.BlockTx {
	// Calculate the total number of transactions.
	var total int
	for _, txs := range m {
		total += len(txs)
	}
	if total == 0 {
		return nil
	}
	if total <= howMany {
		return allTransactions(m)
	}

	final := make([]database.BlockTx, 0, howMany)
	for key := range m {
		if len(m[key]) > 1 {
			sort.Sort(ByNonce(m[key]))
		}
	}

	at := newAdvancedTip(m, howMany)
	for from, num := range at {
		for i := 0; i < num; i++ {
			final = append(final, m[from][i])
		}
	}

	return final
}

type advancedTips struct {
	m map[database.AccountID]int
}
