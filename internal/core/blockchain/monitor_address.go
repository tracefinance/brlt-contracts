package blockchain

import (
	"strings"
	"sync"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// AddressMonitor manages a set of addresses being monitored across different chains
type AddressMonitor struct {
	// Map of chain type -> address -> struct{} (set implementation)
	monitoredAddresses map[types.ChainType]map[string]struct{}
	mutex              sync.RWMutex
	log                logger.Logger
}

// NewAddressSetMonitor creates a new address set monitor
func NewAddressSetMonitor(log logger.Logger) *AddressMonitor {
	return &AddressMonitor{
		monitoredAddresses: make(map[types.ChainType]map[string]struct{}),
		log:                log,
	}
}

// Add adds an address to the monitoring set
func (m *AddressMonitor) Add(addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.monitoredAddresses[addr.ChainType]; !ok {
		m.monitoredAddresses[addr.ChainType] = make(map[string]struct{})
	}

	if _, exists := m.monitoredAddresses[addr.ChainType][normalizedAddr]; !exists {
		m.monitoredAddresses[addr.ChainType][normalizedAddr] = struct{}{}
		m.log.Info("Added address to monitoring list",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(addr.ChainType)))
	} else {
		m.log.Debug("Address already monitored",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(addr.ChainType)))
	}

	return nil
}

// Remove removes an address from the monitoring set
func (m *AddressMonitor) Remove(addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	// We don't strictly need validation here, but it's good practice
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if chainMap, ok := m.monitoredAddresses[addr.ChainType]; ok {
		if _, exists := chainMap[normalizedAddr]; exists {
			delete(chainMap, normalizedAddr)
			m.log.Info("Removed address from monitoring list",
				logger.String("address", addr.Address),
				logger.String("chain_type", string(addr.ChainType)))
			// Clean up the chain map if it becomes empty
			if len(chainMap) == 0 {
				delete(m.monitoredAddresses, addr.ChainType)
			}
		} else {
			m.log.Debug("Address not found in monitoring list for removal",
				logger.String("address", addr.Address),
				logger.String("chain_type", string(addr.ChainType)))
		}
	} else {
		m.log.Debug("Chain type not found in monitoring list for removal",
			logger.String("chain_type", string(addr.ChainType)))
	}

	return nil
}

// IsMonitored checks if any of the given addresses on a specific chain are in the
// monitored addresses list
func (m *AddressMonitor) IsMonitored(chainType types.ChainType, addresses []string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	chainMap, chainExists := m.monitoredAddresses[chainType]
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

// GetAllAddresses returns all monitored addresses for a specific chain
func (m *AddressMonitor) GetAllAddresses(chainType types.ChainType) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	chainMap, chainExists := m.monitoredAddresses[chainType]
	if !chainExists {
		return []string{}
	}

	addresses := make([]string, 0, len(chainMap))
	for addr := range chainMap {
		addresses = append(addresses, addr)
	}

	return addresses
}
