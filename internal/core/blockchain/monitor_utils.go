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

// SubscriptionManager manages contract event subscriptions
type SubscriptionManager struct {
	// Map of chain -> contract address -> subscription
	subscriptions map[types.ChainType]map[string]*ContractSubscription
	mutex         sync.RWMutex
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager() *SubscriptionManager {
	return &SubscriptionManager{
		subscriptions: make(map[types.ChainType]map[string]*ContractSubscription),
	}
}

// GetSubscription retrieves a subscription for a contract
func (r *SubscriptionManager) GetSubscription(chainType types.ChainType, contractAddr string) *ContractSubscription {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	normalizedAddr := strings.ToLower(contractAddr)

	if chainMap, exists := r.subscriptions[chainType]; exists {
		return chainMap[normalizedAddr]
	}
	return nil
}

// AddOrUpdateSubscription adds or updates a subscription
func (r *SubscriptionManager) AddOrUpdateSubscription(sub *ContractSubscription) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Ensure chain map exists
	if _, exists := r.subscriptions[sub.ChainType]; !exists {
		r.subscriptions[sub.ChainType] = make(map[string]*ContractSubscription)
	}

	// Store subscription
	r.subscriptions[sub.ChainType][sub.ContractAddr] = sub
}

// RemoveSubscription removes a subscription
func (r *SubscriptionManager) RemoveSubscription(chainType types.ChainType, contractAddr string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	normalizedAddr := strings.ToLower(contractAddr)

	if chainMap, exists := r.subscriptions[chainType]; exists {
		delete(chainMap, normalizedAddr)

		// Clean up empty chain maps
		if len(chainMap) == 0 {
			delete(r.subscriptions, chainType)
		}
	}
}

// CancelAllSubscriptions cancels all subscriptions
func (r *SubscriptionManager) CancelAllSubscriptions() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, chainMap := range r.subscriptions {
		for _, sub := range chainMap {
			if sub.CancelFunc != nil {
				sub.CancelFunc()
			}
		}
	}

	// Clear all subscriptions
	r.subscriptions = make(map[types.ChainType]map[string]*ContractSubscription)
}

// GetSubscriptionsForChain returns all subscriptions for a chain
func (r *SubscriptionManager) GetSubscriptionsForChain(chainType types.ChainType) []*ContractSubscription {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []*ContractSubscription

	if chainMap, exists := r.subscriptions[chainType]; exists {
		for _, sub := range chainMap {
			result = append(result, sub)
		}
	}

	return result
}

// HasEventSubscription checks if a contract has a subscription for a specific event
func (r *SubscriptionManager) HasEventSubscription(chainType types.ChainType, contractAddr string, eventSig string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	normalizedAddr := strings.ToLower(contractAddr)

	if chainMap, exists := r.subscriptions[chainType]; exists {
		if sub, exists := chainMap[normalizedAddr]; exists {
			_, hasEvent := sub.Events[eventSig]
			return hasEvent
		}
	}

	return false
}
