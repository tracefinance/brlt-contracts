package blockchain

import (
	"context"
	stderrors "errors"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// EVMBlockchain implements Blockchain for EVM compatible chains
type EVMBlockchain struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	chain     types.Chain
	log       logger.Logger
}

// NewEVMBlockchain creates a new EVM blockchain client
func NewEVMBlockchain(chain types.Chain, log logger.Logger) (*EVMBlockchain, error) {
	if chain.RPCUrl == "" {
		return nil, errors.NewInvalidBlockchainConfigError(string(chain.Type), "rpc_url")
	}

	// Create a new Ethereum RPC client
	rpcClient, err := rpc.Dial(chain.RPCUrl)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}

	// Create an Ethereum client from the RPC client
	client := ethclient.NewClient(rpcClient)

	// Create the EVM blockchain client
	evm := &EVMBlockchain{
		client:    client,
		rpcClient: rpcClient,
		chain:     chain,
		log:       log,
	}

	// Try to get the chain ID to verify the connection
	_, err = evm.GetChainID(context.Background())
	if err != nil {
		// Close the connections before returning
		evm.Close()
		return nil, errors.NewBlockchainError(err)
	}

	return evm, nil
}

// GetChainID implements Blockchain.GetChainID
func (c *EVMBlockchain) GetChainID(ctx context.Context) (int64, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return 0, errors.NewRPCError(err)
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
		return nil, errors.NewRPCError(err)
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
		return 0, errors.NewRPCError(err)
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
		if stderrors.Is(err, ethereum.NotFound) {
			return nil, errors.NewTransactionNotFoundError(hash)
		}
		return nil, errors.NewRPCError(err)
	}

	var blockNumber *big.Int
	var timestamp uint64
	var receipt *ethTypes.Receipt

	if !isPending {
		receipt, err = c.client.TransactionReceipt(ctx, txHash)
		if err != nil {
			if stderrors.Is(err, ethereum.NotFound) {
				return nil, errors.NewTransactionNotFoundError(hash)
			}
			return nil, errors.NewRPCError(err)
		}

		blockNumber = receipt.BlockNumber

		// Get block to get timestamp
		block, err := c.client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			return nil, errors.NewRPCError(err)
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
		if stderrors.Is(err, ethereum.NotFound) {
			return nil, errors.NewTransactionNotFoundError(hash)
		}
		return nil, errors.NewRPCError(err)
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
		return 0, errors.NewInvalidGasLimitError(gas)
	}

	return gas, nil
}

// GetGasPrice implements Blockchain.GetGasPrice
func (c *EVMBlockchain) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.NewRPCError(err)
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
		return nil, errors.NewInvalidContractCallError(to, err)
	}

	return result, nil
}

// BroadcastTransaction implements Blockchain.BroadcastTransaction
func (c *EVMBlockchain) BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error) {
	var tx ethTypes.Transaction
	if err := tx.UnmarshalBinary(signedTx); err != nil {
		return "", errors.NewInvalidTransactionError(err)
	}

	if err := c.client.SendTransaction(ctx, &tx); err != nil {
		return "", errors.NewRPCError(err)
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
		return nil, errors.NewRPCError(err)
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

func (c *EVMBlockchain) SubscribeToEvents(ctx context.Context, addresses []string, topics [][]string, fromBlock int64) (<-chan types.Log, <-chan error, error) {
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

	// Set fromBlock if provided
	if fromBlock <= 0 {
		// Get the current block number
		currentBlockNumber, err := c.client.BlockNumber(ctx)
		if err != nil {
			return nil, nil, errors.NewRPCError(err)
		}

		fromBlock = int64(math.Max(float64(currentBlockNumber-50000), 0))
	}

	// Create the filter query
	filterQuery := ethereum.FilterQuery{
		Addresses: ethAddresses,
		Topics:    ethTopics,
		FromBlock: big.NewInt(fromBlock),
	}

	// Create channels for logs and errors
	logChan := make(chan types.Log, 100) // Buffer to prevent lost events during reconnection
	errChan := make(chan error, 10)      // Buffer for error reporting

	// Create a separate cancellation context for the subscription goroutine
	subscriptionCtx, cancelSubscription := context.WithCancel(context.Background())

	// Start the subscription handler goroutine
	go func() {
		defer close(logChan)
		defer close(errChan)
		defer cancelSubscription()

		// Track the last seen block for reconnection
		var lastSeenBlock int64 = fromBlock
		// Backoff parameters
		initialBackoff := 1.0 // seconds
		maxBackoff := 60.0    // seconds
		backoff := initialBackoff

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Create a new filter query with the updated fromBlock
				currentFilterQuery := filterQuery
				currentFilterQuery.FromBlock = big.NewInt(lastSeenBlock)

				// Create a new subscription
				ethLogChan := make(chan ethTypes.Log)
				sub, err := c.client.SubscribeFilterLogs(subscriptionCtx, currentFilterQuery, ethLogChan)

				if err != nil {
					// Report the error and retry with backoff
					reconnectErr := errors.NewRPCError(err)
					select {
					case errChan <- reconnectErr:
					default:
						// Don't block if error channel is full
					}

					c.log.Warn("Subscription failed, retrying with backoff",
						logger.Float64("backoff_seconds", backoff),
						logger.Int64("from_block", lastSeenBlock),
						logger.Error(err))

					// Apply backoff before retrying
					time.Sleep(time.Duration(backoff) * time.Second)
					// Increase backoff with exponential formula, capped at maximum
					backoff = math.Min(backoff*1.5, maxBackoff)
					continue
				}

				// Reset backoff on successful connection
				backoff = initialBackoff

				// Log successful subscription
				c.log.Info("Successfully subscribed to events",
					logger.Int64("from_block", lastSeenBlock))

				// Process events from this subscription
				subscriptionActive := true
				for subscriptionActive {
					select {
					case <-ctx.Done():
						sub.Unsubscribe()
						return
					case err := <-sub.Err():
						// Report the error
						reconnectErr := errors.NewRPCError(err)
						select {
						case errChan <- reconnectErr:
						default:
							// Don't block if error channel is full
						}

						c.log.Warn("Subscription error, reconnecting",
							logger.Float64("backoff_seconds", backoff),
							logger.Int64("from_block", lastSeenBlock),
							logger.Error(err))

						// Apply backoff before reconnecting
						time.Sleep(time.Duration(backoff) * time.Second)
						// Increase backoff with exponential formula, capped at maximum
						backoff = math.Min(backoff*1.5, maxBackoff)

						// Mark subscription as inactive to break inner loop and create new subscription
						sub.Unsubscribe()
						subscriptionActive = false
					case log := <-ethLogChan:
						// Convert ethereum log to our log format
						topics := make([]string, len(log.Topics))
						for j, topic := range log.Topics {
							topics[j] = topic.Hex()
						}

						// Create our log format
						ourLog := types.Log{
							Address:         log.Address.Hex(),
							Topics:          topics,
							Data:            log.Data,
							BlockNumber:     big.NewInt(int64(log.BlockNumber)),
							TransactionHash: log.TxHash.Hex(),
							LogIndex:        log.Index,
						}

						// Update last seen block for reconnection if newer
						if ourLog.BlockNumber != nil && ourLog.BlockNumber.Int64() > lastSeenBlock {
							lastSeenBlock = ourLog.BlockNumber.Int64()
						}

						// Send to output channel
						select {
						case logChan <- ourLog:
						case <-time.After(100 * time.Millisecond):
							// If we can't send quickly, log a warning about buffer pressure
							c.log.Warn("Log channel buffer full, event processing may be delayed")
						}
					}
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
