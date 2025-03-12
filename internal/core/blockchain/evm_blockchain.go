package blockchain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"vault0/internal/types"
)

// EVMBlockchain implements Blockchain for EVM compatible chains
type EVMBlockchain struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	chain     types.Chain
}

// NewEVMBlockchain creates a new EVM blockchain client
func NewEVMBlockchain(chain types.Chain) (*EVMBlockchain, error) {
	if chain.RPCUrl == "" {
		return nil, fmt.Errorf("RPC URL is required for %s", chain.Name)
	}

	// Create a new Ethereum RPC client
	rpcClient, err := rpc.Dial(chain.RPCUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC endpoint: %w", err)
	}

	// Create an Ethereum client from the RPC client
	client := ethclient.NewClient(rpcClient)

	// Create the EVM blockchain client
	evm := &EVMBlockchain{
		client:    client,
		rpcClient: rpcClient,
		chain:     chain,
	}

	// Try to get the chain ID to verify the connection
	_, err = evm.GetChainID(context.Background())
	if err != nil {
		// Close the connections before returning
		evm.Close()
		return nil, fmt.Errorf("evm: failed to get chain ID: %w", err)
	}

	return evm, nil
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
	if err := c.chain.ValidateAddress(address); err != nil {
		return nil, err
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
	if err := c.chain.ValidateAddress(address); err != nil {
		return 0, err
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
	var receipt *ethTypes.Receipt

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
		if err := c.chain.ValidateAddress(tx.From); err != nil {
			return 0, err
		}
		from = common.HexToAddress(tx.From)
	}

	var to *common.Address
	if tx.To != "" {
		if err := c.chain.ValidateAddress(tx.To); err != nil {
			return 0, err
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
	if from != "" && from != types.ZeroAddress {
		if err := c.chain.ValidateAddress(from); err != nil {
			return nil, err
		}
		fromAddress = common.HexToAddress(from)
	}

	if err := c.chain.ValidateAddress(to); err != nil {
		return nil, err
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
	var tx ethTypes.Transaction
	if err := tx.UnmarshalBinary(signedTx); err != nil {
		return "", fmt.Errorf("evm: failed to decode transaction: %w", err)
	}

	if err := c.client.SendTransaction(ctx, &tx); err != nil {
		return "", fmt.Errorf("evm: failed to send transaction: %w", err)
	}

	return tx.Hash().Hex(), nil
}

// FilterLogs implements Blockchain.FilterLogs
func (c *EVMBlockchain) FilterLogs(ctx context.Context, addresses []string, topics [][]string, fromBlock, toBlock int64) ([]types.Log, error) {
	// Convert addresses to Ethereum addresses
	var ethAddresses []common.Address
	if len(addresses) > 0 {
		ethAddresses = make([]common.Address, len(addresses))
		for i, address := range addresses {
			if err := c.chain.ValidateAddress(address); err != nil {
				return nil, err
			}
			ethAddresses[i] = common.HexToAddress(address)
		}
	}

	// Convert topics to Ethereum topics
	var ethTopics [][]common.Hash
	if len(topics) > 0 {
		ethTopics = make([][]common.Hash, len(topics))
		for i, topicSet := range topics {
			if len(topicSet) > 0 {
				ethTopics[i] = make([]common.Hash, len(topicSet))
				for j, topic := range topicSet {
					ethTopics[i][j] = common.HexToHash(topic)
				}
			}
		}
	}

	// Create the filter query
	filterQuery := ethereum.FilterQuery{
		Addresses: ethAddresses,
		Topics:    ethTopics,
	}

	// Set block ranges if provided
	if fromBlock >= 0 {
		filterQuery.FromBlock = big.NewInt(fromBlock)
	}
	if toBlock >= 0 {
		filterQuery.ToBlock = big.NewInt(toBlock)
	}

	// Filter logs
	logs, err := c.client.FilterLogs(ctx, filterQuery)
	if err != nil {
		return nil, fmt.Errorf("evm: failed to filter logs: %w", err)
	}

	// Convert ethereum logs to our log format
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

	return result, nil
}

// SubscribeToEvents implements Blockchain.SubscribeToEvents
func (c *EVMBlockchain) SubscribeToEvents(ctx context.Context, addresses []string, topics [][]string) (<-chan types.Log, <-chan error, error) {
	// Convert addresses to Ethereum addresses
	var ethAddresses []common.Address
	if len(addresses) > 0 {
		ethAddresses = make([]common.Address, len(addresses))
		for i, address := range addresses {
			if err := c.chain.ValidateAddress(address); err != nil {
				return nil, nil, err
			}
			ethAddresses[i] = common.HexToAddress(address)
		}
	}

	// Convert topics to Ethereum topics
	var ethTopics [][]common.Hash
	if len(topics) > 0 {
		ethTopics = make([][]common.Hash, len(topics))
		for i, topicSet := range topics {
			if len(topicSet) > 0 {
				ethTopics[i] = make([]common.Hash, len(topicSet))
				for j, topic := range topicSet {
					ethTopics[i][j] = common.HexToHash(topic)
				}
			}
		}
	}

	// Create the filter query
	filterQuery := ethereum.FilterQuery{
		Addresses: ethAddresses,
		Topics:    ethTopics,
	}

	// Create channels for logs and errors
	logChan := make(chan types.Log)
	errChan := make(chan error)

	// Subscribe to logs
	ethLogChan := make(chan ethTypes.Log)
	sub, err := c.client.SubscribeFilterLogs(ctx, filterQuery, ethLogChan)
	if err != nil {
		return nil, nil, fmt.Errorf("evm: failed to subscribe to logs: %w", err)
	}

	// Handle the subscription in a goroutine
	go func() {
		defer close(logChan)
		defer close(errChan)
		defer sub.Unsubscribe()

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-sub.Err():
				errChan <- fmt.Errorf("evm: subscription error: %w", err)
				return
			case log := <-ethLogChan:
				// Convert ethereum log to our log format
				topics := make([]string, len(log.Topics))
				for j, topic := range log.Topics {
					topics[j] = topic.Hex()
				}

				logChan <- types.Log{
					Address:         log.Address.Hex(),
					Topics:          topics,
					Data:            log.Data,
					BlockNumber:     big.NewInt(int64(log.BlockNumber)),
					TransactionHash: log.TxHash.Hex(),
					LogIndex:        log.Index,
				}
			}
		}
	}()

	return logChan, errChan, nil
}

// Close implements Blockchain.Close
func (c *EVMBlockchain) Close() {
	c.rpcClient.Close()
}

// convertEthereumLogsToLogs converts Ethereum logs to our Log model
func (c *EVMBlockchain) convertEthereumLogsToLogs(logs []*ethTypes.Log) []types.Log {
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
func (c *EVMBlockchain) convertEthereumTransactionToTransaction(tx *ethTypes.Transaction, receipt *ethTypes.Receipt, timestamp uint64) *types.Transaction {
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
	signer := ethTypes.LatestSignerForChainID(big.NewInt(c.chain.ID))
	if sender, err := ethTypes.Sender(signer, tx); err == nil {
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

// Chain implements Blockchain.Chain
func (c *EVMBlockchain) Chain() types.Chain {
	return c.chain
}
