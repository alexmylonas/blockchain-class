package database

import (
	"context"
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

var ErrChainForked = errors.New("blockchain forked, start resync")
var ErrInvalidDifficulty = errors.New("invalid difficulty")
var ErrInvalidHash = errors.New("invalid hash")
var ErrInvalidBlockNumber = errors.New("invalid block number")
var ErrInvalidPrevBlockHash = errors.New("invalid previous block hash")
var ErrInvalidBlockTimestamp = errors.New("invalid block timestamp")
var ErrInvalidStateRoot = errors.New("invalid state root")
var ErrInvalidTransRoot = errors.New("invalid transaction root")

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

type POWArgs struct {
	BeneficiaryID AccountID
	Difficulty    uint16
	MiningReward  uint64
	PrevBlock     Block
	StateRoot     string
	Trans         []BlockTx
	EvHandler     func(v string, args ...any)
}

func POW(ctx context.Context, args POWArgs) (Block, error) {
	// When mining the first block, the previous block hash is the zero hash.
	prevBlockHash := signature.ZeroHash
	if args.PrevBlock.Header.Number > 0 {
		prevBlockHash = args.PrevBlock.Hash()
	}

	// Consturct a merkle tree
	tree, err := merkle.NewTree(args.Trans)
	if err != nil {
		return Block{}, err
	}

	// Construct the block header
	header := BlockHeader{
		Number:        args.PrevBlock.Header.Number + 1,
		PrevBlockHash: prevBlockHash,
		Timestamp:     uint64(time.Now().UTC().UnixMilli()),
		BeneficiaryID: args.BeneficiaryID,
		Difficulty:    args.Difficulty,
		MiningReward:  args.MiningReward,
		StateRoot:     args.StateRoot,
		TransRoot:     tree.RootHex(),
		Nonce:         0,
	}
	// Create the block
	block := Block{
		Header:     header,
		MerkleTree: tree,
	}

	if err := block.performPOW(ctx, args.EvHandler); err != nil {
		return Block{}, err
	}

	return block, nil
}

func (b *Block) performPOW(ctx context.Context, ev func(v string, args ...any)) error {
	ev("database:performPOW:started")
	defer ev("database:performPOW:completed")

	for _, tx := range b.MerkleTree.Values() {
		ev("database:performPOW:tx [%s]", tx)
	}

	nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return err
	}
	b.Header.Nonce = nBig.Uint64()

	ev("database: PerformPOW ")

	var attemps uint64
	for {
		attemps++
		if attemps%1000000 == 0 {
			ev("database: PerformPOW for Attemps [%d]", attemps)
		}

		if ctx.Err() != nil {
			ev("database: PerformPOW: Mining cancelled")
			return ctx.Err()
		}

		hash := b.Hash()
		if !isHashSolved(b.Header.Difficulty, hash) {
			nBig, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
			if err != nil {
				return err
			}
			b.Header.Nonce = nBig.Uint64()
			// Regenerating a random number seems faster than incrementing a number
			// b.Header.Nonce++
			continue
		}
		// We have a solved hash

		ev("database: PerformPOW: Solved prevBlk[%s], newBlk[%s]", b.Header.PrevBlockHash, hash)
		ev("database: PerformPOW: Attempts [%d]", attemps)

		return nil
	}
}

func (b *Block) ValidateBlock(previousBlock Block, stateRoot string, evHandler func(v string, args ...any)) error {
	evHandler("database: ValidateBlock: blk[%d]: check: chain is not forked", b.Header.Number)

	// The node who sent this block has a chain that is two or more blocks ahead of us.
	nextNumber := previousBlock.Header.Number + 1
	if b.Header.Number >= nextNumber+2 {
		return ErrChainForked
	}

	evHandler("database: ValidateBlock: blk[%d]: check: block number is correct", b.Header.Number)

	if b.Header.Difficulty < previousBlock.Header.Difficulty {
		return ErrInvalidDifficulty
	}

	evHandler("database: ValidateBlock: blk[%d]: check: block difficulty is correct", b.Header.Number)

	hash := b.Hash()
	if !isHashSolved(b.Header.Difficulty, hash) {
		return ErrInvalidHash
	}

	evHandler("database: ValidateBlock: blk[%d]: check: block hash is correct", b.Header.Number)

	if b.Header.Number != nextNumber {
		return ErrInvalidBlockNumber
	}

	evHandler("database: ValidateBlock: blk[%d]: check: block number is correct", b.Header.Number)

	if b.Header.PrevBlockHash != previousBlock.Hash() {
		return ErrInvalidPrevBlockHash
	}

	if previousBlock.Header.Timestamp > 0 {
		parentTime := time.Unix(0, int64(previousBlock.Header.Timestamp))
		blockTime := time.Unix(0, int64(b.Header.Timestamp))

		if blockTime.Before(parentTime) {
			return ErrInvalidBlockTimestamp
		}

		evHandler("database: ValidateBlock: blk[%d]: check: block timestamp is correct", b.Header.Number)
	}

	if b.Header.StateRoot != stateRoot {
		return ErrInvalidStateRoot
	}

	evHandler("database: ValidateBlock: blk[%d]: check: block state root is correct", b.Header.Number)

	if b.Header.TransRoot != b.MerkleTree.RootHex() {
		return ErrInvalidTransRoot
	}

	evHandler("database: ValidateBlock: blk[%d]: check: trans root is correct", b.Header.Number)

	return nil
}

func isHashSolved(difficulty uint16, hash string) bool {
	const match = "0x00000000000000000"

	if len(hash) != 66 {
		return false
	}

	difficulty += 2

	return hash[:difficulty] == match[:difficulty]
}
