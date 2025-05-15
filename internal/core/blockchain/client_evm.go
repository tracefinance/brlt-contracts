package blockchain

import (
	"context"
	stderrors "errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

const (
	// Subscription configuration
	subscriptionLogBufferSize = 1000  // Buffer size for the log channel
	subscriptionErrBufferSize = 10    // Buffer size for the error channel
	subscriptionBlockLookback = 50000 // Number of blocks to look back for historical events

	// Backoff configuration
	subscriptionInitialBackoff = 1.0  // Initial backoff in seconds
	subscriptionMaxBackoff     = 60.0 // Maximum backoff in seconds
	subscriptionBackoffFactor  = 1.5  // Factor to increase backoff

	// Channel operation timeouts
	subscriptionChannelTimeout = 100 * time.Millisecond // Timeout for channel operations

	// Retry configuration for block fetching
	blockFetchMaxRetries    = 3                      // Maximum number of retry attempts
	blockFetchInitialDelay  = 500 * time.Millisecond // Initial delay before retrying
	blockFetchBackoffFactor = 2                      // Multiplication factor for backoff

	// Error messages for retry conditions
	errMsgNotFound          = "not found"                      // Error message for block not found
	errMsgUnsupportedTxType = "transaction type not supported" // Error message for unsupported tx type

	// Operation names
	operationFetchBlock = "fetch block" // Operation name for block fetching
)

// EVMClient implements Blockchain for EVM compatible chains
type EVMClient struct {
	client    *ethclient.Client
	rpcClient *rpc.Client
	chain     types.Chain
	log       logger.Logger
}

// NewEVMBlockchainClient creates a new EVM blockchain client
func NewEVMBlockchainClient(chain types.Chain, log logger.Logger) (*EVMClient, error) {
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
	evm := &EVMClient{
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
func (c *EVMClient) GetChainID(ctx context.Context) (int64, error) {
	chainID, err := c.client.ChainID(ctx)
	if err != nil {
		return 0, errors.NewRPCError(err)
	}
	return chainID.Int64(), nil
}

// GetBalance implements Blockchain.GetBalance
func (c *EVMClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
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

// GetTokenBalance implements Blockchain.GetTokenBalance
func (c *EVMClient) GetTokenBalance(ctx context.Context, address string, tokenAddress string) (*big.Int, error) {
	// Validate both addresses
	if err := c.chain.ValidateAddress(address); err != nil {
		return nil, err
	}
	if err := c.chain.ValidateAddress(tokenAddress); err != nil {
		return nil, err
	}

	// ERC20 balanceOf function signature: balanceOf(address)
	// Function selector is the first 4 bytes of keccak256("balanceOf(address)")
	// 0x70a08231 is the function selector for balanceOf(address)
	methodID := crypto.Keccak256([]byte(string(types.ERC20BalanceOfMethod)))[:4]

	// Encode the address parameter - EVM addresses are padded to 32 bytes
	paddedAddress := common.LeftPadBytes(common.HexToAddress(address).Bytes(), 32)

	// Combine the function selector and the padded address parameter
	data := append(methodID, paddedAddress...)

	// Call the token contract
	result, err := c.CallContract(ctx, types.ZeroAddress, tokenAddress, data)
	if err != nil {
		return nil, errors.NewInvalidTokenBalanceError(tokenAddress, err)
	}

	// The result is a 32-byte big-endian integer
	if len(result) < 32 {
		return nil, errors.NewInvalidTokenBalanceError(tokenAddress,
			fmt.Errorf("invalid response length: got %d, want 32", len(result)))
	}

	// Parse the result as a big.Int
	balance := new(big.Int).SetBytes(result)

	return balance, nil
}

// GetNonce implements Blockchain.GetNonce
func (c *EVMClient) GetNonce(ctx context.Context, address string) (uint64, error) {
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
func (c *EVMClient) GetTransaction(ctx context.Context, hash string) (*types.Transaction, error) {
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
		var block *ethTypes.Block
		block, err = c.client.BlockByNumber(ctx, blockNumber)
		if err != nil {
			return nil, errors.NewRPCError(err)
		}
		timestamp = block.Time()
	}

	return c.convertEthereumTransactionToTransaction(tx, receipt, timestamp), nil
}

// GetBlock implements Blockchain.GetBlock
func (c *EVMClient) GetBlock(ctx context.Context, identifier string) (*types.Block, error) {
	var fetchBlockFn func() (any, error)

	// Check if the identifier is a special keyword
	switch strings.ToLower(identifier) {
	case "latest":
		fetchBlockFn = func() (any, error) {
			return c.client.BlockByNumber(ctx, nil) // nil means latest block
		}
	case "earliest":
		fetchBlockFn = func() (any, error) {
			return c.client.BlockByNumber(ctx, big.NewInt(0)) // 0 is the genesis block
		}
	case "pending":
		fetchBlockFn = func() (any, error) {
			return c.client.BlockByNumber(ctx, big.NewInt(-1)) // -1 is pending block
		}
	default:
		// Check if identifier is a hash (0x...) or a block number
		if strings.HasPrefix(identifier, "0x") {
			blockHash := common.HexToHash(identifier)
			fetchBlockFn = func() (any, error) {
				return c.client.BlockByHash(ctx, blockHash)
			}
		} else {
			// Try to parse it as a block number
			blockNum, parseErr := strconv.ParseInt(identifier, 10, 64)
			if parseErr != nil {
				return nil, errors.NewInvalidBlockIdentifierError(identifier)
			}
			fetchBlockFn = func() (any, error) {
				return c.client.BlockByNumber(ctx, big.NewInt(blockNum))
			}
		}
	}

	// Context information for logging
	contextInfo := map[string]any{
		"identifier": identifier,
	}

	// Use retry operation with the fetch function
	result, err := c.retryOperation(
		operationFetchBlock,
		contextInfo,
		fetchBlockFn,
		func(err error) bool {
			return strings.Contains(err.Error(), errMsgNotFound) ||
				strings.Contains(err.Error(), errMsgUnsupportedTxType)
		},
	)

	if err != nil {
		if strings.Contains(err.Error(), errMsgNotFound) {
			return nil, errors.NewBlockNotFoundError(identifier)
		}
		return nil, errors.NewRPCError(err)
	}

	block, ok := result.(*ethTypes.Block)
	if !ok || block == nil {
		return nil, errors.NewBlockNotFoundError(identifier)
	}

	// Convert the Ethereum block to our Block model
	return c.convertEthereumBlockToBlock(block), nil
}

// GetTransactionReceipt implements Blockchain.GetTransactionReceipt
func (c *EVMClient) GetTransactionReceipt(ctx context.Context, hash string) (*types.TransactionReceipt, error) {
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

	var contractAddressPtr *string
	if receipt.ContractAddress != (common.Address{}) {
		addr := receipt.ContractAddress.Hex()
		contractAddressPtr = &addr
	}

	return &types.TransactionReceipt{
		Hash:              receipt.TxHash.Hex(),
		ContractAddress:   contractAddressPtr,
		ChainType:         c.chain.Type,
		BlockNumber:       receipt.BlockNumber,
		Status:            receipt.Status,
		GasUsed:           receipt.GasUsed,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		LogsBloom:         receipt.Bloom.Bytes(),
		Logs:              c.convertEthereumLogsToLogs(receipt.Logs),
	}, nil
}

// EstimateGas implements Blockchain.EstimateGas
func (c *EVMClient) EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error) {
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
func (c *EVMClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	gasPrice, err := c.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, errors.NewRPCError(err)
	}
	return gasPrice, nil
}

// CallContract implements Blockchain.CallContract
func (c *EVMClient) CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error) {
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
func (c *EVMClient) BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error) {
	var tx ethTypes.Transaction
	if err := tx.UnmarshalBinary(signedTx); err != nil {
		return "", errors.NewInvalidTransactionError(err)
	}

	if err := c.client.SendTransaction(ctx, &tx); err != nil {
		return "", errors.NewRPCError(err)
	}

	return tx.Hash().Hex(), nil
}

// FilterContractLogs implements Blockchain.FilterContractLogs
func (c *EVMClient) FilterContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock, toBlock int64) ([]types.Log, error) {
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

	// Convert event signature and args to topics
	ethTopics, err := c.convertEventToTopics(eventSignature, eventArgs)
	if err != nil {
		return nil, err
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
			ChainType:       c.chain.Type,
			Topics:          topics,
			Data:            log.Data,
			BlockNumber:     big.NewInt(int64(log.BlockNumber)),
			TransactionHash: log.TxHash.Hex(),
			LogIndex:        log.Index,
		}
	}

	return result, nil
}

