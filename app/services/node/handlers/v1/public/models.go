package public

import "github.com/ardanlabs/blockchain/foundation/blockchain/database"

type tx struct {
	FromAccount database.AccountID `json:"from"`
	ToAccount   database.AccountID `json:"to"`

	FromName string `json:"from_name"`
	ToName   string `json:"to_name"`

	ChainID      uint16   `json:"chain_id"`
	Nonce        uint64   `json:"nonce"`
	Value        uint64   `json:"value"`
	Tip          uint64   `json:"tip"`
	Data         []byte   `json:"data"`
	TimeStamp    uint64   `json:"timestamp"`
	GasPrice     uint64   `json:"gas_price"`
	GasUnits     uint64   `json:"gas_units"`
	Sig          string   `json:"sig"`
	Proof        []string `json:"proof"`
	ProofOfOrder []int64  `json:"proof_order"`
}

type act struct {
	AccountID database.AccountID `json:"account"`
	Name      string             `json:"name"`
	Balance   uint64             `json:"balance"`
	Nonce     uint64             `json:"nonce"`
}

type actInfo struct {
	LatestBlock block `json:"latest_block"`
	Uncommitted int   `json:"uncommitted"`
	Accounts    []act `json:"accounts"`
}

type block struct {
	Hash   string `json:"hash"`
	Number uint64 `json:"number"`
}
