package selector

import (
	"fmt"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

const (
	StrategyTip         = "tip"
	StrategyTipAdvanced = "tip_advanced"
)

var strategies = map[string]Func{
	StrategyTip:         tipSelect,
	StrategyTipAdvanced: advancedTipSelect,
}

type Func func(transactions map[database.AccountID][]database.BlockTx, howMany int) []database.BlockTx

func Retrieve(strategy string) (Func, error) {
	if fn, ok := strategies[strategy]; ok {
		return fn, nil
	}

	return nil, fmt.Errorf("selector: unknown strategy %q", strategy)
}

type ByNonce []database.BlockTx

func (bn ByNonce) Len() int {
	return len(bn)
}

func (bn ByNonce) Less(i, j int) bool {
	return bn[i].Nonce < bn[j].Nonce
}

func (bn ByNonce) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}

type ByTip []database.BlockTx

func (bt ByTip) Len() int {
	return len(bt)
}

func (bt ByTip) Less(i, j int) bool {
	return bt[i].Tip > bt[j].Tip
}

func (bt ByTip) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}
