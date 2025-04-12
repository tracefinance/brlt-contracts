package transaction

import (
	"context"
	"math/big"
	"strings"
	"time"

	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// MonitorAddress adds an address to the in-memory monitoring list
func (s *transactionService) MonitorAddress(ctx context.Context, addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup

	s.addressMutex.Lock()
	defer s.addressMutex.Unlock()

	if _, ok := s.monitoredAddresses[addr.ChainType]; !ok {
		s.monitoredAddresses[addr.ChainType] = make(map[string]struct{})
	}

	if _, exists := s.monitoredAddresses[addr.ChainType][normalizedAddr]; !exists {
		s.monitoredAddresses[addr.ChainType][normalizedAddr] = struct{}{}
		s.log.Info("Added address to monitoring list",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(addr.ChainType)))
	} else {
		s.log.Debug("Address already monitored",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(addr.ChainType)))
	}

	return nil
}

// UnmonitoredAddress removes an address from the in-memory monitoring list
func (s *transactionService) UnmonitoredAddress(ctx context.Context, addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	// We don't strictly need validation here, but it's good practice
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup

	s.addressMutex.Lock()
	defer s.addressMutex.Unlock()

	if chainMap, ok := s.monitoredAddresses[addr.ChainType]; ok {
		if _, exists := chainMap[normalizedAddr]; exists {
			delete(chainMap, normalizedAddr)
			s.log.Info("Removed address from monitoring list",
				logger.String("address", addr.Address),
				logger.String("chain_type", string(addr.ChainType)))
			// Clean up the chain map if it becomes empty
			if len(chainMap) == 0 {
				delete(s.monitoredAddresses, addr.ChainType)
			}
		} else {
			s.log.Debug("Address not found in monitoring list for removal",
				logger.String("address", addr.Address),
				logger.String("chain_type", string(addr.ChainType)))
		}
	} else {
		s.log.Debug("Chain type not found in monitoring list for removal",
			logger.String("chain_type", string(addr.ChainType)))
	}

	return nil
}

// SubscribeToTransactionEvents starts listening for new blocks and processing transactions
func (s *transactionService) SubscribeToTransactionEvents(ctx context.Context) {
	s.eventCtx, s.eventCancel = context.WithCancel(ctx)

	// Get list of unique chain types from chains
	for _, chain := range s.chains.Chains {
		// Start a goroutine for each chain to subscribe to new blocks
		go s.subscribeToChainBlocks(s.eventCtx, chain)

		// Start a goroutine for each chain to subscribe to ERC20 transfers
		go s.subscribeToERC20Transfers(s.eventCtx, chain)
	}

	// Subscribe to token events to dynamically update ERC20 token subscriptions
	go s.subscribeToTokenEvents(s.eventCtx)
}

// UnsubscribeFromTransactionEvents stops listening for blockchain events
func (s *transactionService) UnsubscribeFromTransactionEvents() {
	if s.eventCancel != nil {
		s.eventCancel()
		s.eventCancel = nil
	}

	// Close the transaction events channel
	close(s.transactionEvents)
}

