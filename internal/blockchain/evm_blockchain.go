package blockchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"vault0/internal/wallet"
)

// EVMClient implements Blockchain for EVM compatible chains
type EVMClient struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	chain     Chain
}

// NewEVMClient creates a new EVM client for the specified chain
func NewEVMClient(chain Chain) (*EVMClient, error) {
	if chain.RPCUrl == "" {
		return nil, fmt.Errorf("evm: RPC URL is required: %w", ErrRPCConnectionFailed)
	}

	rpcClient, err := rpc.Dial(chain.RPCUrl)
	if err != nil {
		return nil, fmt.Errorf("evm: failed to connect to RPC: %w", err)
	}

	client := ethclient.NewClient(rpcClient)

	// Verify the connection with a chain ID check
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		rpcClient.Close()
		return nil, fmt.Errorf("evm: failed to get chain ID: %w", err)
	}

	// Verify that the chain ID matches what we expect
	if chain.ID != 0 && chain.ID != chainID.Int64() {
		rpcClient.Close()
		return nil, fmt.Errorf("evm: chain ID mismatch, expected %d, got %d: %w",
			chain.ID, chainID.Int64(), ErrChainNotSupported)
	}

	return &EVMClient{
		client:    client,
		rpcClient: rpcClient,
		chain:     chain,
	}, nil
}

// GetChainID implements Blockchain.GetChainID
func (c *EVMClient) GetChainID(ctx context.Context) (int64, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return 0, fmt.Errorf("evm: failed to get chain ID: %w", err)
	}
	return chainID.Int64(), nil
}

// GetBalance implements Blockchain.GetBalance
func (c *EVMClient) GetBalance(ctx context.Context, address string, blockNumber *big.Int) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("evm: invalid address format: %w", ErrInvalidAddress)
	}

	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("evm: failed to get balance: %w", err)
	}

	return balance, nil
}

// GetNonce implements Blockchain.GetNonce
func (c *EVMClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	if !common.IsHexAddress(address) {
		return 0, fmt.Errorf("evm: invalid address format: %w", ErrInvalidAddress)
	}

	nonce, err := c.client.PendingNonceAt(ctx, common.HexToAddress(address))
	if err != nil {
		return 0, fmt.Errorf("evm: failed to get nonce: %w", err)
	}

	return nonce, nil
}

// convertTxToTransaction converts an Ethereum transaction to wallet.Transaction
func (c *EVMClient) convertTxToTransaction(tx *types.Transaction, receipt *types.Receipt, blockNumber *big.Int, timestamp uint64) *wallet.Transaction {
	var status string
	if receipt != nil {
		if receipt.Status == 1 {
			status = "success"
		} else {
			status = "failed"
		}
	} else {
		status = "pending"
	}

	from := ""
	signer := types.LatestSignerForChainID(big.NewInt(c.chain.ID))
	if sender, err := types.Sender(signer, tx); err == nil {
		from = sender.Hex()
	}

	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	// Convert chain type
	var chainType wallet.ChainType
	switch c.chain.Type {
	case Ethereum:
		chainType = wallet.ChainType("ethereum")
	case Polygon:
		chainType = wallet.ChainType("polygon")
	case Base:
		chainType = wallet.ChainType("base")
	default:
		chainType = wallet.ChainType(string(c.chain.Type))
	}

	return &wallet.Transaction{
		Chain:     chainType,
		Hash:      tx.Hash().Hex(),
		From:      from,
		To:        to,
		Value:     tx.Value(),
		Data:      tx.Data(),
		Nonce:     tx.Nonce(),
		GasPrice:  tx.GasPrice(),
		GasLimit:  tx.Gas(),
		Type:      wallet.TransactionType("native"),
		Status:    status,
		Timestamp: int64(timestamp),
	}
}

// GetTransaction implements Blockchain.GetTransaction
func (c *EVMClient) GetTransaction(ctx context.Context, hash string) (*wallet.Transaction, error) {
	if !strings.HasPrefix(hash, "0x") {
		hash = "0x" + hash
	}

	txHash := common.HexToHash(hash)
	tx, isPending, err := c.client.TransactionByHash(ctx, txHash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, fmt.Errorf("evm: transaction not found: %w", err)
		}
		return nil, fmt.Errorf("evm: failed to get transaction: %w", err)
	}

	var blockNumber *big.Int
	var timestamp uint64
	var receipt *types.Receipt

	if !isPending {
		receipt, err = c.client.TransactionReceipt(ctx, txHash)
		if err != nil {
			return nil, fmt.Errorf("evm: failed to get transaction receipt: %w", err)
		}

		blockNumber = receipt.BlockNumber

		// Get block to get timestamp
		block, err := c.client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			return nil, fmt.Errorf("evm: failed to get block: %w", err)
		}
		timestamp = block.Time()
	}

	return c.convertTxToTransaction(tx, receipt, blockNumber, timestamp), nil
}

