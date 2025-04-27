package transaction

import (
	"context"
	"math/big"
	"strings"
	"sync"

	"vault0/internal/core/blockchain"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// EVMMonitor implements Monitor for EVM-compatible blockchains
type evmMonitor struct {
	log                logger.Logger
	blockchainRegistry blockchain.Factory
	tokenStore         tokenstore.TokenStore
	chains             *types.Chains

	// Transaction events channel
	transactionEvents chan *types.Transaction

	// Address monitoring state
	monitoredAddresses map[types.ChainType]map[string]struct{}
	addressMutex       sync.RWMutex

	// Contract subscription registry
	subscriptionRegistry *SubscriptionRegistry

	// Context for event subscription
	eventCtx    context.Context
	eventCancel context.CancelFunc
}

// NewEVMMonitor creates a new instance of Monitor for EVM-compatible blockchains
func NewEVMMonitor(
	log logger.Logger,
	blockchainRegistry blockchain.Factory,
	tokenStore tokenstore.TokenStore,
	chains *types.Chains,
) Monitor {
	return &evmMonitor{
		log:                  log,
		blockchainRegistry:   blockchainRegistry,
		tokenStore:           tokenStore,
		chains:               chains,
		transactionEvents:    make(chan *types.Transaction, 100), // Buffer size
		monitoredAddresses:   make(map[types.ChainType]map[string]struct{}),
		subscriptionRegistry: NewSubscriptionRegistry(),
	}
}

// TransactionEvents returns a channel that emits raw blockchain transactions.
func (s *evmMonitor) TransactionEvents() <-chan *types.Transaction {
	return s.transactionEvents
}

// MonitorAddress adds an address to the in-memory monitoring list
func (s *evmMonitor) MonitorAddress(ctx context.Context, addr *types.Address) error {
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

// UnmonitorAddress removes an address from the in-memory monitoring list
func (s *evmMonitor) UnmonitorAddress(ctx context.Context, addr *types.Address) error {
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

// MonitorContractAddress adds a contract address to monitor for specific events
func (s *evmMonitor) MonitorContractAddress(ctx context.Context, addr *types.Address, events []string) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	if err := addr.Validate(); err != nil {
		return err
	}

	// Require at least one event to be specified
	if len(events) == 0 {
		return errors.NewInvalidInputError("At least one event must be specified", "events", nil)
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup
	chainType := addr.ChainType

	// Get existing subscription or create a new one
	existingSub := s.subscriptionRegistry.GetSubscription(chainType, normalizedAddr)

	// Create event map
	eventMap := make(map[string]struct{})
	if existingSub != nil {
		// Copy existing events
		for event := range existingSub.Events {
			eventMap[event] = struct{}{}
		}
	}

	// Add new events
	updated := false
	for _, event := range events {
		if _, exists := eventMap[event]; !exists {
			eventMap[event] = struct{}{}
			updated = true
		}
	}

	if !updated && existingSub != nil {
		s.log.Debug("Contract already monitored with these events",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(chainType)))
		return nil
	}

	// Create or update subscription
	sub := &ContractSubscription{
		ChainType:    chainType,
		ContractAddr: normalizedAddr,
		Events:       eventMap,
	}

	// Cancel existing subscription if any
	if existingSub != nil && existingSub.CancelFunc != nil {
		existingSub.CancelFunc()
	}

	// Store the new subscription
	s.subscriptionRegistry.AddOrUpdateSubscription(sub)

	s.log.Info("Added contract to monitoring list with events",
		logger.String("address", addr.Address),
		logger.String("chain_type", string(chainType)),
		logger.Int("event_count", len(eventMap)))

	// Start new subscription if we have an active event context
	if s.eventCtx != nil {
		s.startContractSubscription(chainType, normalizedAddr)
	}

	return nil
}

// UnmonitorContractAddress removes a contract address from the monitoring list
func (s *evmMonitor) UnmonitorContractAddress(ctx context.Context, addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup
	chainType := addr.ChainType

	// Get existing subscription
	existingSub := s.subscriptionRegistry.GetSubscription(chainType, normalizedAddr)
	if existingSub == nil {
		s.log.Debug("Contract not found in monitoring list for removal",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(chainType)))
		return nil
	}

	// Cancel subscription if active
	if existingSub.CancelFunc != nil {
		existingSub.CancelFunc()
	}

	// Remove from registry
	s.subscriptionRegistry.RemoveSubscription(chainType, normalizedAddr)

	s.log.Info("Removed contract from monitoring list",
		logger.String("address", addr.Address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// startContractSubscription starts a new subscription for a contract
func (s *evmMonitor) startContractSubscription(chainType types.ChainType, contractAddr string) {
	// Get the chain configuration
	chain, exists := s.chains.Chains[chainType]
	if !exists {
		s.log.Warn("Cannot start contract subscription, chain not found",
			logger.String("chain_type", string(chainType)))
		return
	}

	// Get the subscription
	sub := s.subscriptionRegistry.GetSubscription(chainType, contractAddr)
	if sub == nil {
		s.log.Warn("Cannot start contract subscription, subscription not found",
			logger.String("chain_type", string(chainType)),
			logger.String("contract_addr", contractAddr))
		return
	}

	// Convert event set to slice
	eventSigs := make([]string, 0, len(sub.Events))
	for event := range sub.Events {
		eventSigs = append(eventSigs, event)
	}

	if len(eventSigs) == 0 {
		s.log.Warn("No events to monitor for contract, skipping subscription",
			logger.String("chain_type", string(chainType)),
			logger.String("contract_addr", contractAddr))
		return
	}

	// Create a context for this subscription
	subCtx, cancel := context.WithCancel(s.eventCtx)
	sub.CancelFunc = cancel

	// Update the subscription in registry with cancel function
	s.subscriptionRegistry.AddOrUpdateSubscription(sub)

	// Get blockchain client
	client, err := s.blockchainRegistry.NewClient(chainType)
	if err != nil {
		s.log.Error("Failed to get blockchain client for contract subscription",
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	s.log.Info("Starting contract event subscriptions",
		logger.String("chain_type", string(chainType)),
		logger.String("contract_addr", contractAddr),
		logger.Int("event_count", len(eventSigs)))

	// Start a goroutine for each event
	for _, eventSig := range eventSigs {
		go func(eventSignature string) {
			// Subscribe to contract events
			logCh, errCh, err := client.SubscribeContractLogs(
				subCtx,
				[]string{contractAddr},
				eventSignature,
				nil, // No specific args filter
				0,   // Start from recent blocks
			)

			if err != nil {
				s.log.Error("Failed to subscribe to contract event",
					logger.String("chain_type", string(chainType)),
					logger.String("contract_addr", contractAddr),
					logger.String("event_signature", eventSignature),
					logger.Error(err))
				return
			}

			// Process the logs
			for {
				select {
				case <-subCtx.Done():
					s.log.Info("Contract event subscription stopped",
						logger.String("chain_type", string(chainType)),
						logger.String("contract_addr", contractAddr),
						logger.String("event_signature", eventSignature))
					return
				case err := <-errCh:
					s.log.Warn("Contract event subscription error",
						logger.String("chain_type", string(chainType)),
						logger.String("contract_addr", contractAddr),
						logger.String("event_signature", eventSignature),
						logger.Error(err))
				case log := <-logCh:
					// Process the log based on event signature
					s.processContractEventLog(subCtx, chain, log, eventSignature)
				}
			}
		}(eventSig)
	}
}

// SubscribeToTransactionEvents starts listening for new blocks and processing transactions
func (s *evmMonitor) SubscribeToTransactionEvents(ctx context.Context) {
	s.eventCtx, s.eventCancel = context.WithCancel(ctx)

	// Get list of unique chain types from chains
	for _, chain := range s.chains.Chains {
		// Start a goroutine for each chain to subscribe to new blocks
		go s.subscribeToChainBlocks(s.eventCtx, chain)

		// Start subscriptions for all monitored contracts on this chain
		subs := s.subscriptionRegistry.GetSubscriptionsForChain(chain.Type)
		for _, sub := range subs {
			go s.startContractSubscription(chain.Type, sub.ContractAddr)
		}
	}
}

// UnsubscribeFromTransactionEvents stops listening for blockchain events
func (s *evmMonitor) UnsubscribeFromTransactionEvents() {
	// Cancel all contract subscriptions
	s.subscriptionRegistry.CancelAllSubscriptions()

	// Cancel the main event context
	if s.eventCancel != nil {
		s.eventCancel()
		s.eventCancel = nil
	}

	// Close the transaction events channel
	close(s.transactionEvents)
}

// subscribeToChainBlocks subscribes to new blocks for a specific chain
func (s *evmMonitor) subscribeToChainBlocks(ctx context.Context, chain types.Chain) {
	// Get blockchain client for the chain type
	client, err := s.blockchainRegistry.NewClient(chain.Type)
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
func (s *evmMonitor) processBlock(chainType types.ChainType, block *types.Block) {
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
func (s *evmMonitor) processERC20TransferLog(ctx context.Context, chain types.Chain, log types.Log) {
	// Check if we have enough topics (event signature + from + to)
	if len(log.Topics) < 3 {
		s.log.Warn("Invalid ERC20 transfer log format: insufficient topics",
			logger.String("tx_hash", log.TransactionHash),
			logger.Int("topic_count", len(log.Topics)))
		return
	}

	// Parse 'from' address from the second topic (index 1)
	fromAddrObj, err := log.ParseAddressFromTopic(1)
	if err != nil {
		s.log.Warn("Failed to parse 'from' address from ERC20 transfer topic",
			logger.String("tx_hash", log.TransactionHash),
			logger.Int("topic_index", 1),
			logger.String("topic_value", log.Topics[1]),
			logger.Error(err))
		return
	}
	fromAddr := fromAddrObj.ToChecksum()

	// Parse 'to' address from the third topic (index 2)
	toAddrObj, err := log.ParseAddressFromTopic(2)
	if err != nil {
		s.log.Warn("Failed to parse 'to' address from ERC20 transfer topic",
			logger.String("tx_hash", log.TransactionHash),
			logger.Int("topic_index", 2),
			logger.String("topic_value", log.Topics[2]),
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
	client, err := s.blockchainRegistry.NewClient(chain.Type)
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
	tokenAddrObj, err := types.NewAddress(chain.Type, log.Address)
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

// emitTransactionEvent sends a raw transaction to the transaction events channel.
func (s *evmMonitor) emitTransactionEvent(tx *types.Transaction) {
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
func (s *evmMonitor) isAddressMonitored(chainType types.ChainType, addresses []string) bool {
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

// processContractEventLog processes a contract event log based on its signature
func (s *evmMonitor) processContractEventLog(ctx context.Context, chain types.Chain, log types.Log, eventSig string) {
	// Check if this contract/event combination is monitored
	if !s.isContractMonitored(chain.Type, log.Address, eventSig) {
		return
	}

	// Process based on event signature
	switch eventSig {
	case string(types.ERC20TransferEventSignature):
		// Process ERC20 Transfer event
		s.processERC20TransferLog(ctx, chain, log)
	// Additional cases for other ERC20 events
	case string(types.ERC20ApprovalEventSignature):
		// Process ERC20 Approval event if needed
		s.log.Debug("Received ERC20 Approval event",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))

	// Add cases for MultiSig events
	case string(types.MultiSigDepositedEventSig):
		s.log.Debug("Received MultiSig Deposited event",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))

	case string(types.MultiSigWithdrawalRequestedEventSig):
		s.log.Debug("Received MultiSig WithdrawalRequested event",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))

	// Add other MultiSig event cases
	case string(types.MultiSigWithdrawalSignedEventSig),
		string(types.MultiSigWithdrawalExecutedEventSig),
		string(types.MultiSigRecoveryRequestedEventSig),
		string(types.MultiSigRecoveryCancelledEventSig),
		string(types.MultiSigRecoveryExecutedEventSig),
		string(types.MultiSigRecoveryCompletedEventSig),
		string(types.MultiSigTokenSupportedEventSig),
		string(types.MultiSigTokenRemovedEventSig),
		string(types.MultiSigNonSupportedTokenRecoveredEventSig),
		string(types.MultiSigTokenWhitelistedEventSig),
		string(types.MultiSigRecoveryAddressChangeProposedEventSig),
		string(types.MultiSigRecoveryAddressChangeSignatureAddedEventSig),
		string(types.MultiSigRecoveryAddressChangedEventSig):
		// Log basic info for now - these would be processed according to their specific logic
		s.log.Debug("Received MultiSig event",
			logger.String("event", eventSig),
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))

	default:
		s.log.Debug("Received unknown contract event",
			logger.String("event_signature", eventSig),
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))
	}
}

// isContractMonitored checks if a contract address is monitored for a specific event
func (s *evmMonitor) isContractMonitored(chainType types.ChainType, contractAddr string, eventSig string) bool {
	return s.subscriptionRegistry.HasEventSubscription(chainType, contractAddr, eventSig)
}