// subscribeToChainBlocks subscribes to new blocks for a specific chain
func (s *transactionService) subscribeToChainBlocks(ctx context.Context, chain types.Chain) {
	// Get blockchain client for the chain type
	client, err := s.blockchainRegistry.GetBlockchain(chain.Type)
	if err != nil {
		s.log.Error("Failed to get blockchain client",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	s.log.Info("Starting new block subscription",
		logger.String("chain_type", string(chain.Type)))

	// Subscribe to new block headers
	blockCh, errCh, err := client.SubscribeNewHead(ctx)
	if err != nil {
		s.log.Error("Failed to subscribe to new blocks",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Process new blocks
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Block subscription stopped",
				logger.String("chain_type", string(chain.Type)))
			return
		case err := <-errCh:
			s.log.Warn("Block subscription error",
				logger.String("chain_type", string(chain.Type)),
				logger.Error(err))
		case block := <-blockCh:
			s.processBlock(chain.Type, &block)
		}
	}
}

// processBlock processes a new block, emitting all transactions found
func (s *transactionService) processBlock(chainType types.ChainType, block *types.Block) {
	s.log.Debug("Processing new block",
		logger.String("chain_type", string(chainType)),
		logger.Int64("block_number", block.Number.Int64()),
		logger.String("block_hash", block.Hash),
		logger.Int("transaction_count", block.TransactionCount))

	// Process each transaction in the block
	for _, tx := range block.Transactions {
		// Check if the transaction involves a monitored address
		if !s.isAddressMonitored(chainType, []string{tx.From, tx.To}) {
			s.log.Debug("Skipping transaction processing, neither address is monitored",
				logger.String("tx_hash", tx.Hash),
				logger.String("chain", string(chainType)),
				logger.String("from_address", tx.From),
				logger.String("to_address", tx.To))
			return
		}

		// Set the timestamp to the block timestamp if not already set
		if tx.Timestamp == 0 {
			tx.Timestamp = block.Timestamp.Unix()
		}

		s.emitTransactionEvent(tx)
	}
}

// processERC20TransferLog processes an ERC20 Transfer event log and emits the transaction
func (s *transactionService) processERC20TransferLog(ctx context.Context, chain types.Chain, log types.Log) {
	// Check if we have enough topics (event signature + from + to)
	if len(log.Topics) < 3 {
		s.log.Warn("Invalid ERC20 transfer log format",
			logger.String("tx_hash", log.TransactionHash))
		return
	}

	// Extract and parse 'from' address using the address utilities
	// The topics in ERC20 Transfer events are 32 bytes with padding, so we need to extract the last 20 bytes
	fromTopicAddress := "0x" + log.Topics[1][len(log.Topics[1])-40:] // Last 40 hex chars (20 bytes)
	fromAddrObj, err := types.NewAddress(fromTopicAddress, chain.Type)
	if err != nil {
		s.log.Warn("Invalid from address in ERC20 transfer",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("raw_address", fromTopicAddress),
			logger.Error(err))
		return
	}
	fromAddr := fromAddrObj.ToChecksum()

	// Extract and parse 'to' address using the address utilities
	toTopicAddress := "0x" + log.Topics[2][len(log.Topics[2])-40:] // Last 40 hex chars (20 bytes)
	toAddrObj, err := types.NewAddress(toTopicAddress, chain.Type)
	if err != nil {
		s.log.Warn("Invalid to address in ERC20 transfer",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("raw_address", toTopicAddress),
			logger.Error(err))
		return
	}
	toAddr := toAddrObj.ToChecksum()

	// Check if either the 'from' or 'to' address is monitored before fetching details
	if !s.isAddressMonitored(chain.Type, []string{fromAddr, toAddr}) {
		s.log.Debug("Skipping ERC20 transfer processing, neither address is monitored",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("from_address", fromAddr),
			logger.String("to_address", toAddr))
		return
	}

	// Get blockchain client to fetch transaction details and block information
	client, err := s.blockchainRegistry.GetBlockchain(chain.Type)
	if err != nil {
		s.log.Error("Failed to get blockchain client for transaction details",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Fetch the full transaction to get gas price and gas limit
	var gasPrice *big.Int
	var gasLimit uint64
	var gasUsed uint64
	var nonce uint64
	var timestamp int64

	// Get full transaction details from blockchain
	fullTx, err := client.GetTransaction(ctx, log.TransactionHash)
	if err != nil {
		s.log.Warn("Failed to fetch transaction details, continuing with limited data",
			logger.String("tx_hash", log.TransactionHash),
			logger.Error(err))
	} else {
		// Extract gas details from the transaction
		gasPrice = fullTx.GasPrice
		gasLimit = fullTx.GasLimit
		gasUsed = fullTx.GasUsed
		nonce = fullTx.Nonce
		timestamp = fullTx.Timestamp
	}

	// If timestamp is not available from the transaction, try to get it from the block
	if timestamp == 0 && log.BlockNumber != nil {
		block, err := client.GetBlock(ctx, log.BlockNumber.String())
		if err != nil {
			s.log.Warn("Failed to fetch block for timestamp, continuing without it",
				logger.String("block_number", log.BlockNumber.String()),
				logger.Error(err))
		} else {
			timestamp = block.Timestamp.Unix()
		}
	}

	// Create a new transaction directly from the log data
	// Parse the token address using the address utilities
	tokenAddrObj, err := types.NewAddress(log.Address, chain.Type)
	if err != nil {
		s.log.Warn("Invalid token address in ERC20 transfer",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("raw_address", log.Address),
			logger.Error(err))
		return
	}
	tokenAddress := tokenAddrObj.ToChecksum()

	// Parse the transfer amount from log data
	var value *big.Int
	if len(log.Data) > 0 {
		value = new(big.Int).SetBytes(log.Data)
	} else {
		value = big.NewInt(0)
	}

	// Create a transaction record with all available details
	tx := &types.Transaction{
		Chain:        chain.Type,
		Hash:         log.TransactionHash,
		From:         fromAddr,
		To:           toAddr,
		Value:        value,
		Type:         types.TransactionTypeERC20,
		TokenAddress: tokenAddress,
		Status:       types.TransactionStatusSuccess, // ERC20 transfer logs occur only for successful transfers
		BlockNumber:  log.BlockNumber,
		Timestamp:    timestamp,
		GasPrice:     gasPrice,
		GasLimit:     gasLimit,
		GasUsed:      gasUsed,
		Nonce:        nonce,
	}

	// Get token details from token store to enrich the emitted transaction
	token, err := s.tokenStore.GetToken(ctx, tokenAddress)
	if err == nil && token != nil {
		tx.TokenSymbol = token.Symbol
	} else {
		s.log.Warn("Token not found in token store for ERC20 transfer, emitting without symbol",
			logger.String("token_address", tokenAddress),
			logger.String("chain", string(chain.Type)))
	}

	// Emit the raw transaction (filtered by monitored addresses inside emitTransactionEvent)
	s.emitTransactionEvent(tx)
}

// subscribeToTokenEvents listens for token event notifications and updates ERC20 subscriptions
func (s *transactionService) subscribeToTokenEvents(ctx context.Context) {
	tokenEvents := s.tokenStore.TokenEvents()
	s.log.Info("Started token events subscription")

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Token events subscription stopped")
			return
		case event, ok := <-tokenEvents:
			if !ok {
				// Channel was closed
				s.log.Info("Token events channel closed")
				return
			}

			// Only react to token added events
			if event.EventType == tokenstore.TokenEventAdded && event.Token != nil {
				// Skip native tokens
				if event.Token.IsNative() {
					continue
				}

				s.log.Info("New token added, updating ERC20 subscription",
					logger.String("symbol", event.Token.Symbol),
					logger.String("address", event.Token.Address),
					logger.String("chain", string(event.Token.ChainType)))

				// For simplicity, restart the entire ERC20 subscription for this chain
				// A more optimized approach would be to add this token to existing subscriptions
				if chain, exists := s.chains.Chains[event.Token.ChainType]; exists {
					// Create a new context for the restarted subscription
					tokenCtx, cancel := context.WithCancel(ctx)

					// Start new subscription
					go func(chain types.Chain) {
						s.subscribeToERC20Transfers(tokenCtx, chain)
					}(chain)

					// Cancel any previous subscription for this chain after a short delay
					// This ensures we have the new subscription running before canceling the old one
					go func() {
						time.Sleep(5 * time.Second)
						cancel()
					}()
				}
			}
		}
	}
}

// subscribeToERC20Transfers subscribes to ERC20 token transfer events on a specific chain
func (s *transactionService) subscribeToERC20Transfers(ctx context.Context, chain types.Chain) {
	// Get blockchain client for the chain type
	client, err := s.blockchainRegistry.GetBlockchain(chain.Type)
	if err != nil {
		s.log.Error("Failed to get blockchain client for ERC20 subscription",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Get tokens from the token store for this chain
	tokenPage, err := s.tokenStore.ListTokensByChain(ctx, chain.Type, 0, 0)
	if err != nil {
		s.log.Error("Failed to get tokens for ERC20 subscription",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Filter out native tokens and collect token addresses
	var tokenAddresses []string
	for _, token := range tokenPage.Items {
		// Skip native tokens (they have empty contract addresses)
		if token.Address == "" || token.IsNative() {
			continue
		}
		tokenAddresses = append(tokenAddresses, token.Address)
	}

	// If no token addresses found, log and exit
	if len(tokenAddresses) == 0 {
		s.log.Info("No ERC20 tokens found in token store, skipping subscription",
			logger.String("chain_type", string(chain.Type)))
		return
	}

	s.log.Info("Starting ERC20 transfer event subscription",
		logger.String("chain_type", string(chain.Type)),
		logger.Int("token_count", len(tokenAddresses)))

	// Subscribe to Transfer events for specific token contracts
	logCh, errCh, err := client.SubscribeContractLogs(
		ctx,
		tokenAddresses,
		string(types.ERC20TransferEventSignature),
		nil, // No specific args filter, emitTransactionEvent handles address filtering
		0,   // Start from recent blocks
	)

	if err != nil {
		s.log.Error("Failed to subscribe to ERC20 transfers",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Process the logs
	for {
		select {
		case <-ctx.Done():
			s.log.Info("ERC20 subscription stopped",
				logger.String("chain_type", string(chain.Type)))
			return
		case err := <-errCh:
			s.log.Warn("ERC20 subscription error",
				logger.String("chain_type", string(chain.Type)),
				logger.Error(err))
		case log := <-logCh:
			s.processERC20TransferLog(ctx, chain, log)
		}
	}
}

// emitTransactionEvent sends a raw transaction to the transaction events channel.
func (s *transactionService) emitTransactionEvent(tx *types.Transaction) {
	// Check is now performed in processBlock and processERC20TransferLog

	select {
	case s.transactionEvents <- tx:
		s.log.Debug("Emitted transaction event",
			logger.String("tx_hash", tx.Hash),
			logger.String("chain", string(tx.Chain)))
	default:
		s.log.Warn("Transaction events channel is full, dropping event",
			logger.String("tx_hash", tx.Hash),
			logger.String("chain", string(tx.Chain)))
	}
}

// isAddressMonitored checks if any of the given addresses on a specific chain are in the
// monitored addresses list. It acquires a read lock on the address mutex.
func (s *transactionService) isAddressMonitored(chainType types.ChainType, addresses []string) bool {
	s.addressMutex.RLock()
	defer s.addressMutex.RUnlock()

	chainMap, chainExists := s.monitoredAddresses[chainType]
	if !chainExists {
		return false
	}

	for _, addr := range addresses {
		normalizedAddr := strings.ToLower(addr)
		if _, monitored := chainMap[normalizedAddr]; monitored {
			return true
		}
	}

	return false
}
