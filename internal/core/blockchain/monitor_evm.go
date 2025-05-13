package blockchain

import (
	"context"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// EventHandler is a function that processes a contract event log
type EventHandler func(context.Context, types.Log)

// EVMMonitor implements Monitor for EVM-compatible blockchains
type evmMonitor struct {
	log    logger.Logger
	client BlockchainClient

	// Transaction events channel
	transactionEvents chan *types.Transaction

	// Address monitoring component
	addressMonitor *AddressMonitor

	// Contract subscription contractMonitor
	contractMonitor *ContractMonitor

	// Event handlers map
	eventHandlers map[string]EventHandler

	// Context for event subscription
	eventCtx    context.Context
	eventCancel context.CancelFunc
}

// NewEVMMonitor creates a new instance of Monitor for EVM-compatible blockchains
func NewEVMMonitor(
	log logger.Logger,
	client BlockchainClient,
) BLockchainEventMonitor {
	monitor := &evmMonitor{
		log:               log,
		client:            client,
		transactionEvents: make(chan *types.Transaction, 100), // Buffer size
		addressMonitor:    NewAddressMonitor(log),
		contractMonitor:   NewContractMonitor(log),
		eventHandlers:     make(map[string]EventHandler),
	}

	// Register event handlers
	monitor.registerEventHandlers()

	return monitor
}

// registerEventHandlers sets up the event handler map
func (s *evmMonitor) registerEventHandlers() {
	// ERC20 events
	s.eventHandlers[string(types.ERC20TransferEvent)] = s.processERC20TransferLog
	s.eventHandlers[string(types.ERC20ApprovalEvent)] = s.logBasicEvent("ERC20 Approval")

	// MultiSig events
	s.eventHandlers[string(types.MultiSigDepositedEvent)] = s.logBasicEvent("MultiSig Deposited")
	s.eventHandlers[string(types.MultiSigWithdrawalRequestedEvent)] = s.logBasicEvent("MultiSig WithdrawalRequested")
	s.eventHandlers[string(types.MultiSigWithdrawalSignedEvent)] = s.logBasicEvent("MultiSig WithdrawalSigned")
	s.eventHandlers[string(types.MultiSigWithdrawalExecutedEvent)] = s.logBasicEvent("MultiSig WithdrawalExecuted")
	s.eventHandlers[string(types.MultiSigRecoveryRequestedEvent)] = s.logBasicEvent("MultiSig RecoveryRequested")
	s.eventHandlers[string(types.MultiSigRecoveryCancelledEvent)] = s.logBasicEvent("MultiSig RecoveryCancelled")
	s.eventHandlers[string(types.MultiSigRecoveryExecutedEvent)] = s.logBasicEvent("MultiSig RecoveryExecuted")
	s.eventHandlers[string(types.MultiSigRecoveryCompletedEvent)] = s.logBasicEvent("MultiSig RecoveryCompleted")
	s.eventHandlers[string(types.MultiSigTokenSupportedEvent)] = s.logBasicEvent("MultiSig TokenSupported")
	s.eventHandlers[string(types.MultiSigTokenRemovedEvent)] = s.logBasicEvent("MultiSig TokenRemoved")
	s.eventHandlers[string(types.MultiSigNonSupportedTokenRecoveredEvent)] = s.logBasicEvent("MultiSig NonSupportedTokenRecovered")
	s.eventHandlers[string(types.MultiSigTokenWhitelistedEvent)] = s.logBasicEvent("MultiSig TokenWhitelisted")
	s.eventHandlers[string(types.MultiSigRecoveryAddressChangeProposedEvent)] = s.logBasicEvent("MultiSig RecoveryAddressChangeProposed")
	s.eventHandlers[string(types.MultiSigRecoveryAddressChangeSignatureAddedEvent)] = s.logBasicEvent("MultiSig RecoveryAddressChangeSignatureAdded")
	s.eventHandlers[string(types.MultiSigRecoveryAddressChangedEvent)] = s.logBasicEvent("MultiSig RecoveryAddressChanged")
}

// logBasicEvent returns an event handler that just logs basic information about the event
func (s *evmMonitor) logBasicEvent(eventName string) EventHandler {
	return func(ctx context.Context, log types.Log) {
		s.log.Debug("Received "+eventName+" event",
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))
	}
}

