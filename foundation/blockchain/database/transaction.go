package database

import (
	"crypto/ecdsa"
	"errors"
	"math/big"

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
