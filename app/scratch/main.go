package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

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

func signTx(tx Tx, privateKey ecdsa.PrivateKey) (s, v, st []byte, err error) {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal transaction: %w", err)
	}

	stamp := []byte(fmt.Sprintf("\x19Ardan Signed Message:\n%d", len(data)))

	vn := crypto.Keccak256(stamp, data)

	sig, err := crypto.Sign(vn, &privateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to sign data: %w", err)
	}
	fmt.Println("SIG:", string(hexutil.Encode(sig)))

	return sig, vn, stamp, nil
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

	sig, v, _, err := signTx(tx, *privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

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

	sig2, v2, _, err := signTx(tx2, *privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign transaction 2: %w", err)
	}

	publicKey2, err := crypto.SigToPub(v2, sig2)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	address2 := crypto.PubkeyToAddress(*publicKey2).String()
	fmt.Println("PUB:", address2)

	hexSig := hexutil.Encode(sig2)

	vv, r, s, _ := ToVRSFromHexSignature(hexSig)

	fmt.Println("V:", vv)
	fmt.Println("R:", r)
	fmt.Println("S:", s)

	return nil

}

func ToVRSFromHexSignature(sigStr string) (v, r, s *big.Int, err error) {
	// Decode the signature
	sig, err := hex.DecodeString(sigStr[2:])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Extract the V, R, and S values
	r = big.NewInt(0).SetBytes(sig[:32])
	s = big.NewInt(0).SetBytes(sig[32:64])
	v = big.NewInt(0).SetBytes([]byte{sig[64]})

	return v, r, s, nil
}
