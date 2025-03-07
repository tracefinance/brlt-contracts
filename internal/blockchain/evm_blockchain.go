package blockchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"vault0/internal/types"
)

// EVMBlockchain implements Blockchain for EVM compatible chains
type EVMBlockchain struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	chain     Chain
}

// NewEVMBlockchain creates a new EVM blockchain client for the specified chain
func NewEVMBlockchain(chain Chain) (*EVMBlockchain, error) {
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

	return &EVMBlockchain{
		client:    client,
		rpcClient: rpcClient,
		chain:     chain,
	}, nil
}

// GetChainID implements Blockchain.GetChainID
func (c *EVMBlockchain) GetChainID(ctx context.Context) (int64, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return 0, fmt.Errorf("evm: failed to get chain ID: %w", err)
	}
	return chainID.Int64(), nil
}

// GetBalance implements Blockchain.GetBalance
func (c *EVMBlockchain) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("evm: invalid address format: %w", ErrInvalidAddress)
	}

	addr := common.HexToAddress(address)
	balance, err := c.client.BalanceAt(ctx, addr, nil) // Use nil for latest block
	if err != nil {
		return nil, fmt.Errorf("evm: failed to get balance: %w", err)
	}

	return balance, nil
}

// GetNonce implements Blockchain.GetNonce
func (c *EVMBlockchain) GetNonce(ctx context.Context, address string) (uint64, error) {
	if !common.IsHexAddress(address) {
		return 0, fmt.Errorf("evm: invalid address format: %w", ErrInvalidAddress)
	}

	nonce, err := c.client.PendingNonceAt(ctx, common.HexToAddress(address))
	if err != nil {
		return 0, fmt.Errorf("evm: failed to get nonce: %w", err)
	}

	return nonce, nil
}

// GetTransaction implements Blockchain.GetTransaction
func (c *EVMBlockchain) GetTransaction(ctx context.Context, hash string) (*types.Transaction, error) {
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
	var receipt *ethtypes.Receipt

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

	return c.convertEthereumTransactionToTransaction(tx, receipt, timestamp), nil
}

// GetTransactionReceipt implements Blockchain.GetTransactionReceipt
func (c *EVMBlockchain) GetTransactionReceipt(ctx context.Context, hash string) (*types.TransactionReceipt, error) {
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

	return &types.TransactionReceipt{
		Hash:              receipt.TxHash.Hex(),
		BlockNumber:       receipt.BlockNumber,
		Status:            receipt.Status,
		GasUsed:           receipt.GasUsed,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		LogsBloom:         receipt.Bloom.Bytes(),
		Logs:              c.convertEthereumLogsToLogs(receipt.Logs),
	}, nil
}

// EstimateGas implements Blockchain.EstimateGas
func (c *EVMBlockchain) EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error) {
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
func (c *EVMBlockchain) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("evm: failed to get gas price: %w", err)
	}
	return gasPrice, nil
}

// CallContract implements Blockchain.CallContract
func (c *EVMBlockchain) CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error) {
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

// BroadcastTransaction implements Blockchain.BroadcastTransaction
func (c *EVMBlockchain) BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error) {
	var tx ethtypes.Transaction
	if err := tx.UnmarshalBinary(signedTx); err != nil {
		return "", fmt.Errorf("evm: failed to decode transaction: %w", err)
	}

	if err := c.client.SendTransaction(ctx, &tx); err != nil {
		return "", fmt.Errorf("evm: failed to send transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}

// Close implements Blockchain.Close
func (c *EVMBlockchain) Close() {
	c.rpcClient.Close()
}

// convertEthereumLogsToLogs converts Ethereum logs to our Log model
func (c *EVMBlockchain) convertEthereumLogsToLogs(logs []*ethtypes.Log) []types.Log {
	result := make([]types.Log, len(logs))
	for i, log := range logs {
		topics := make([]string, len(log.Topics))
		for j, topic := range log.Topics {
			topics[j] = topic.Hex()
		}

		result[i] = types.Log{
			Address:         log.Address.Hex(),
			Topics:          topics,
			Data:            log.Data,
			BlockNumber:     big.NewInt(int64(log.BlockNumber)),
			TransactionHash: log.TxHash.Hex(),
			LogIndex:        log.Index,
		}
	}
	return result
}

// convertEthereumTransactionToTransaction converts an Ethereum transaction to common.Transaction
func (c *EVMBlockchain) convertEthereumTransactionToTransaction(tx *ethtypes.Transaction, receipt *ethtypes.Receipt, timestamp uint64) *types.Transaction {
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
	signer := ethtypes.LatestSignerForChainID(big.NewInt(c.chain.ID))
	if sender, err := ethtypes.Sender(signer, tx); err == nil {
		from = sender.Hex()
	}

	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	// Map chain type to common.ChainType
	var chainType types.ChainType
	switch c.chain.Type {
	case types.ChainTypeEthereum:
		chainType = types.ChainTypeEthereum
	case types.ChainTypePolygon:
		chainType = types.ChainTypePolygon
	case types.ChainTypeBase:
		chainType = types.ChainTypeBase
	default:
		chainType = types.ChainType(string(c.chain.Type))
	}

	return &types.Transaction{
		Chain:     chainType,
		Hash:      tx.Hash().Hex(),
		From:      from,
		To:        to,
		Value:     tx.Value(),
		Data:      tx.Data(),
		Nonce:     tx.Nonce(),
		GasPrice:  tx.GasPrice(),
		GasLimit:  tx.Gas(),
		Type:      types.TransactionTypeNative,
		Status:    status,
		Timestamp: int64(timestamp),
	}
}

// ChainType returns the type of the blockchain
func (c *EVMBlockchain) ChainType() types.ChainType {
	return c.chain.Type
}
