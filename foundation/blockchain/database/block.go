package database

import (
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

var ErrChainForked = errors.New("blockchain forked, start resync")

type BlockData struct {
	Hash   string      `json:"hash"`
	Header BlockHeader `json:"block"`
	Trans  []BlockTx   `json:"tx"`
}

func NewBlockData(block Block) BlockData {
	blockData := BlockData{
		Hash:   block.Hash(),
		Header: block.Header,
		Trans:  block.MerkleTree.Values(),
	}

	return blockData
}

func ToBlock(blockData BlockData) (Block, error) {
	tree, err := merkle.NewTree(blockData.Trans)
	if err != nil {
		return Block{}, err
	}

	block := Block{
		Header:     blockData.Header,
		MerkleTree: tree,
	}
	return block, nil
}

// ================ BLOCK HEADER =================

type BlockHeader struct {
	Number        uint64    `json:"number"`          // Ethereum: Block Number in chain
	PrevBlockHash string    `json:"prev_block_hash"` // Bitcoin: Hash of previous block
	Timestamp     uint64    `json:"timestamp"`       // Bitcoin: Timestamp of block was mined
	BeneficiaryID AccountID `json:"beneficiary"`     // Ethereum: Address of miner
	Difficulty    uint16    `json:"difficulty"`      // Ethereum: Difficulty of block
	MiningReward  uint64    `json:"mining_reward"`   // Ethereum: Mining reward of block
	StateRoot     string    `json:"state_root"`      // Ethereum: State root of block
	TransRoot     string    `json:"trans_root"`      // Both: Represents the merkle tree root has for the transactions in the block
	Nonce         uint64    `json:"nonce"`           // Both: Value identified to solve the hash of the block
}

type Block struct {
	Header     BlockHeader
	MerkleTree *merkle.Tree[BlockTx]
}

func (b *Block) Hash() string {
	if b.Header.Number == 0 {
		return signature.ZeroHash
	}

	// CORE NOTE: Hashing the block header and not the whole block so the blockchain
	// can be cryptographically check by only needding the block header and not the full
	// blocks with the transactions data. This will support the ability to have pruned nodes
	// and light clients in the future.
	// - A pruned node stores all the block headers but only a small number of full block
	//   maybe the last 100 blocks. This allows for full cryptographic verification of the
	//   blocks and transactions without all the extra storage.
	// - A light client keeps blocks headers and just sufficient information to follow the
	//   the latest set of block being product. The DO NOT validate blocks but can prove a trasaction
	//   was included in a block.

	return signature.Hash(b.Header)
}
