package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const KennedyPub = "0xF01813E4B85e178A83e29B8E7bF26BD830a25f32"

const CeasarPub = "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76"

type Tx struct {
	FromID string `json:"from"`
	ToID   string `json:"to"`
	Amount uint16 `json:"amount"`
}

func main() {
	err := run2()
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

func run2() error {
	privateKey, err := crypto.LoadECDSA("zblock/accounts/pavel.ecdsa")
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}
	fmt.Println(privateKey.PublicKey)
	address2 := crypto.PubkeyToAddress(privateKey.PublicKey).String()
	fmt.Println("PUB:", address2)
	return nil

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

	fmt.Println("========== TX =================")
	newTx, err := database.NewTx(1, 1, KennedyPub, CeasarPub, 1000, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to create new tx: %w", err)
	}

	signedTx, err := newTx.Sign(privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign tx: %w", err)
	}

	fmt.Println("SIGNED TX:", signedTx)
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
