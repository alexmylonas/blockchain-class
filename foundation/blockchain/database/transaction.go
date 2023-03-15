package database

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

type Tx struct {
	ChainID uint16    `json:"chain_id"`
	Nonce   uint64    `json:"nonce"`
	FromID  AccountID `json:"from"`
	ToID    AccountID `json:"to"`
	Value   uint64    `json:"value"`
	Tip     uint64    `json:"tip"`
	Data    []byte    `json:"data"`
}

func NewTx(chainID uint16, nonce uint64, from, to AccountID, value, tip uint64, data []byte) (Tx, error) {
	if !from.IsAccountID() {
		return Tx{}, errors.New("invalid from account id")
	}
	if !to.IsAccountID() {
		return Tx{}, errors.New("invalid to account id")
	}
	return Tx{
		ChainID: chainID,
		Nonce:   nonce,
		FromID:  from,
		ToID:    to,
		Value:   value,
		Tip:     tip,
		Data:    data,
	}, nil
}

// ==============================

type SignedTx struct {
	Tx
	V *big.Int `json:"v"` // Ethereum Recovery Identifier, either 29 or 30 for ardan chain
	R *big.Int `json:"r"` // Ethereum: First coordinate of the ECDSA signature
	S *big.Int `json:"s"` // Ethereum: Second coordinate of the ECDSA signature
}

func (tx Tx) Sign(privateKey *ecdsa.PrivateKey) (SignedTx, error) {

	v, r, s, err := signature.Sign(tx, privateKey)
	if err != nil {
		return SignedTx{}, err
	}

	return SignedTx{
		Tx: tx,
		V:  v,
		R:  r,
		S:  s,
	}, nil
}

func (tx SignedTx) Validate(chainID uint16) error {
	if tx.ChainID != chainID {
		return errors.New("invalid chain id")
	}

	if !tx.FromID.IsAccountID() {
		return errors.New("invalid from account id")
	}

	if !tx.ToID.IsAccountID() {
		return errors.New("invalid to account id")
	}

	if tx.FromID == tx.ToID {
		return errors.New("from and to account ids are the same")
	}

	if err := signature.VerifySignature(tx.V, tx.R, tx.S); err != nil {
		return err
	}

	address, err := signature.FromAddress(tx.Tx, tx.V, tx.R, tx.S)
	if err != nil {
		return err
	}

	if address != string(tx.FromID) {
		return errors.New("signature does not match from account id")
	}
	return nil
}

// SignatureString returns the signature as a string.
func (tx SignedTx) SignatureString() string {
	return signature.SignatureString(tx.V, tx.R, tx.S)
}

// String returns the transaction as a string.
func (tx SignedTx) String() string {
	return fmt.Sprintf("%s:%d", tx.FromID, tx.Nonce)
}

// ======== BLOCK ========
// BlockTx is a transaction that has been added to a block.
type BlockTx struct {
	SignedTx
	TimeStamp uint64 `json:"timestamp"`
	GasPrice  uint64 `json:"gas_price"`
	GasUnits  uint64 `json:"gas_units"`
}

func NewBlockTx(signedTx SignedTx, gasPrice, gasUnits uint64) BlockTx {
	return BlockTx{
		SignedTx:  signedTx,
		TimeStamp: uint64(time.Now().Unix()),
		GasPrice:  gasPrice,
		GasUnits:  gasUnits,
	}
}

// Hash implementes the merkle Hashbable interface for providin a hash
// for the BlockTx.
func (tx BlockTx) Hash() ([]byte, error) {
	str := signature.Hash(tx)

	// Need to remove the 0x prefix.
	return hex.DecodeString(str[2:])
}

// Equals returns true if the two transactions are equal. If the nonce and the signatures
// are equal, then the transactions are equal.
func (tx BlockTx) Equals(otherTx BlockTx) bool {
	txSig := signature.ToSignatureBytes(tx.V, tx.R, tx.S)
	otherTxSig := signature.ToSignatureBytes(otherTx.V, otherTx.R, otherTx.S)
	return bytes.Equal(txSig, otherTxSig) && tx.Nonce == otherTx.Nonce
}
