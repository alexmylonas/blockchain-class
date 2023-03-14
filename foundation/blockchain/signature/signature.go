package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
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

func ToSignatureBytes(v, r, s *big.Int) []byte {
	// Create a buffer to hold the signature
	sig := make([]byte, crypto.SignatureLength)

	rBytes := make([]byte, 32)
	r.FillBytes(rBytes)
	copy(sig, rBytes)

	sBytes := make([]byte, 32)
	s.FillBytes(sBytes)
	copy(sig[32:], sBytes)

	sig[64] = byte(v.Uint64() - ardanID)

	return sig
}