// convertLogsToBlockchainLogs converts EVM logs to our Log model
func (c *EVMClient) convertLogsToBlockchainLogs(logs []*types.Log) []Log {
	blockchainLogs := make([]Log, len(logs))
	for i, log := range logs {
		topics := make([]string, len(log.Topics))
		for j, topic := range log.Topics {
			topics[j] = topic.Hex()
		}

		blockchainLogs[i] = Log{
			Address:         log.Address.Hex(),
			Topics:          topics,
			Data:            log.Data,
			BlockNumber:     big.NewInt(int64(log.BlockNumber)),
			TransactionHash: log.TxHash.Hex(),
			LogIndex:        uint(log.Index),
		}
	}
	return blockchainLogs
}

// GetTransactionReceipt implements Blockchain.GetTransactionReceipt
func (c *EVMClient) GetTransactionReceipt(ctx context.Context, hash string) (*TransactionReceipt, error) {
	if !strings.HasPrefix(hash, "0x") {
		hash = "0x" + hash
	}

	receipt, err := c.client.TransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, fmt.Errorf("evm: transaction receipt not found: %w", err)
		}
		return nil, fmt.Errorf("evm: failed to get transaction receipt: %w", err)
	}

	return &TransactionReceipt{
		Hash:              receipt.TxHash.Hex(),
		BlockNumber:       receipt.BlockNumber,
		Status:            receipt.Status,
		GasUsed:           receipt.GasUsed,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		LogsBloom:         receipt.Bloom.Bytes(),
		Logs:              c.convertLogsToBlockchainLogs(receipt.Logs),
	}, nil
}

// EstimateGas implements Blockchain.EstimateGas
func (c *EVMClient) EstimateGas(ctx context.Context, tx *wallet.Transaction) (uint64, error) {
	var from common.Address
	if tx.From != "" {
		if !common.IsHexAddress(tx.From) {
			return 0, fmt.Errorf("evm: invalid from address: %w", ErrInvalidAddress)
		}
		from = common.HexToAddress(tx.From)
	}

	var to *common.Address
	if tx.To != "" {
		if !common.IsHexAddress(tx.To) {
			return 0, fmt.Errorf("evm: invalid to address: %w", ErrInvalidAddress)
		}
		addr := common.HexToAddress(tx.To)
		to = &addr
	}

	callMsg := ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: tx.Value,
		Data:  tx.Data,
	}

	gas, err := c.client.EstimateGas(ctx, callMsg)
	if err != nil {
		return 0, fmt.Errorf("evm: failed to estimate gas: %w", err)
	}

	return gas, nil
}

// GetGasPrice implements Blockchain.GetGasPrice
func (c *EVMClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("evm: failed to get gas price: %w", err)
	}
	return gasPrice, nil
}

// CallContract implements Blockchain.CallContract
func (c *EVMClient) CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error) {
	var fromAddress common.Address
	if from != "" {
		if !common.IsHexAddress(from) {
			return nil, fmt.Errorf("evm: invalid from address: %w", ErrInvalidAddress)
		}
		fromAddress = common.HexToAddress(from)
	}

	if !common.IsHexAddress(to) {
		return nil, fmt.Errorf("evm: invalid to address: %w", ErrInvalidAddress)
	}
	toAddress := common.HexToAddress(to)

	callMsg := ethereum.CallMsg{
		From: fromAddress,
		To:   &toAddress,
		Data: data,
	}

	result, err := c.client.CallContract(ctx, callMsg, nil)
	if err != nil {
		return nil, fmt.Errorf("evm: contract call failed: %w", err)
	}

	return result, nil
}

// SendTransaction implements Blockchain.SendTransaction
func (c *EVMClient) SendTransaction(ctx context.Context, rawTx []byte) (string, error) {
	var tx types.Transaction
	if err := tx.UnmarshalBinary(rawTx); err != nil {
		return "", fmt.Errorf("evm: failed to decode transaction: %w", err)
	}

	if err := c.client.SendTransaction(ctx, &tx); err != nil {
		return "", fmt.Errorf("evm: failed to send transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}

// Close implements Blockchain.Close
func (c *EVMClient) Close() {
	c.rpcClient.Close()
}
