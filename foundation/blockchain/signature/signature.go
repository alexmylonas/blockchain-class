package signature

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	// Convert the V, R, and S values to a signature
	sig := ToSignatureBytes(v, r, s)

	// Capture the public key from the signature
	publicKey, err := crypto.SigToPub(data, sig)
	if err != nil {
		return "", err
	}

	// Create the address from the public key
	return crypto.PubkeyToAddress(*publicKey).String(), nil

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

	// Check the public key validates the data and signature
	rs := sig[:crypto.RecoveryIDOffset]
	if !crypto.VerifySignature(publickKeyBytes, data, rs) {
		return nil, nil, nil, errors.New("invalid signature produced")
	}

	v, r, s = toSignatureValues(sig)
	return v, r, s, nil
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
func toSignatureValues(sig []byte) (v, r, s *big.Int) {
	// Extract the V, R, and S values
	r = big.NewInt(0).SetBytes(sig[:32])
	s = big.NewInt(0).SetBytes(sig[32:64])
	v = big.NewInt(0).SetBytes([]byte{sig[64] + ardanID})

	return v, r, s
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

func ToSignatureBytesWithArdanID(v, r, s *big.Int) []byte {
	sig := ToSignatureBytes(v, r, s)
	sig[64] = byte(v.Uint64())

	return sig
}

func VerifySignature(v, r, s *big.Int) error {
	// check the recover id is ether 0 or 1
	uintV := v.Uint64() - ardanID
	if uintV != 0 && uintV != 1 {
		return errors.New("invalid signature recover id")
	}

	// check the signature values are in the valid range
	if !crypto.ValidateSignatureValues(byte(uintV), r, s, false) {
		return errors.New("invalid signature values")
	}

	return nil
}

func SignatureString(v, r, s *big.Int) string {
	return hexutil.Encode(ToSignatureBytesWithArdanID(v, r, s))
}
