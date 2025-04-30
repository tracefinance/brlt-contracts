package blockchain

import (
	"context"
	"strings"
	"sync"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// EVMMonitor implements Monitor for EVM-compatible blockchains
type evmMonitor struct {
	log    logger.Logger
	client BlockchainClient

	// Transaction events channel
	transactionEvents chan *types.Transaction

	// Address monitoring state
	monitoredAddresses map[types.ChainType]map[string]struct{}
	addressMutex       sync.RWMutex

	// Contract subscription manager
	manager *SubscriptionManager

	// Context for event subscription
	eventCtx    context.Context
	eventCancel context.CancelFunc
}

// NewEVMMonitor creates a new instance of Monitor for EVM-compatible blockchains
func NewEVMMonitor(
	log logger.Logger,
	client BlockchainClient,
) BLockchainEventMonitor {
	return &evmMonitor{
		log:                log,
		client:             client,
		transactionEvents:  make(chan *types.Transaction, 100), // Buffer size
		monitoredAddresses: make(map[types.ChainType]map[string]struct{}),
		manager:            NewSubscriptionManager(),
	}
}

// TransactionEvents returns a channel that emits raw blockchain transactions.
func (s *evmMonitor) TransactionEvents() <-chan *types.Transaction {
	return s.transactionEvents
}

// MonitorAddress adds an address to the in-memory monitoring list
func (s *evmMonitor) MonitorAddress(addr *types.Address) error {
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
func (s *evmMonitor) UnmonitorAddress(addr *types.Address) error {
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
func (s *evmMonitor) MonitorContractAddress(addr *types.Address, events []string) error {
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
	existingSub := s.manager.GetSubscription(chainType, normalizedAddr)

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
	s.manager.AddOrUpdateSubscription(sub)

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
func (s *evmMonitor) UnmonitorContractAddress(addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup
	chainType := addr.ChainType

	// Get existing subscription
	existingSub := s.manager.GetSubscription(chainType, normalizedAddr)
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
	s.manager.RemoveSubscription(chainType, normalizedAddr)

	s.log.Info("Removed contract from monitoring list",
		logger.String("address", addr.Address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// startContractSubscription starts a new subscription for a contract
func (s *evmMonitor) startContractSubscription(chainType types.ChainType, contractAddr string) {
	// Get the subscription
	sub := s.manager.GetSubscription(chainType, contractAddr)
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
	s.manager.AddOrUpdateSubscription(sub)

	s.log.Info("Starting contract event subscriptions",
		logger.String("chain_type", string(chainType)),
		logger.String("contract_addr", contractAddr),
		logger.Int("event_count", len(eventSigs)))

	// Start a goroutine for each event
	for _, eventSig := range eventSigs {
		go func(eventSignature string) {
			// Subscribe to contract events
			logCh, errCh, err := s.client.SubscribeContractLogs(
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
					s.processContractEventLog(subCtx, log, eventSignature)
				}
			}
		}(eventSig)
	}
}

// SubscribeToTransactionEvents starts listening for new blocks and processing transactions
func (s *evmMonitor) SubscribeToTransactionEvents(ctx context.Context) {
	s.eventCtx, s.eventCancel = context.WithCancel(ctx)
	go s.subscribeToChainBlocks(s.eventCtx)

	subs := s.manager.GetSubscriptionsForChain(s.client.Chain().Type)
	for _, sub := range subs {
		go s.startContractSubscription(s.client.Chain().Type, sub.ContractAddr)
	}
}

// UnsubscribeFromTransactionEvents stops listening for blockchain events
func (s *evmMonitor) UnsubscribeFromTransactionEvents() {
	// Cancel all contract subscriptions
	s.manager.CancelAllSubscriptions()

	// Cancel the main event context
	if s.eventCancel != nil {
		s.eventCancel()
		s.eventCancel = nil
	}

	// Close the transaction events channel
	close(s.transactionEvents)
}

// subscribeToChainBlocks subscribes to new blocks for a specific chain
func (s *evmMonitor) subscribeToChainBlocks(ctx context.Context) {
	// Subscribe to new block headers
	blockCh, errCh, err := s.client.SubscribeNewHead(ctx)
	if err != nil {
		s.log.Error("Failed to subscribe to new blocks",
			logger.Error(err))
		return
	}

	// Process new blocks
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Block subscription stopped")
			return
		case err := <-errCh:
			s.log.Warn("Block subscription error",
				logger.Error(err))
		case block := <-blockCh:
			s.processBlock(&block)
		}
	}
}

// processBlock processes a new block, emitting all transactions found
func (s *evmMonitor) processBlock(block *types.Block) {
	s.log.Debug("Processing new block",
		logger.Int64("block_number", block.Number.Int64()),
		logger.String("block_hash", block.Hash),
		logger.Int("transaction_count", block.TransactionCount))

	// Process each transaction in the block
	for _, tx := range block.Transactions {
		// Check if the transaction involves a monitored address
		if !s.isAddressMonitored(block.ChainType, []string{tx.From, tx.To}) {
			s.log.Debug("Skipping transaction processing, neither address is monitored",
				logger.String("tx_hash", tx.Hash),
				logger.String("chain", string(block.ChainType)),
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
func (s *evmMonitor) processERC20TransferLog(ctx context.Context, log types.Log) {
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
	// Also check the contract address emitting the log
	if !s.isAddressMonitored(log.ChainType, []string{fromAddr, toAddr, log.Address}) {
		s.log.Debug("Skipping ERC20 transfer processing, related addresses not monitored",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("from_address", fromAddr),
			logger.String("to_address", toAddr),
			logger.String("contract_address", log.Address))
		return
	}

	// Get full transaction details from blockchain
	// This transaction object should already be correctly structured (with embedded BaseTransaction)
	// and populated with execution details by the blockchain client implementation.
	fullTx, err := s.client.GetTransaction(ctx, log.TransactionHash)
	if err != nil {
		s.log.Warn("Failed to fetch full transaction details for ERC20 log",
			logger.String("tx_hash", log.TransactionHash),
			logger.Error(err))
		// Optionally, we could construct a partial transaction from the log here if needed,
		// but for now, we'll only emit if we get the full details.
		return
	}

	// Double-check if the fetched transaction involves monitored addresses
	// (The addresses in the log might differ slightly from the tx.From/tx.To in edge cases)
	if !s.isAddressMonitored(log.ChainType, []string{fullTx.From, fullTx.To}) {
		s.log.Debug("Skipping fetched transaction processing, addresses not monitored",
			logger.String("tx_hash", fullTx.Hash),
			logger.String("chain", string(log.ChainType)),
			logger.String("from_address", fullTx.From),
			logger.String("to_address", fullTx.To))
		return
	}

	// Emit the fully populated transaction fetched from the client
	s.emitTransactionEvent(fullTx)
}

// emitTransactionEvent sends a raw transaction to the transaction events channel.
func (s *evmMonitor) emitTransactionEvent(tx *types.Transaction) {
	select {
	case s.transactionEvents <- tx:
		s.log.Debug("Emitted transaction event",
			logger.String("tx_hash", tx.Hash),
			logger.String("chain", string(tx.ChainType)))
	default:
		s.log.Warn("Transaction events channel is full, dropping event",
			logger.String("tx_hash", tx.Hash),
			logger.String("chain", string(tx.ChainType)))
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
func (s *evmMonitor) processContractEventLog(ctx context.Context, log types.Log, eventSig string) {
	// Check if this contract/event combination is monitored
	if !s.isContractMonitored(log.ChainType, log.Address, eventSig) {
		return
	}

	// Process based on event signature
	switch eventSig {
	case string(types.ERC20TransferEventSignature):
		s.processERC20TransferLog(ctx, log)
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
	return s.manager.HasEventSubscription(chainType, contractAddr, eventSig)
}
