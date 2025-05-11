package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/transaction"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type MonitorService interface {
	// MonitorAddress adds an address to the list of addresses being actively monitored
	// for incoming transactions.
	MonitorAddress(address types.Address) error

	// UnmonitorAddress removes an address from the list of actively monitored addresses.
	UnmonitorAddress(address types.Address) error

	// MonitorContractAddress adds a contract address to monitor for specific events.
	// If events is empty, all known events for the contract will be monitored.
	MonitorContractAddress(address types.Address, events []string) error

	// UnmonitorContractAddress removes a contract address from monitoring for all events.
	UnmonitorContractAddress(address types.Address) error

	// StartTransactionMonitoring starts the process of listening to blockchain events,
	// transforming them, mapping them, and emitting them.
	StartTransactionMonitoring(ctx context.Context) error

	// StopTransactionMonitoring stops the monitoring process.
	StopTransactionMonitoring()

	// TransactionEvents returns a channel that emits processed transactions
	TransactionEvents() <-chan *TransactionEvent
}

// TransactionEvent represents a transaction event with its status
type TransactionEvent struct {
	Transaction types.CoreTransaction
	IsNew       bool
}

// NewMonitorService creates a new transaction monitoring service
func NewMonitorService(
	log logger.Logger,
	blockchainFactory blockchain.Factory,
	chains *types.Chains,
	transformer TransformerService,
	txFactory transaction.Factory,
	repository Repository,
) MonitorService {
	return &monitorService{
		monitorMutex:          sync.RWMutex{},
		transactionEventsChan: make(chan *TransactionEvent, 100), // Buffer size of 100 events
		log:                   log,
		blockchainFactory:     blockchainFactory,
		chains:                chains,
		transformer:           transformer,
		txFactory:             txFactory,
		repository:            repository,
	}
}

type monitorService struct {
	// Monitoring lifecycle management
	monitorCtx            context.Context        // Context for monitoring goroutines
	monitorCancel         context.CancelFunc     // Function to cancel the monitoring context
	monitorMutex          sync.RWMutex           // Mutex for concurrent access to monitor state
	transactionEventsChan chan *TransactionEvent // Channel for emitting transformed/mapped transactions

	// Dependencies
	log               logger.Logger       // Logger for service operations
	blockchainFactory blockchain.Factory  // Factory for creating blockchain monitors
	chains            *types.Chains       // Provider of chain configurations
	transformer       TransformerService  // Transformer for raw transactions
	txFactory         transaction.Factory // Factory for creating transaction mappers
	repository        Repository          // Repository for persisting transactions
}

// MonitorAddress adds an address to the list of monitored addresses.
func (s *monitorService) MonitorAddress(address types.Address) error {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()
	// Get the monitor for the specific chain of the address
	monitor, err := s.blockchainFactory.NewMonitor(address.ChainType)
	if err != nil {
		s.log.Error("Failed to get blockchain monitor for chain",
			logger.String("chain_type", string(address.ChainType)),
			logger.Error(err),
		)
		return err
	}
	err = monitor.MonitorAddress(&address)
	if err != nil {
		s.log.Error("Failed to monitor address on blockchain",
			logger.String("chain_type", string(address.ChainType)),
			logger.String("address", address.String()),
			logger.Error(err),
		)
		return err
	}

	return nil
}

// UnmonitorAddress removes an address from the list of monitored addresses.
func (s *monitorService) UnmonitorAddress(address types.Address) error {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()

	// Get the monitor for the specific chain of the address
	monitor, err := s.blockchainFactory.NewMonitor(address.ChainType)
	if err != nil {
		s.log.Error("Failed to get blockchain monitor for chain",
			logger.String("chain_type", string(address.ChainType)),
			logger.Error(err),
		)
		return err
	}
	err = monitor.UnmonitorAddress(&address)
	if err != nil {
		s.log.Error("Failed to unmonitor address on blockchain",
			logger.String("chain_type", string(address.ChainType)),
			logger.String("address", address.String()),
			logger.Error(err),
		)
		return err
	}

	return nil
}

