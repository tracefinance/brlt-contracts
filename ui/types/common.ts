/**
 * Defines common reusable types across the frontend.
 */

/**
 * Supported blockchain network types.
 * Based on internal/types/chain.go
 */
export type ChainType =
  | 'ethereum'
  | 'polygon'
  | 'base'
  // Add other supported chain types here

/**
 * Supported token standard types.
 * Based on common standards and potential backend usage.
 */
export type TokenType =
  | 'erc20'
  | 'erc721'
  | 'erc1155'
  | 'native' // For native chain currency (e.g., ETH, MATIC)
  // Add other supported token types here 