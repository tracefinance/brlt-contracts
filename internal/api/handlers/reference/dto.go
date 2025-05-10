package reference

import (
	"vault0/internal/types"
)

// ChainResponse defines the structure for the chain reference API response.
// swagger:model ChainResponse
type ChainResponse struct {
	// The unique identifier for the blockchain network.
	// example: "1"
	ID string `json:"id"`
	// The type of the blockchain (e.g., ethereum, polygon).
	// example: ethereum
	Type types.ChainType `json:"type"`
	// The layer classification of the blockchain (e.g., layer1, layer2).
	// example: layer1
	Layer types.ChainLayer `json:"layer"`
	// The human-readable name of the blockchain network.
	// example: Ethereum
	Name string `json:"name"`
	// The native currency symbol of the blockchain.
	// example: ETH
	Symbol string `json:"symbol"`
	// The URL of the block explorer for the network.
	// example: https://etherscan.io
	ExplorerURL string `json:"explorer_url"`
}

// TokenResponse defines the structure for the token reference API response.
// swagger:model TokenResponse
type TokenResponse struct {
	// The contract address of the token. For native tokens, this is typically the zero address.
	// example: 0x0000000000000000000000000000000000000000
	Address string `json:"address"`
	// The blockchain type the token exists on.
	// example: ethereum
	ChainType types.ChainType `json:"chain_type"`
	// The token symbol (ticker).
	// example: ETH
	Symbol string `json:"symbol"`
	// Number of decimal places the token supports.
	// example: 18
	Decimals uint8 `json:"decimals"`
	// The type of token (native or erc20).
	// example: native
	Type types.TokenType `json:"type"`
}
