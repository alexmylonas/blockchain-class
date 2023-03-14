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

	privateKey, err := crypto.LoadECDSA("zblock/accounts/kennedy.ecdsa")
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	tx := Tx{
		FromID: 1,
		ToID:   2,
		Amount: 1000,
	}

	data, err := json.Marshal(tx)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	v := crypto.Keccak256(data)

	sig, err := crypto.Sign(v, privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign data: %w", err)
	}

	s := hexutil.Encode(sig)
	fmt.Println("SIG:", string(s))

	// ==================================================
	// OVER THE WIRE

	publicKey, err := crypto.SigToPub(v, sig)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	address := crypto.PubkeyToAddress(*publicKey).String()
	fmt.Println("PUB:", address)
	return nil

}
