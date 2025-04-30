package blockchain

import (
	"context"

	"vault0/internal/logger"
	"vault0/internal/types"
)

// BLockchainEventMonitor handles blockchain event subscriptions for transactions
type BLockchainEventMonitor interface {
	// SubscribeToTransactionEvents starts listening for new blocks and processing transactions.
	// This method:
	// 1. Subscribes to new block headers for all supported chains
	// 2. Processes transactions in those blocks against active wallets
	// 3. Emits transaction events
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the subscription
	SubscribeToTransactionEvents(ctx context.Context)

	// UnsubscribeFromTransactionEvents stops listening for blockchain events.
	// This should be called when shutting down the service.
	UnsubscribeFromTransactionEvents()

	// MonitorAddress adds an address to the list of addresses whose transactions should be emitted.
	MonitorAddress(addr *types.Address) error

	// UnmonitorAddress removes an address from the monitoring list.
	UnmonitorAddress(addr *types.Address) error

	// MonitorContractAddress adds a contract address to monitor for specific events.
	// The events parameter can be a string representation of an event signature from
	// types.ERC20EventSignature or types.MultiSigEventSignature.
	// If events is empty, all known events for the contract will be monitored.
	MonitorContractAddress(addr *types.Address, events []string) error

	// UnmonitorContractAddress removes a contract address from monitoring for all events.
	UnmonitorContractAddress(addr *types.Address) error

	// TransactionEvents returns a channel that emits raw blockchain transactions.
	// These events include all transactions detected on monitored chains.
	// The channel is closed when UnsubscribeFromTransactionEvents is called.
	TransactionEvents() <-chan *types.Transaction
}

// NewMonitor creates a new instance of Monitor
func NewMonitor(
	log logger.Logger,
	client BlockchainClient,
) BLockchainEventMonitor {
	return NewEVMMonitor(log, client)
}
