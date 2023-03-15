package database

import (
	"errors"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

type Database struct {
	mu       sync.RWMutex
	genesis  genesis.Genesis
	accounts map[AccountID]Account
}

func New(genesis genesis.Genesis, evHandler func(v string, args ...any)) (*Database, error) {
	db := Database{
		genesis:  genesis,
		accounts: make(map[AccountID]Account),
	}

	for accountStr, balance := range genesis.Balances {
		accountID, err := ToAccountID(accountStr)
		if err != nil {
			return nil, err
		}
		db.accounts[accountID] = newAccount(accountID, balance)

		evHandler("Account %s, Balance: %d", accountID, balance)
	}

	return &db, nil
}

func (db *Database) Remove(accountID AccountID) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.accounts, accountID)
}

func (db *Database) GetAccount(accountID AccountID) (Account, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	account, ok := db.accounts[accountID]
	return account, ok
}

func (db *Database) GetAccounts() []Account {
	db.mu.RLock()
	defer db.mu.RUnlock()

	accounts := make([]Account, 0, len(db.accounts))
	for _, account := range db.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

func (db *Database) Query(accountId AccountID) (Account, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	account, exists := db.accounts[accountId]
	if !exists {
		return Account{}, errors.New("account not found")
	}

	return account, nil
}

func (db *Database) Copy() map[AccountID]Account {
	db.mu.RLock()
	defer db.mu.RUnlock()

	accounts := make(map[AccountID]Account)
	for accountID, account := range db.accounts {
		accounts[accountID] = account
	}
	return accounts
}
