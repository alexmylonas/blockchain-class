package nameservice

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ethereum/go-ethereum/crypto"
)

type NameService struct {
	accounts map[database.AccountID]string
}

func New(root string) (*NameService, error) {
	ns := NameService{
		accounts: make(map[database.AccountID]string),
	}

	fn := func(fileName string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path.Ext(fileName) != ".ecdsa" {
			return nil
		}
		privateKey, err := crypto.LoadECDSA(fileName)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		accountID := database.PublicKeyToAccountID(privateKey.PublicKey)
		ns.accounts[accountID] = strings.TrimSuffix(path.Base(fileName), ".ecdsa")

		return nil
	}

	if err := filepath.Walk(root, fn); err != nil {
		return nil, err
	}

	return &ns, nil
}

func (ns *NameService) Lookup(accountID database.AccountID) string {
	if name, ok := ns.accounts[accountID]; ok {
		return name
	}
	return string(accountID)
}

func (ns *NameService) Copy() map[database.AccountID]string {
	accounts := make(map[database.AccountID]string)
	for ac, name := range ns.accounts {
		accounts[ac] = name
	}
	return accounts
}
