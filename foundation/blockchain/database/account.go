package database

import "errors"

// The last 20 bytes of the public key
type AccountID string

func ToAccountID(hex string) (AccountID, error) {
	a := AccountID(hex)
	if !a.IsAccountID() {
		return "", errors.New("invalid account id")
	}
	return AccountID(hex), nil
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
