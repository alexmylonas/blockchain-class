package mempool

import (
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
)

type Mempool struct {
	mu       sync.RWMutex
	pool     map[string]database.BlockTx
	selectFn selector.Func
}

func New() (*Mempool, error) {
	// TODO: Implement the advanced tip strategy.
	return NewWithStrategy(selector.StrategyTip)
}

func NewWithStrategy(strategy string) (*Mempool, error) {
	selectFn, err := selector.Retrieve(strategy)
	if err != nil {
		return nil, err
	}

	mp := Mempool{
		pool:     make(map[string]database.BlockTx),
		selectFn: selectFn,
	}

	return &mp, nil
}

func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.pool)
}

func (mp *Mempool) Upsert(tx database.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// CORE NOTE: Different blockchains will have different algorithms to limit
	// the size of the mempool. This is a simple example of how to do it..
	// Some limit based on the amount of being consumed by the transactions.
	// if a limit is met, then either the transaction with the lowest fee
	// or the oldest transaction is removed.

	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	// Ethereum uses a nonce to prevent replay attacks. If the nonce is not the next one in the sequence
	// then the transaction is rejected. This is a simple example of how to do it.

	// Ethereum also has a concept of gas. Gas is the amount of work that is being done by the transaction.
	// The gas is paid for by the sender of the transaction. If the sender does not have enough gas to

	// Ethereum requires a bump of at least 10% for the gas price
	// to replace a transaction in the mempool.

	if etx, exists := mp.pool[key]; exists {
		if tx.Tip < uint64(math.Round(float64(etx.Tip)*1.10)) {
			return fmt.Errorf("replacing a transaction requires a 10%% bump in the tip")
		}
	}

	mp.pool[key] = tx

	return nil
}

func (mp *Mempool) Delete(tx database.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	delete(mp.pool, key)

	return nil
}

func (mp *Mempool) Truncate() error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.pool = make(map[string]database.BlockTx)

	return nil
}

func (mp *Mempool) PickBest(howMany ...uint16) []database.BlockTx {
	number := 0
	if len(howMany) > 0 {
		number = int(howMany[0])
	}

	m := make(map[database.AccountID][]database.BlockTx)
	mp.mu.RLock()
	{
		if number == 0 {
			number = len(mp.pool)
		}

		for key, tx := range mp.pool {
			account := accountFromMapKey(key)
			m[account] = append(m[account], tx)
		}
	}
	mp.mu.RUnlock()

	return mp.selectFn(m, number)

}

func mapKey(tx database.BlockTx) (string, error) {
	return fmt.Sprintf("%s:%d", tx.FromID, tx.Nonce), nil
}

func accountFromMapKey(key string) database.AccountID {
	return database.AccountID(strings.Split(key, ":")[0])
}
