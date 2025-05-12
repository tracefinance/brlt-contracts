package blockchain

import (
	"context"
	"strings"
	"sync"
	"vault0/internal/types"
)

// ContractSubscription represents an active subscription for a contract's events
type ContractSubscription struct {
	ChainType    types.ChainType
	ContractAddr string              // Normalized (lowercase) contract address
	Events       map[string]struct{} // Set of event signatures being monitored
	CancelFunc   context.CancelFunc  // Function to cancel all subscriptions for this contract
}

// ContractMonitor manages contract event subscriptions across different chains
// It provides thread-safe access to subscription data and operations
type ContractMonitor struct {
	// Map of chain -> contract address -> subscription
	subscriptions map[types.ChainType]map[string]*ContractSubscription
	mutex         sync.RWMutex
}

// NewContractMonitor creates a new subscription manager
func NewContractMonitor() *ContractMonitor {
	return &ContractMonitor{
		subscriptions: make(map[types.ChainType]map[string]*ContractSubscription),
	}
}

// GetSubscription retrieves a subscription for a contract
// Returns nil if no subscription exists for the given chain and contract address
func (m *ContractMonitor) GetSubscription(chainType types.ChainType, contractAddr string) *ContractSubscription {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	normalizedAddr := strings.ToLower(contractAddr)

	if chainMap, exists := m.subscriptions[chainType]; exists {
		return chainMap[normalizedAddr]
	}
	return nil
}

// AddOrUpdateSubscription adds or updates a subscription
// If a subscription already exists for the given chain and contract address,
// it will be replaced with the new subscription
func (m *ContractMonitor) AddOrUpdateSubscription(sub *ContractSubscription) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Ensure chain map exists
	if _, exists := m.subscriptions[sub.ChainType]; !exists {
		m.subscriptions[sub.ChainType] = make(map[string]*ContractSubscription)
	}

	// Store subscription
	m.subscriptions[sub.ChainType][sub.ContractAddr] = sub
}

// RemoveSubscription removes a subscription
// If the subscription doesn't exist, this is a no-op
func (m *ContractMonitor) RemoveSubscription(chainType types.ChainType, contractAddr string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	normalizedAddr := strings.ToLower(contractAddr)

	if chainMap, exists := m.subscriptions[chainType]; exists {
		delete(chainMap, normalizedAddr)

		// Clean up empty chain maps
		if len(chainMap) == 0 {
			delete(m.subscriptions, chainType)
		}
	}
}

// CancelAllSubscriptions cancels all subscriptions by calling their CancelFunc
// and clears the subscription map
func (m *ContractMonitor) CancelAllSubscriptions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, chainMap := range m.subscriptions {
		for _, sub := range chainMap {
			if sub.CancelFunc != nil {
				sub.CancelFunc()
			}
		}
	}

	// Clear all subscriptions
	m.subscriptions = make(map[types.ChainType]map[string]*ContractSubscription)
}

// GetSubscriptionsForChain returns all subscriptions for a chain
// Returns an empty slice if no subscriptions exist for the given chain
func (m *ContractMonitor) GetSubscriptionsForChain(chainType types.ChainType) []*ContractSubscription {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var result []*ContractSubscription

	if chainMap, exists := m.subscriptions[chainType]; exists {
		result = make([]*ContractSubscription, 0, len(chainMap))
		for _, sub := range chainMap {
			result = append(result, sub)
		}
	}

	return result
}

// HasEventSubscription checks if a contract has a subscription for a specific event
func (m *ContractMonitor) HasEventSubscription(chainType types.ChainType, contractAddr string, eventSig string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	normalizedAddr := strings.ToLower(contractAddr)

	if chainMap, exists := m.subscriptions[chainType]; exists {
		if sub, exists := chainMap[normalizedAddr]; exists {
			_, hasEvent := sub.Events[eventSig]
			return hasEvent
		}
	}

	return false
}
