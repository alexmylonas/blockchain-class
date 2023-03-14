package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const ZeroHash = "0x0000000000000000000000000000000000000000000000000000000000000000"

const ardanID = 29

func Hash(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ZeroHash
	}

	hash := sha256.Sum256(data)
	return hexutil.Encode(hash[:])
}

func stamp(value any) ([]byte, error) {
	// Marshal the value to JSON
	v, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	stamp := []byte(fmt.Sprintf("\x19Ardan Signed Message:\n%d", len(v)))

	// Create a hash of the data
	data := crypto.Keccak256(stamp, v)

	return data, nil

}

func FromAddress(value any, v, r, s *big.Int) (string, error) {
	// Prepare the data to be signed
	data, err := stamp(value)
	if err != nil {
		return "", err
	}

	// Recover the public key
	publicKey, err := crypto.SigToPub(data, v, r, s)
	if err != nil {
		return "", err
	}

	// Extract the bytes for the original public key
	publickKeyBytes := crypto.FromECDSAPub(publicKey)

	// Create the address from the public key
	address := crypto.PubkeyToAddress(*publicKey)

	// Return the address as a string
	return address.String(), nil
}
func Sign(value any, privateKey *ecdsa.PrivateKey) (v, r, s *big.Int, err error) {

	// Prepare the data to be signed
	data, err := stamp(value)
	if err != nil {
		return nil, nil, nil, err
	}

	sig, err := crypto.Sign(data, privateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// Extract the bytes for the original public key
	publickKeyOrg := privateKey.Public()
	publickKeyECDSA, ok := publickKeyOrg.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, nil, err
	}
	publickKeyBytes := crypto.FromECDSAPub(publickKeyECDSA)

}