func (c *EVMClient) SubscribeContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock int64) (<-chan types.Log, <-chan error, error) {
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

	// Convert event signature and args to topics
	ethTopics, err := c.convertEventToTopics(eventSignature, eventArgs)
	if err != nil {
		return nil, nil, err
	}

	// Set fromBlock if provided
	initialFromBlock := fromBlock
	if fromBlock <= 0 {
		// Get the current block number
		currentBlockNumber, err := c.client.BlockNumber(ctx)
		if err != nil {
			return nil, nil, errors.NewRPCError(err)
		}

		initialFromBlock = int64(math.Max(float64(currentBlockNumber-subscriptionBlockLookback), 0))
	}

	// Create channels for logs and errors
	logChan := make(chan types.Log, subscriptionLogBufferSize)
	errChan := make(chan error, subscriptionErrBufferSize)

	// Start the subscription handler goroutine
	go func() {
		defer close(logChan)
		defer close(errChan)

		c.handleSubscription(
			ctx,
			"contract events",
			initialFromBlock,
			func(subscriptionCtx context.Context, lastSeenBlock int64) (ethereum.Subscription, any, error) {
				// Create the filter query with the current last seen block
				filterQuery := ethereum.FilterQuery{
					Addresses: ethAddresses,
					Topics:    ethTopics,
					FromBlock: big.NewInt(lastSeenBlock),
				}

				// Create a new subscription
				ethLogChan := make(chan ethTypes.Log)
				sub, err := c.client.SubscribeFilterLogs(subscriptionCtx, filterQuery, ethLogChan)

				// We need to convert the specific channel to an any type for the generic handler
				return sub, any(ethLogChan), err
			},
			func(item any, _ int64) (int64, bool) {
				// Process the log item
				log := item.(ethTypes.Log)

				// Convert ethereum log to our log format
				topics := make([]string, len(log.Topics))
				for j, topic := range log.Topics {
					topics[j] = topic.Hex()
				}

				// Create our log format
				ourLog := types.Log{
					Address:         log.Address.Hex(),
					ChainType:       c.chain.Type,
					Topics:          topics,
					Data:            log.Data,
					BlockNumber:     big.NewInt(int64(log.BlockNumber)),
					TransactionHash: log.TxHash.Hex(),
					LogIndex:        log.Index,
				}

				// Send to output channel
				select {
				case logChan <- ourLog:
					// Return block number for tracking
					if ourLog.BlockNumber != nil {
						return ourLog.BlockNumber.Int64(), true
					}
					return 0, true
				case <-time.After(subscriptionChannelTimeout):
					// If we can't send quickly, log a warning about buffer pressure
					c.log.Warn("Log channel buffer full, event processing may be delayed")
					return 0, false
				}
			},
			errChan,
		)
	}()

	return logChan, errChan, nil
}