// TransactionEvents returns a channel that emits raw blockchain transactions.
func (s *evmMonitor) TransactionEvents() <-chan *types.Transaction {
	return s.transactionEvents
}

// MonitorAddress adds an address to the monitoring list
func (s *evmMonitor) MonitorAddress(addr *types.Address) error {
	return s.addressMonitor.Add(addr)
}

// UnmonitorAddress removes an address from the monitoring list
func (s *evmMonitor) UnmonitorAddress(addr *types.Address) error {
	return s.addressMonitor.Remove(addr)
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

	chainType := addr.ChainType

	// Get existing subscription or create a new one
	existingSub := s.contractMonitor.GetSubscription(chainType, addr.Address)

	updated := false
	eventCount := 0

	if existingSub != nil {
		// Add new events to existing subscription
		for _, event := range events {
			if _, exists := existingSub.Events[event]; !exists {
				existingSub.AddEvent(event)
				updated = true
			}
		}
		eventCount = len(existingSub.Events)
	} else {
		// Create new event set for new subscription
		eventSet := make(EventSet)
		for _, event := range events {
			eventSet[event] = struct{}{}
		}
		// Store the new subscription
		s.contractMonitor.Add(chainType, addr.Address, eventSet)
		updated = true
		eventCount = len(eventSet)
	}

	if !updated && existingSub != nil {
		s.log.Debug("Contract already monitored with these events",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(chainType)))
		return nil
	}

	// Cancel existing subscription if needed
	if updated && existingSub != nil {
		existingSub.Cancel()
	}

	s.log.Info("Added contract to monitoring list with events",
		logger.String("address", addr.Address),
		logger.String("chain_type", string(chainType)),
		logger.Int("event_count", eventCount))

	// Start new subscription if we have an active event context
	if s.eventCtx != nil {
		s.startContractSubscription(chainType, addr.Address)
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

	chainType := addr.ChainType

	// Get existing subscription
	// The GetSubscription method handles normalization internally
	existingSub := s.contractMonitor.GetSubscription(chainType, addr.Address)
	if existingSub == nil {
		s.log.Debug("Contract not found in monitoring list for removal",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(chainType)))
		return nil
	}

	// Cancel subscription if active
	existingSub.Cancel()

	// Remove from registry
	// The RemoveSubscription method handles normalization internally
	s.contractMonitor.Remove(chainType, addr.Address)

	s.log.Info("Removed contract from monitoring list",
		logger.String("address", addr.Address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// startContractSubscription starts a new subscription for a contract
func (s *evmMonitor) startContractSubscription(chainType types.ChainType, contractAddr string) {
	// Get the subscription
	sub := s.contractMonitor.GetSubscription(chainType, contractAddr)
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

	subs := s.contractMonitor.GetSubscriptionsForChain(s.client.Chain().Type)
	for _, sub := range subs {
		go s.startContractSubscription(s.client.Chain().Type, sub.ContractAddr)
	}
}

// UnsubscribeFromTransactionEvents stops listening for blockchain events
func (s *evmMonitor) UnsubscribeFromTransactionEvents() {
	// Cancel all contract subscriptions
	s.contractMonitor.CancelAllSubscriptions()

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
			block.ChainType = s.client.Chain().Type
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
		if !s.addressMonitor.IsMonitored(tx.ChainType, []string{tx.From, tx.To}) {
			s.log.Debug("Skipping transaction processing, neither address is monitored",
				logger.String("tx_hash", tx.Hash),
				logger.String("chain", string(block.ChainType)),
				logger.String("from_address", tx.From),
				logger.String("to_address", tx.To))
			continue
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
	if !s.addressMonitor.IsMonitored(log.ChainType, []string{fromAddr, toAddr, log.Address}) {
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

// processContractEventLog processes a contract event log based on its signature
func (s *evmMonitor) processContractEventLog(ctx context.Context, log types.Log, eventSig string) {
	// Check if this contract/event combination is monitored
	if !s.contractMonitor.IsMonitored(log.ChainType, log.Address, eventSig) {
		return
	}

	// Find and execute the appropriate handler
	if handler, exists := s.eventHandlers[eventSig]; exists {
		handler(ctx, log)
	} else {
		// Default handler for unknown events
		s.log.Debug("Received unknown contract event",
			logger.String("event_signature", eventSig),
			logger.String("tx_hash", log.TransactionHash),
			logger.String("contract", log.Address))
	}
}