// MonitorContractAddress adds a contract address to monitor for specific events.
func (s *monitorService) MonitorContractAddress(address types.Address, events []string) error {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()

	// Get the monitor for the specific chain of the address
	monitor, err := s.blockchainFactory.NewMonitor(address.ChainType)
	if err != nil {
		s.log.Error("Failed to get blockchain monitor for chain",
			logger.String("chain_type", string(address.ChainType)),
			logger.Error(err),
		)
		return err
	}

	err = monitor.MonitorContractAddress(&address, events)
	if err != nil {
		s.log.Error("Failed to monitor contract address on blockchain",
			logger.String("chain_type", string(address.ChainType)),
			logger.String("address", address.String()),
			logger.Error(err),
		)
		return err
	}

	return nil
}

// UnmonitorContractAddress removes a contract address from monitoring for all events.
func (s *monitorService) UnmonitorContractAddress(address types.Address) error {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()

	// Get the monitor for the specific chain of the address
	monitor, err := s.blockchainFactory.NewMonitor(address.ChainType)
	if err != nil {
		s.log.Error("Failed to get blockchain monitor for chain",
			logger.String("chain_type", string(address.ChainType)),
			logger.Error(err),
		)
		return err
	}

	err = monitor.UnmonitorContractAddress(&address)
	if err != nil {
		s.log.Error("Failed to unmonitor contract address on blockchain",
			logger.String("chain_type", string(address.ChainType)),
			logger.String("address", address.String()),
			logger.Error(err),
		)
		return err
	}

	return nil
}

// StartTransactionMonitoring starts the process of listening to blockchain events,
// transforming them, mapping them, and emitting them.
func (s *monitorService) StartTransactionMonitoring(ctx context.Context) error {
	s.monitorMutex.Lock()

	// Check if already monitoring
	if s.monitorCtx != nil {
		s.monitorMutex.Unlock()
		s.log.Info("Transaction event monitoring is already active")
		return nil
	}

	// Create a new context based on the input ctx, but cancellable by StopTransactionMonitoring
	s.monitorCtx, s.monitorCancel = context.WithCancel(ctx)
	s.monitorMutex.Unlock() // Unlock after setting context

	s.log.Info("Starting transaction event monitoring for configured chains")

	// Loop through configured chains and start a processor for each
	var startedMonitors int
	for _, chain := range s.chains.List() {
		monitor, err := s.blockchainFactory.NewMonitor(chain.Type)
		if err != nil {
			s.log.Error("Failed to get blockchain monitor for chain, cannot start event processing",
				logger.String("chain_type", string(chain.Type)),
				logger.Error(err),
			)
			continue
		}

		rawEvents := monitor.TransactionEvents()
		if rawEvents == nil {
			s.log.Warn("Blockchain monitor event channel is nil, cannot process events",
				logger.String("chain_type", string(chain.Type)),
			)
			continue
		}

		// Subscribe to transaction events from blockchain monitor
		monitor.SubscribeToTransactionEvents(s.monitorCtx)

		// Process raw transaction events
		go s.processRawTransactionEvents(s.monitorCtx, chain.Type, rawEvents)

		startedMonitors++
		s.log.Info("Started event processor for chain", logger.String("chain_type", string(chain.Type)))
	}

	if startedMonitors == 0 {
		s.log.Warn("No transaction event monitors were started. Check chain configurations and monitor setup.")
		// Cancel the context if no monitors started
		s.monitorMutex.Lock()
		if s.monitorCancel != nil {
			s.monitorCancel()
			s.monitorCtx = nil
			s.monitorCancel = nil
		}
		s.monitorMutex.Unlock()
		return errors.NewOperationFailedError("start transaction monitoring", fmt.Errorf("no monitors started"))
	}

	s.log.Info("Transaction event monitoring started", logger.Int("started_monitors", startedMonitors))
	return nil
}

// StopTransactionMonitoring stops the monitoring process.
func (s *monitorService) StopTransactionMonitoring() {
	s.monitorMutex.Lock()
	defer s.monitorMutex.Unlock()

	if s.monitorCtx == nil {
		return
	}

	s.log.Info("Stopping transaction event monitoring")

	// Cancel the monitoring context to signal goroutines to stop
	if s.monitorCancel != nil {
		s.monitorCancel()
	}

	// Reset context and cancel function
	s.monitorCtx = nil
	s.monitorCancel = nil

	// Close the transformed events channel
	// Ensure this is safe if multiple Stop calls happen (though mutex prevents concurrent calls)
	// Check if channel is already closed? Go idiomatically handles closing closed channels by panicking.
	// We rely on the mutex to prevent double-close panics.
	close(s.transactionEventsChan)

	s.log.Info("Stopped transaction event monitoring")
}

