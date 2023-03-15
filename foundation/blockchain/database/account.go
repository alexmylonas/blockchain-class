package database

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

// The last 20 bytes of the public key
type AccountID string

type Account struct {
	AccountID AccountID
	Balance   uint64
	Nonce     uint64
}

func newAccount(accountID AccountID, balance uint64) Account {
	return Account{
		AccountID: accountID,
		Balance:   balance,
	}
}

func ToAccountID(hex string) (AccountID, error) {
	a := AccountID(hex)
	if !a.IsAccountID() {
		return "", errors.New("invalid account id")
	}
	return AccountID(hex), nil
}

func PublicKeyToAccountID(pk ecdsa.PublicKey) AccountID {
	return AccountID(crypto.PubkeyToAddress(pk).String())
}

func (a AccountID) IsAccountID() bool {
	const addressLength = 20
	if a.has0xPrefix() {
		a = a[2:]
	}

	return len(a) == addressLength*2 && a.IsHex()
}

func (a AccountID) has0xPrefix() bool {
	return a[:2] == "0x"
}

func (a AccountID) IsHex() bool {
	for _, c := range a {
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			continue
		}
		return false
	}
	return true
}
