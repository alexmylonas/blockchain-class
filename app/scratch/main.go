package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Tx struct {
	FromID string `json:"from"`
	ToID   string `json:"to"`
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
		FromID: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
		ToID:   "Aaron",
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

	publicKey, err := crypto.SigToPub(v, sig)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	address := crypto.PubkeyToAddress(*publicKey).String()
	fmt.Println("PUB:", address)

	// ==================================================
	// OVER THE WIRE

	tx2 := Tx{
		FromID: "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32",
		ToID:   "Aarjon",
		Amount: 1000,
	}

	data2, err := json.Marshal(tx2)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	v2 := crypto.Keccak256(data2)

	publicKey2, err := crypto.SigToPub(v2, sig)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	address2 := crypto.PubkeyToAddress(*publicKey2).String()
	fmt.Println("PUB:", address2)
	return nil

}