// TransactionEvents returns a channel that emits processed transactions (mapped type or transformed *types.Transaction).
func (s *monitorService) TransactionEvents() <-chan *TransactionEvent {
	return s.transactionEventsChan
}

// Helper method to save transactions
func (s *monitorService) saveTransaction(ctx context.Context, tx *types.Transaction) (bool, error) {
	// Convert to service transaction before saving
	serviceTx := FromCoreTransaction(tx)
	if serviceTx == nil {
		return false, fmt.Errorf("failed to convert transaction to service model")
	}

	// Check if transaction already exists
	existingTx, err := s.repository.GetByHash(ctx, tx.Hash)
	if err == nil && existingTx != nil {
		// Transaction already exists, update it
		existingTx.Status = serviceTx.Status
		if serviceTx.BlockNumber != nil {
			existingTx.BlockNumber = serviceTx.BlockNumber
		}
		if tx.GasUsed > 0 {
			existingTx.GasUsed = sql.NullInt64{
				Int64: int64(tx.GasUsed),
				Valid: true,
			}
		}
		if tx.Timestamp > 0 {
			existingTx.Timestamp = sql.NullInt64{
				Int64: tx.Timestamp,
				Valid: true,
			}
		}
		existingTx.UpdatedAt = time.Now()

		err = s.repository.Update(ctx, existingTx)
		if err != nil {
			s.log.Error("Failed to update transaction",
				logger.String("tx_hash", tx.Hash),
				logger.Error(err),
			)
			return false, err
		}
		return false, nil
	}

	// Transaction doesn't exist, create it
	err = s.repository.Create(ctx, serviceTx)
	if err != nil {
		s.log.Error("Failed to save transaction",
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		return false, err
	}
	return true, nil
}

// processRawTransactionEvents listens to raw events from a specific blockchain monitor,
// transforms, maps, and emits them on the service's transformedEventsChan.
func (s *monitorService) processRawTransactionEvents(ctx context.Context, chainType types.ChainType, rawEvents <-chan *types.Transaction) {
	// Get logger specific to this processor if needed, or use service logger
	procLog := s.log.With(logger.String("processor_id", string(chainType)))
	procLog.Info("Starting raw transaction event processor")

	decoder, err := s.txFactory.NewDecoder(chainType)
	if err != nil {
		procLog.Warn("Failed to create transaction decoder",
			logger.String("chain_type", string(chainType)),
			logger.Error(err),
		)
		return
	}

	for {
		select {
		case <-ctx.Done():
			procLog.Info("Stopping raw transaction event processor due to context cancellation")
			return

		case rawTx, ok := <-rawEvents:
			if !ok {
				procLog.Warn("Raw transaction events channel closed, stopping processor")
				return
			}

			if rawTx == nil {
				procLog.Warn("Received nil transaction from raw events channel, skipping")
				continue
			}

			// 1. Apply transformers
			transformedTx := s.transformer.Apply(ctx, rawTx)

			if transformedTx == nil {
				procLog.Error("Transaction is nil after transformation, cannot save or process further",
					logger.String("original_tx_hash", rawTx.Hash))
				continue
			}

			// 2. Save or update transaction
			isNew, err := s.saveTransaction(ctx, transformedTx)
			if err != nil {
				procLog.Error("Failed to save/update transaction",
					logger.String("tx_hash", transformedTx.Hash),
					logger.Error(err),
				)
				continue
			}

			// 3. Map the transformed transaction
			mappedTx, err := decoder.DecodeTransaction(ctx, transformedTx)
			if err != nil {
				procLog.Warn("Failed to map transaction to specific type",
					logger.String("tx_hash", transformedTx.Hash),
					logger.Error(err),
				)
			}

			// 4. Create TransactionEvent
			txEvent := &TransactionEvent{
				Transaction: mappedTx,
				IsNew:       isNew,
			}

			// 5. Emit the event (non-blocking)
			select {
			case s.transactionEventsChan <- txEvent:
				procLog.Debug("Emitted transaction event",
					logger.String("tx_hash", transformedTx.Hash),
					logger.Bool("is_new", isNew),
				)
			case <-ctx.Done():
				procLog.Info("Stopping emission due to context cancellation during send")
				return
			default:
				// Channel buffer is full, log and drop the event
				procLog.Warn("Transaction events channel is full, dropping event",
					logger.String("tx_hash", transformedTx.Hash),
				)
			}
		}
	}
}
