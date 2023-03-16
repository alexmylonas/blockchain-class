package cmd

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var (
	url   string
	nonce uint64
	from  string
	to    string
	value uint64
	tip   uint64
	data  []byte
)

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a transaction",
	Run:   sendRun,
}

func init() {
	rootCmd.AddCommand(sendCmd)
	rootCmd.Flags().StringVarP(&url, "url", "u", "http://localhost:8080", "URL of the blockchain node")
	rootCmd.Flags().Uint64VarP(&nonce, "nonce", "n", 0, "Nonce of the transaction")
	rootCmd.Flags().StringVarP(&from, "from", "f", "", "From address")
	rootCmd.Flags().StringVarP(&to, "to", "t", "", "To address")
	rootCmd.Flags().Uint64VarP(&value, "value", "v", 0, "Value of the transaction")
	rootCmd.Flags().Uint64VarP(&tip, "tip", "p", 0, "Tip of the transaction")
	rootCmd.Flags().BytesHexVarP(&data, "data", "d", nil, "Data of the transaction")
}

func sendRun(cmd *cobra.Command, args []string) {
	privateKey, err := crypto.LoadECDSA(getPrivateKeyPath())
	if err != nil {
		log.Fatal(err)
	}

	sendWithDetails(privateKey)
}

func sendWithDetails(privateKey *ecdsa.PrivateKey) {
	// TODO: Get the public key from the private key.
	fromAccount, err := database.ToAccountID(from)
	if err != nil {
		log.Fatal(err)
	}
	toAccount, err := database.ToAccountID(to)
	if err != nil {
		log.Fatal(err)
	}
	const chainId = 1

	tx, err := database.NewTx(chainId, nonce, fromAccount, toAccount, value, tip, data)
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := tx.Sign(privateKey)
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(signedTx)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(url+"/v1/tx/commit", "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

}