// SubscribeNewHead implements Blockchain.SubscribeNewHead
func (c *EVMClient) SubscribeNewHead(ctx context.Context) (<-chan types.Block, <-chan error, error) {
	// Create channels for blocks and errors
	blockChan := make(chan types.Block, subscriptionLogBufferSize)
	errChan := make(chan error, subscriptionErrBufferSize)

	// Start the subscription handler goroutine
	go func() {
		defer close(blockChan)
		defer close(errChan)

		c.handleSubscription(
			ctx,
			"block headers",
			0, // fromBlock is not used for header subscriptions
			func(subscriptionCtx context.Context, _ int64) (ethereum.Subscription, any, error) {
				// Create a new subscription for headers
				headers := make(chan *ethTypes.Header)
				sub, err := c.client.SubscribeNewHead(subscriptionCtx, headers)

				// We need to convert the specific channel to an any type for the generic handler
				return sub, any(headers), err
			},
			func(item any, _ int64) (int64, bool) {
				// Process the header item
				header := item.(*ethTypes.Header)

				// Context information for logging
				contextInfo := map[string]any{
					"hash": header.Hash().Hex(),
				}

				// Retry the operation
				result, err := c.retryOperation(
					operationFetchBlock,
					contextInfo,
					func() (any, error) {
						return c.client.BlockByHash(context.Background(), header.Hash())
					},
					func(err error) bool {
						return strings.Contains(err.Error(), errMsgNotFound) ||
							strings.Contains(err.Error(), errMsgUnsupportedTxType)
					},
				)

				// Handle the result or error
				if err != nil {
					// Log the error
					c.log.Warn("Failed to fetch full block details for header after retries",
						logger.String("hash", header.Hash().Hex()),
						logger.Int64("block_number", header.Number.Int64()),
						logger.String("chain", string(c.chain.Type)),
						logger.Error(err))

					// Return a partial block with just header information
					headerOnlyBlock := &types.Block{
						Hash:       header.Hash().Hex(),
						Number:     header.Number,
						ParentHash: header.ParentHash.Hex(),
						Timestamp:  time.Unix(int64(header.Time), 0),
						// We don't have transaction info
						TransactionCount: 0,
						Transactions:     nil,
						Miner:            header.Coinbase.Hex(),
						GasUsed:          header.GasUsed,
						GasLimit:         header.GasLimit,
						// Other fields based on header
						Size:       0, // Unknown
						Difficulty: header.Difficulty,
						Extra:      header.Extra,
					}

					// Send partial block to output channel
					select {
					case blockChan <- *headerOnlyBlock:
						// Return block number for tracking
						if headerOnlyBlock.Number != nil {
							return headerOnlyBlock.Number.Int64(), true
						}
						return 0, true
					case <-time.After(subscriptionChannelTimeout):
						// If we can't send quickly, log a warning about buffer pressure
						c.log.Warn("Block channel buffer full, event processing may be delayed")
						return 0, false
					}
				}

				// Cast the result to the expected type
				block := result.(*ethTypes.Block)

				// Convert ethereum block to our block model
				ourBlock := c.convertEthereumBlockToBlock(block)

				// Send to output channel
				select {
				case blockChan <- *ourBlock:
					// Return block number for tracking
					if ourBlock.Number != nil {
						return ourBlock.Number.Int64(), true
					}
					return 0, true
				case <-time.After(subscriptionChannelTimeout):
					// If we can't send quickly, log a warning about buffer pressure
					c.log.Warn("Block channel buffer full, event processing may be delayed")
					return 0, false
				}
			},
			errChan,
		)
	}()

	return blockChan, errChan, nil
}

