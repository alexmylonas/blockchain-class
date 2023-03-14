package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Tx struct {
	FromID uint16 `json:"from"`
	ToID   uint16 `json:"to"`
	Amount uint16 `json:"amount"`
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}
func run() error {
	tx := Tx{
		FromID: 1,
		ToID:   2,
		Amount: 1000,
	}
	privateKey, err := crypto.LoadECDSA("zblock/accounts/kennedy.ecdsa")
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	v := crypto.Keccak256Hash(data)

	sig, err := crypto.Sign(v.Bytes(), privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}

	s := hexutil.Encode(sig)
	fmt.Println(string(s))

	return nil

}
