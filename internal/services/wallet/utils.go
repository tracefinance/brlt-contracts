package wallet

import (
	"vault0/internal/types"
)

// WalletMap provides efficient lookup of wallets by their addresses
type WalletMap struct {
	wallets    map[string]*Wallet
	chainType  types.ChainType
}

// NewWalletMap creates a new WalletMap instance with normalized wallet addresses
func NewWalletMap(wallets []*Wallet, chainType types.ChainType) *WalletMap {
	addressToWallet := make(map[string]*Wallet)
	for _, w := range wallets {
		if addr, err := types.NewAddress(w.Address, chainType); err == nil {
			addressToWallet[addr.String()] = w
		}
	}
	return &WalletMap{
		wallets:    addressToWallet,
		chainType:  chainType,
	}
}

// Get returns the wallet associated with the given address, if it exists
func (m *WalletMap) Get(address string) *Wallet {
	if addr, err := types.NewAddress(address, m.chainType); err == nil {
		return m.wallets[addr.String()]
	}
	return nil
}