// handleSubscription is a generic function that handles subscription creation, reconnection,
// and error handling for both SubscribeContractLogs and SubscribeNewHead.
//
// Parameters:
//   - ctx: The context for subscription lifetime
//   - name: Name of the subscription for logging
//   - createSubscription: Function that creates the actual subscription
//   - processItem: Function that processes an item received from the subscription
//   - errChan: Channel to report errors to the caller
//
// The createSubscription function should:
//   - Create a subscription and return it along with a channel for receiving items
//   - Return an error if subscription creation fails
//
// The processItem function should:
//   - Process an item from the subscription and return true if processing was successful
//   - Return the latest block number for reconnection tracking (if applicable)
func (c *EVMClient) handleSubscription(
	ctx context.Context,
	name string,
	initialFromBlock int64,
	createSubscription func(context.Context, int64) (ethereum.Subscription, any, error),
	processItem func(item any, lastSeenBlock int64) (int64, bool),
	errChan chan<- error,
) {
	// Create a separate cancellation context for the subscription goroutine
	subscriptionCtx, cancelSubscription := context.WithCancel(context.Background())
	defer cancelSubscription()

	// Track the last seen block for reconnection
	var lastSeenBlock int64 = initialFromBlock
	// Backoff parameters
	backoff := subscriptionInitialBackoff

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Create a new subscription
			sub, itemChan, err := createSubscription(subscriptionCtx, lastSeenBlock)

			if err != nil {
				// Report the error and retry with backoff
				reconnectErr := errors.NewRPCError(err)
				select {
				case errChan <- reconnectErr:
				default:
					// Don't block if error channel is full
				}

				c.log.Warn(fmt.Sprintf("%s subscription failed, retrying with backoff", name),
					logger.Float64("backoff_seconds", backoff),
					logger.Int64("from_block", lastSeenBlock),
					logger.Error(err))

				// Apply backoff before retrying
				time.Sleep(time.Duration(backoff) * time.Second)
				// Increase backoff with exponential formula, capped at maximum
				backoff = math.Min(backoff*subscriptionBackoffFactor, subscriptionMaxBackoff)
				continue
			}

			// Reset backoff on successful connection
			backoff = subscriptionInitialBackoff

			// Log successful subscription
			c.log.Info(fmt.Sprintf("Successfully subscribed to %s", name),
				logger.Int64("from_block", lastSeenBlock),
				logger.String("chain", string(c.chain.Type)))

			// Create a reflection-based channel reader that works with any channel type
			// This approach allows us to handle different channel types (ethTypes.Header, ethTypes.Log)
			chanValue := reflect.ValueOf(itemChan)
			cases := []reflect.SelectCase{
				{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(subscriptionCtx.Done())},
				{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(sub.Err())},
				{Dir: reflect.SelectRecv, Chan: chanValue},
			}

			// Process items from this subscription
			subscriptionActive := true
			for subscriptionActive {
				chosen, value, ok := reflect.Select(cases)
				switch chosen {
				case 0: // Context done
					sub.Unsubscribe()
					return
				case 1: // Subscription error
					if !ok {
						// Subscription closed without error
						subscriptionActive = false
						continue
					}

					err := value.Interface().(error)
					// Report the error
					reconnectErr := errors.NewRPCError(err)
					select {
					case errChan <- reconnectErr:
					default:
						// Don't block if error channel is full
					}

					c.log.Warn(fmt.Sprintf("%s subscription error, reconnecting", name),
						logger.Float64("backoff_seconds", backoff),
						logger.Int64("from_block", lastSeenBlock),
						logger.Error(err))

					// Apply backoff before reconnecting
					time.Sleep(time.Duration(backoff) * time.Second)
					// Increase backoff with exponential formula, capped at maximum
					backoff = math.Min(backoff*subscriptionBackoffFactor, subscriptionMaxBackoff)

					// Mark subscription as inactive to break inner loop and create new subscription
					sub.Unsubscribe()
					subscriptionActive = false
				case 2: // Item received
					if !ok {
						// Channel closed
						subscriptionActive = false
						continue
					}

					item := value.Interface()
					// Process the item
					newLastSeenBlock, success := processItem(item, lastSeenBlock)

					// Update last seen block if processing was successful and a newer block was seen
					if success && newLastSeenBlock > lastSeenBlock {
						lastSeenBlock = newLastSeenBlock
					}
				}
			}
		}
	}
}

