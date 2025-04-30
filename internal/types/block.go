package types

import (
	"math/big"
	"time"
)

// Block represents a block in a blockchain.
// It contains common fields that are relevant across different blockchain implementations.
type Block struct {
	// ChainType is the type of blockchain this block belongs to
	ChainType ChainType
	// Hash is the block hash
	Hash string

	// Number is the block number/height
	Number *big.Int

	// ParentHash is the hash of the parent block
	ParentHash string

	// Timestamp is the block creation time
	Timestamp time.Time

	// TransactionCount is the number of transactions in the block
	TransactionCount int

	// Transactions is the list of full transaction objects in the block
	Transactions []*Transaction

	// Miner is the address of the miner/validator who produced the block
	Miner string

	// GasUsed is the total gas used by all transactions in the block
	GasUsed uint64

	// GasLimit is the maximum gas allowed in this block
	GasLimit uint64

	// Size is the size of the block in bytes
	Size uint64

	// Difficulty is the block's difficulty (for PoW chains)
	Difficulty *big.Int

	// Extra is any additional data included in the block
	Extra []byte
}
