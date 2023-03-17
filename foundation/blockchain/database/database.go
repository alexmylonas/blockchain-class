package database

import (
	"errors"
	"sort"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

type Database struct {
	mu          sync.RWMutex
	genesis     genesis.Genesis
	latestBlock Block
	accounts    map[AccountID]Account
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

func (db *Database) UpdateLatestBlock(block Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.latestBlock = block
}

func (db *Database) LatestBlock() Block {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.latestBlock
}

func (db *Database) HashState() string {
	accounts := make([]Account, 0, len(db.accounts))
	db.mu.RLock()
	{
		for _, account := range db.accounts {
			accounts = append(accounts, account)
		}
	}
	db.mu.RUnlock()

	sort.Sort(byAccount(accounts))

	return signature.Hash(accounts)
}

func (db *Database) ApplyMiningReward(block Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	account := db.accounts[block.Header.BeneficiaryID]

	account.Balance += db.genesis.MiningReward

	db.accounts[block.Header.BeneficiaryID] = account
}

func (db *Database) ApplyTransaction(block Block, tx BlockTx) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	from, exists := db.accounts[tx.FromID]
	if !exists {
		return errors.New("from account not found")
	}

	to, exists := db.accounts[tx.ToID]
	if !exists {
		to = newAccount(tx.ToID, 0)
	}

	bnfc, exists := db.accounts[block.Header.BeneficiaryID]
	if !exists {
		bnfc = newAccount(block.Header.BeneficiaryID, 0)
	}

	gasFee := tx.GasPrice * tx.GasUnits
	if gasFee > from.Balance {
		gasFee = from.Balance
	}

	from.Balance -= gasFee
	bnfc.Balance += gasFee

	db.accounts[tx.FromID] = from
	db.accounts[block.Header.BeneficiaryID] = bnfc

	// Perform basic accounting checks
	{
		if tx.Nonce != (from.Nonce + 1) {
			return errors.New("invalid nonce")
		}
		if from.Balance == 0 || from.Balance < (tx.Value+tx.Tip) {
			return errors.New("insufficient funds")
		}
	}

	// Perform the transfer
	from.Balance -= tx.Value
	to.Balance += tx.Value

	// Give benefiaciary the tip
	from.Balance -= tx.Tip
	bnfc.Balance += tx.Tip

	from.Nonce = tx.Nonce

	// Update the final changes to the accounts
	db.accounts[tx.FromID] = from
	db.accounts[tx.ToID] = to
	db.accounts[block.Header.BeneficiaryID] = bnfc

	return nil
}