// Close implements Blockchain.Close
func (c *EVMClient) Close() {
	c.rpcClient.Close()
}

// convertEthereumLogsToLogs converts Ethereum logs to our Log model
func (c *EVMClient) convertEthereumLogsToLogs(logs []*ethTypes.Log) []types.Log {
	result := make([]types.Log, len(logs))
	for i, log := range logs {
		topics := make([]string, len(log.Topics))
		for j, topic := range log.Topics {
			topics[j] = topic.Hex()
		}

		result[i] = types.Log{
			Address:         log.Address.Hex(),
			ChainType:       c.chain.Type,
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
func (c *EVMClient) convertEthereumTransactionToTransaction(tx *ethTypes.Transaction, receipt *ethTypes.Receipt, timestamp uint64) *types.Transaction {
	var status types.TransactionStatus
	var gasUsed uint64
	var blockNumber *big.Int

	if receipt != nil {
		gasUsed = receipt.GasUsed
		blockNumber = receipt.BlockNumber
		if receipt.Status == 1 {
			status = types.TransactionStatusSuccess
		} else {
			status = types.TransactionStatusFailed
		}
	} else {
		status = types.TransactionStatusPending
	}

	from := ""
	signer := ethTypes.LatestSignerForChainID(big.NewInt(c.chain.ID))
	if sender, err := ethTypes.Sender(signer, tx); err == nil {
		from = sender.Hex()
	} else {
		c.log.Warn("Failed to derive sender from transaction", logger.String("tx_hash", tx.Hash().Hex()), logger.Error(err))
	}

	var to string
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	// Determine transaction type
	txType := types.TransactionTypeNative
	// If 'to' is empty and has data, it's a contract deployment transaction
	if to == "" && len(tx.Data()) > 0 {
		txType = types.TransactionTypeDeploy
	} else if len(tx.Data()) > 0 {
		// If 'to' is set and has data, it's a contract call
		txType = types.TransactionTypeContractCall
	}

	// Populate BaseTransaction
	baseTx := types.BaseTransaction{
		ChainType: c.chain.Type,
		Hash:      tx.Hash().Hex(),
		From:      from,
		To:        to,
		Value:     tx.Value(),
		Data:      tx.Data(),
		Nonce:     tx.Nonce(),
		GasPrice:  tx.GasPrice(),
		GasLimit:  tx.Gas(),
		Type:      txType,
	}

	// Create the full Transaction struct
	return &types.Transaction{
		BaseTransaction: baseTx,
		GasUsed:         gasUsed,
		Status:          status,
		Timestamp:       int64(timestamp),
		BlockNumber:     blockNumber,
	}
}

// Chain implements Blockchain.Chain
func (c *EVMClient) Chain() types.Chain {
	return c.chain
}

// convertEventToTopics converts an event signature and arguments to Ethereum topics
func (c *EVMClient) convertEventToTopics(eventSignature string, eventArgs []any) ([][]common.Hash, error) {
	// Generate the event signature hash (topic[0])
	eventID := crypto.Keccak256Hash([]byte(eventSignature))

	// Initialize topics with the event ID as the first topic
	topics := make([][]common.Hash, 1)
	topics[0] = []common.Hash{eventID}

	// If no arguments are provided, return just the event ID topic
	if len(eventArgs) == 0 {
		return topics, nil
	}

	// Parse the event signature to extract parameter types and indexed status
	// Format: "EventName(type1 indexed param1, type2 param2, ...)"
	leftParenIndex := strings.Index(eventSignature, "(")
	rightParenIndex := strings.LastIndex(eventSignature, ")")

	if leftParenIndex == -1 || rightParenIndex == -1 || leftParenIndex >= rightParenIndex {
		return nil, errors.NewInvalidEventSignatureError(eventSignature)
	}

	// Extract parameter part: "type1 indexed param1, type2 param2, ..."
	paramsPart := eventSignature[leftParenIndex+1 : rightParenIndex]

	// Split parameters
	var indexedParams []bool
	if paramsPart != "" {
		params := strings.Split(paramsPart, ",")
		indexedParams = make([]bool, len(params))

		for i, param := range params {
			param = strings.TrimSpace(param)
			indexedParams[i] = strings.Contains(param, "indexed")
		}
	}

	// Count indexed parameters to know how many argument topics to create
	indexedCount := 0
	for _, indexed := range indexedParams {
		if indexed {
			indexedCount++
		}
	}

	// Check if provided arguments match the number of indexed parameters
	if len(eventArgs) > indexedCount {
		return nil, errors.NewInvalidEventArgsError(fmt.Sprintf("Expected %d indexed args, got %d", indexedCount, len(eventArgs)))
	}

	// Add topics for each indexed parameter
	argIndex := 0
	for i, indexed := range indexedParams {
		if indexed && argIndex < len(eventArgs) {
			// For each indexed parameter, create a topic from the corresponding argument
			arg := eventArgs[argIndex]
			argIndex++

			var topic common.Hash

			// Handle different argument types
			switch v := arg.(type) {
			case string:
				// Check if it's an address
				if strings.HasPrefix(v, "0x") && len(v) == 42 {
					// It's an address
					topic = common.HexToHash(v)
				} else {
					// Convert string to bytes32
					topic = crypto.Keccak256Hash([]byte(v))
				}
			case []byte:
				// Convert bytes to hash
				topic = crypto.Keccak256Hash(v)
			case int:
				// Convert int to hash
				bigInt := big.NewInt(int64(v))
				topic = common.BytesToHash(bigInt.Bytes())
			case int64:
				// Convert int64 to hash
				bigInt := big.NewInt(v)
				topic = common.BytesToHash(bigInt.Bytes())
			case *big.Int:
				// Convert big.Int to hash
				topic = common.BytesToHash(v.Bytes())
			case common.Address:
				// Convert address to hash
				topic = common.BytesToHash(v.Bytes())
			case common.Hash:
				// Use hash directly
				topic = v
			case nil:
				// nil value means match any value for this topic
				topics = append(topics, nil)
				continue
			default:
				return nil, errors.NewUnsupportedEventArgTypeError(i + 1)
			}

			// Add topic for this argument
			if len(topics) <= i+1 {
				// Expand topics slice if needed to accommodate this argument
				newTopics := make([][]common.Hash, i+2)
				copy(newTopics, topics)
				topics = newTopics
			}

			if topics[i+1] == nil {
				topics[i+1] = []common.Hash{topic}
			} else {
				topics[i+1] = append(topics[i+1], topic)
			}
		}
	}

	return topics, nil
}

// convertEthereumBlockToBlock converts an Ethereum block to our Block model
func (c *EVMClient) convertEthereumBlockToBlock(block *ethTypes.Block) *types.Block {
	// Extract and convert all transactions
	ethTransactions := block.Transactions()
	transactions := make([]*types.Transaction, len(ethTransactions))

	blockTimestamp := int64(block.Time())
	blockNumber := block.Number()

	for i, tx := range ethTransactions {
		// For transactions in a mined block, we know they've been processed.
		// We will set Status to Mined, but won't fetch the receipt here for efficiency.
		// GetTransaction can be called later for full status (Success/Failed) and GasUsed.
		from := ""
		signer := ethTypes.LatestSignerForChainID(big.NewInt(c.chain.ID))
		if sender, err := ethTypes.Sender(signer, tx); err == nil {
			from = sender.Hex()
		} else {
			c.log.Warn("Failed to derive sender from block transaction", logger.String("tx_hash", tx.Hash().Hex()), logger.Error(err))
		}

		var to string
		if tx.To() != nil {
			to = tx.To().Hex()
		}

		// Determine transaction type
		txType := types.TransactionTypeNative
		if to == "" && len(tx.Data()) > 0 {
			txType = types.TransactionTypeDeploy
		} else if len(tx.Data()) > 0 {
			txType = types.TransactionTypeContractCall
		}

		// Populate BaseTransaction
		baseTx := types.BaseTransaction{
			ChainType: c.chain.Type,
			Hash:      tx.Hash().Hex(),
			From:      from,
			To:        to,
			Value:     tx.Value(),
			Data:      tx.Data(),
			Nonce:     tx.Nonce(),
			GasPrice:  tx.GasPrice(),
			GasLimit:  tx.Gas(),
			Type:      txType,
		}

		// Create the full Transaction struct for the block context
		transactions[i] = &types.Transaction{
			BaseTransaction: baseTx,
			Status:          types.TransactionStatusMined, // Default status for tx in a mined block
			Timestamp:       blockTimestamp,
			BlockNumber:     blockNumber,
			// GasUsed is not available without fetching the receipt
		}
	}

	return &types.Block{
		Hash:             block.Hash().Hex(),
		Number:           block.Number(),
		ParentHash:       block.ParentHash().Hex(),
		Timestamp:        time.Unix(blockTimestamp, 0),
		TransactionCount: len(transactions),
		Transactions:     transactions,
		Miner:            block.Coinbase().Hex(),
		GasUsed:          block.GasUsed(),
		GasLimit:         block.GasLimit(),
		Size:             uint64(block.Size()),
		Difficulty:       block.Difficulty(),
		Extra:            block.Extra(),
	}
}

// retryOperation executes the given operation with exponential backoff retries
// Returns the operation result and a boolean indicating success
func (c *EVMClient) retryOperation(
	operationName string,
	contextInfo map[string]any,
	operation func() (any, error),
	shouldRetry func(error) bool,
) (any, error) {
	var lastErr error
	retryDelay := blockFetchInitialDelay

	for attempt := 0; attempt < blockFetchMaxRetries; attempt++ {
		// Execute the operation
		result, err := operation()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if we should retry
		if attempt < blockFetchMaxRetries-1 && shouldRetry(err) {
			// Create log fields from contextInfo map
			logFields := make([]logger.Field, 0, len(contextInfo)+3)
			for key, value := range contextInfo {
				switch v := value.(type) {
				case string:
					logFields = append(logFields, logger.String(key, v))
				case int:
					logFields = append(logFields, logger.Int(key, v))
				case int64:
					logFields = append(logFields, logger.Int64(key, v))
				case float64:
					logFields = append(logFields, logger.Float64(key, v))
				default:
					logFields = append(logFields, logger.Any(key, v))
				}
			}

			// Add standard retry information
			logFields = append(logFields,
				logger.Int("attempt", attempt+1),
				logger.Duration("delay", retryDelay),
				logger.Error(err),
			)

			// Log the retry attempt
			c.log.Debug("Retrying "+operationName, logFields...)

			// Wait before retry with exponential backoff
			time.Sleep(retryDelay)
			retryDelay *= blockFetchBackoffFactor
			continue
		}

		// If we get here, we're not retrying
		break
	}

	return nil, lastErr
}
